package internal

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	localSQLiteConfigName = "memo.sqlite"
	defaultLocalSQLite    = "~/.go_tools_memo.sqlite3"
)

var (
	localSQLitePath  string
	localSQLiteReady bool
	localSQLiteMu    sync.Mutex
)

type sqliteCountRow struct {
	Count int64 `json:"count"`
}

type sqliteRecordRow struct {
	ID           string `json:"id"`
	AddDate      int64  `json:"add_date"`
	ModifiedDate int64  `json:"modified_date"`
	MyProblem    int64  `json:"my_problem"`
	Finished     int64  `json:"finished"`
	Hold         int64  `json:"hold"`
	Title        string `json:"title"`
}

type sqliteRecordTagRow struct {
	RecordID string `json:"record_id"`
	Tag      string `json:"tag"`
	Position int    `json:"position"`
}

type sqliteTagRow struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Count        int64  `json:"count"`
	ModifiedDate int64  `json:"modified_date"`
}

func initLocalSQLite() {
	localSQLiteMu.Lock()
	defer localSQLiteMu.Unlock()
	if localSQLiteReady {
		return
	}
	if localSQLitePath == "" {
		m := utilsw.GetAllConfig()
		configuredPath := strings.TrimSpace(m.GetOrDefault(localSQLiteConfigName, "").(string))
		if configuredPath == "" {
			configuredPath = defaultLocalSQLite
		}
		localSQLitePath = utilsw.ExpandUser(configuredPath)
	}
	if err := os.MkdirAll(filepath.Dir(localSQLitePath), 0o755); err != nil {
		panic(err)
	}
	if _, err := execSQLiteScript(localSQLitePath, `
PRAGMA journal_mode = WAL;
CREATE TABLE IF NOT EXISTS records (
	id TEXT PRIMARY KEY,
	add_date INTEGER NOT NULL,
	modified_date INTEGER NOT NULL,
	my_problem INTEGER NOT NULL,
	finished INTEGER NOT NULL,
	hold INTEGER NOT NULL,
	title TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS record_tags (
	record_id TEXT NOT NULL,
	tag TEXT NOT NULL,
	position INTEGER NOT NULL,
	PRIMARY KEY(record_id, position),
	FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS tags (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	count INTEGER NOT NULL,
	modified_date INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_records_modified_date ON records(modified_date);
CREATE INDEX IF NOT EXISTS idx_records_add_date ON records(add_date);
CREATE INDEX IF NOT EXISTS idx_records_finished ON records(finished);
CREATE INDEX IF NOT EXISTS idx_record_tags_record_id ON record_tags(record_id);
CREATE INDEX IF NOT EXISTS idx_record_tags_tag ON record_tags(tag);
`, false); err != nil {
		panic(err)
	}
	localSQLiteReady = true
}

func execSQLiteScript(path, script string, jsonOutput bool) ([]byte, error) {
	args := []string{"-batch", "-bail"}
	if jsonOutput {
		args = append(args, "-json")
	}
	args = append(args, path)
	cmd := exec.Command("sqlite3", args...)
	cmd.Stdin = strings.NewReader("PRAGMA foreign_keys = ON;\n" + script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("sqlite3 failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil

}

func runSQLiteScript(script string, jsonOutput bool) ([]byte, error) {
	initLocalSQLite()
	return execSQLiteScript(localSQLitePath, script, jsonOutput)
}

func sqliteTextExpr(s string) string {
	return fmt.Sprintf("CAST(X'%s' AS TEXT)", strings.ToUpper(hex.EncodeToString([]byte(s))))
}

func sqliteBoolNum(val bool) int {
	if val {
		return 1
	}
	return 0
}

func sqliteWhereClause(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(parts, " AND ")
}

func sqliteTextListExpr(values []string) string {
	if len(values) == 0 {
		return ""
	}
	exprs := make([]string, len(values))
	for i, value := range values {
		exprs[i] = sqliteTextExpr(value)
	}
	return strings.Join(exprs, ", ")
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	res := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		res = append(res, value)
	}
	return res
}

func sqliteTagCountStatements(tags []string, val int) string {
	if len(tags) == 0 || val == 0 {
		return ""
	}
	now := time.Now().UnixNano()
	buf := strings.Builder{}
	for _, tag := range tags {
		fmt.Fprintf(&buf, "INSERT INTO tags(id, name, count, modified_date) VALUES (%s, %s, %d, %d) "+
			"ON CONFLICT(name) DO UPDATE SET count = count + (%d), modified_date = %d;\n",
			sqliteTextExpr(primitive.NewObjectID().Hex()), sqliteTextExpr(tag), val, now, val, now)
	}
	buf.WriteString("DELETE FROM tags WHERE count < 1;\n")
	return buf.String()
}

func sqliteParseRecordRows(rows []sqliteRecordRow, tagRows []sqliteRecordTagRow) ([]*Record, error) {
	tagsByRecord := make(map[string][]string, len(rows))
	for _, row := range tagRows {
		tagsByRecord[row.RecordID] = append(tagsByRecord[row.RecordID], row.Tag)
	}
	res := make([]*Record, 0, len(rows))
	for _, row := range rows {
		id, err := primitive.ObjectIDFromHex(row.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, &Record{
			ID:           id,
			Tags:         tagsByRecord[row.ID],
			AddDate:      time.Unix(0, row.AddDate),
			ModifiedDate: time.Unix(0, row.ModifiedDate),
			MyProblem:    row.MyProblem != 0,
			Finished:     row.Finished != 0,
			Hold:         row.Hold != 0,
			Title:        row.Title,
		})
	}
	return res, nil
}

func sqliteRecordExists(id primitive.ObjectID) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) AS count FROM records WHERE id = %s;", sqliteTextExpr(id.Hex()))
	out, err := runSQLiteScript(query, true)
	if err != nil {
		return false, err
	}
	var rows []sqliteCountRow
	if err = json.Unmarshal(out, &rows); err != nil {
		return false, err
	}
	return len(rows) == 1 && rows[0].Count > 0, nil
}

func sqliteLoadRecord(id primitive.ObjectID) (*Record, error) {
	query := fmt.Sprintf(`
SELECT id, add_date, modified_date, my_problem, finished, hold, title
FROM records
WHERE id = %s;
`, sqliteTextExpr(id.Hex()))
	out, err := runSQLiteScript(query, true)
	if err != nil {
		return nil, err
	}
	var rows []sqliteRecordRow
	if err = json.Unmarshal(out, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	tagQuery := fmt.Sprintf(`
SELECT record_id, tag, position
FROM record_tags
WHERE record_id = %s
ORDER BY position;
`, sqliteTextExpr(id.Hex()))
	tagOut, err := runSQLiteScript(tagQuery, true)
	if err != nil {
		return nil, err
	}
	var tagRows []sqliteRecordTagRow
	if err = json.Unmarshal(tagOut, &tagRows); err != nil {
		return nil, err
	}
	records, err := sqliteParseRecordRows(rows, tagRows)
	if err != nil {
		return nil, err
	}
	return records[0], nil
}

func sqliteListRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool, title string, prefix bool) ([]*Record, error) {
	if tags == nil {
		tags = []string{}
	}
	if limit <= 0 {
		limit = math.MaxInt64
	}
	direction := "ASC"
	if reverse {
		direction = "DESC"
	}
	where := make([]string, 0, 4)
	if !includeFinished {
		where = append(where, "r.finished = 0")
	}
	if len(tags) > 0 {
		if useAnd {
			uniqueTags := uniqueStrings(tags)
			where = append(where, fmt.Sprintf(`r.id IN (
	SELECT record_id
	FROM record_tags
	WHERE tag IN (%s)
	GROUP BY record_id
	HAVING COUNT(DISTINCT tag) = %d
)`, sqliteTextListExpr(uniqueTags), len(uniqueTags)))
		} else if prefix {
			parts := make([]string, 0, len(tags))
			for _, tag := range tags {
				parts = append(parts, fmt.Sprintf("instr(rt.tag, %s) > 0", sqliteTextExpr(tag)))
			}
			where = append(where, fmt.Sprintf(`EXISTS (
	SELECT 1
	FROM record_tags rt
	WHERE rt.record_id = r.id AND (%s)
)`, strings.Join(parts, " OR ")))
		} else {
			where = append(where, fmt.Sprintf(`EXISTS (
	SELECT 1
	FROM record_tags rt
	WHERE rt.record_id = r.id AND rt.tag IN (%s)
)`, sqliteTextListExpr(tags)))
		}
	}
	if title != "" {
		where = append(where, fmt.Sprintf("instr(lower(r.title), lower(%s)) > 0", sqliteTextExpr(title)))
	}
	query := fmt.Sprintf(`
SELECT r.id, r.add_date, r.modified_date, r.my_problem, r.finished, r.hold, r.title
FROM records r
%s
ORDER BY r.modified_date %s, r.add_date %s
LIMIT %d;
`, sqliteWhereClause(where), direction, direction, limit)
	out, err := runSQLiteScript(query, true)
	if err != nil {
		return nil, err
	}
	var rows []sqliteRecordRow
	if err = json.Unmarshal(out, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*Record{}, nil
	}
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	tagQuery := fmt.Sprintf(`
SELECT record_id, tag, position
FROM record_tags
WHERE record_id IN (%s)
ORDER BY record_id, position;
`, sqliteTextListExpr(ids))
	tagOut, err := runSQLiteScript(tagQuery, true)
	if err != nil {
		return nil, err
	}
	var tagRows []sqliteRecordTagRow
	if err = json.Unmarshal(tagOut, &tagRows); err != nil {
		return nil, err
	}
	return sqliteParseRecordRows(rows, tagRows)
}

func sqliteSaveRecord(r *Record, noUpdateModifiedDate bool) error {
	exists, err := sqliteRecordExists(r.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !noUpdateModifiedDate {
		r.ModifiedDate = time.Now()
	}
	buf := strings.Builder{}
	buf.WriteString("BEGIN IMMEDIATE;\n")
	fmt.Fprintf(&buf, "INSERT INTO records(id, add_date, modified_date, my_problem, finished, hold, title) VALUES (%s, %d, %d, %d, %d, %d, %s);\n",
		sqliteTextExpr(r.ID.Hex()), r.AddDate.UnixNano(), r.ModifiedDate.UnixNano(), sqliteBoolNum(r.MyProblem),
		sqliteBoolNum(r.Finished), sqliteBoolNum(r.Hold), sqliteTextExpr(r.Title))
	for i, tag := range r.Tags {
		fmt.Fprintf(&buf, "INSERT INTO record_tags(record_id, tag, position) VALUES (%s, %s, %d);\n",
			sqliteTextExpr(r.ID.Hex()), sqliteTextExpr(tag), i)
	}
	buf.WriteString(sqliteTagCountStatements(r.Tags, 1))
	buf.WriteString("COMMIT;\n")
	_, err = runSQLiteScript(buf.String(), false)
	return err
}

func sqliteDeleteRecord(r *Record) error {
	if len(r.Tags) == 0 {
		loaded, err := sqliteLoadRecord(r.ID)
		if err != nil {
			return err
		}
		if loaded != nil {
			r.Tags = loaded.Tags
		}
	}
	buf := strings.Builder{}
	buf.WriteString("BEGIN IMMEDIATE;\n")
	fmt.Fprintf(&buf, "DELETE FROM records WHERE id = %s;\n", sqliteTextExpr(r.ID.Hex()))
	buf.WriteString(sqliteTagCountStatements(r.Tags, -1))
	buf.WriteString("COMMIT;\n")
	_, err := runSQLiteScript(buf.String(), false)
	return err
}

func sqliteUpdateRecord(r *Record) error {
	buf := strings.Builder{}
	buf.WriteString("BEGIN IMMEDIATE;\n")
	fmt.Fprintf(&buf, "UPDATE records SET add_date = %d, modified_date = %d, my_problem = %d, finished = %d, hold = %d, title = %s WHERE id = %s;\n",
		r.AddDate.UnixNano(), r.ModifiedDate.UnixNano(), sqliteBoolNum(r.MyProblem), sqliteBoolNum(r.Finished),
		sqliteBoolNum(r.Hold), sqliteTextExpr(r.Title), sqliteTextExpr(r.ID.Hex()))
	fmt.Fprintf(&buf, "DELETE FROM record_tags WHERE record_id = %s;\n", sqliteTextExpr(r.ID.Hex()))
	for i, tag := range r.Tags {
		fmt.Fprintf(&buf, "INSERT INTO record_tags(record_id, tag, position) VALUES (%s, %s, %d);\n",
			sqliteTextExpr(r.ID.Hex()), sqliteTextExpr(tag), i)
	}
	buf.WriteString("COMMIT;\n")
	_, err := runSQLiteScript(buf.String(), false)
	return err
}

func sqliteIncrementTagCount(tags []string, val int) error {
	if len(tags) == 0 || val == 0 {
		return nil
	}
	buf := strings.Builder{}
	buf.WriteString("BEGIN IMMEDIATE;\n")
	buf.WriteString(sqliteTagCountStatements(tags, val))
	buf.WriteString("COMMIT;\n")
	_, err := runSQLiteScript(buf.String(), false)
	return err
}

func sqliteListTags(limit int64, reverse bool) ([]Tag, error) {
	if limit <= 0 {
		limit = math.MaxInt64
	}
	direction := "ASC"
	if reverse {
		direction = "DESC"
	}
	where := make([]string, 0, 1)
	if !ListSpecial {
		parts := make([]string, 0, SpecialTagPatterns.Size())
		for val := range SpecialTagPatterns.Iter().Iterate() {
			parts = append(parts, fmt.Sprintf("instr(name, %s) != 1", sqliteTextExpr(val.(string))))
		}
		if len(parts) > 0 {
			where = append(where, strings.Join(parts, " AND "))
		}
	}
	query := fmt.Sprintf(`
SELECT id, name, count, modified_date
FROM tags
%s
ORDER BY name %s
LIMIT %d;
`, sqliteWhereClause(where), direction, limit)
	out, err := runSQLiteScript(query, true)
	if err != nil {
		return nil, err
	}
	var rows []sqliteTagRow
	if err = json.Unmarshal(out, &rows); err != nil {
		return nil, err
	}
	res := make([]Tag, 0, len(rows))
	for _, row := range rows {
		id, parseErr := primitive.ObjectIDFromHex(row.ID)
		if parseErr != nil {
			id = primitive.NilObjectID
		}
		res = append(res, Tag{
			ID:           id,
			Name:         row.Name,
			Count:        row.Count,
			ModifiedDate: time.Unix(0, row.ModifiedDate),
		})
	}
	return res, nil
}

func ListTags(limit int64, reverse bool) ([]Tag, error) {
	if Remote.Get().(bool) {
		InitRemote()
		if limit <= 0 {
			limit = math.MaxInt64
		}
		op := options.Find().SetLimit(limit)
		if reverse {
			op.SetSort(bson.M{"name": -1})
		} else {
			op.SetSort(bson.M{"name": 1})
		}
		filter := bson.M{}
		if !ListSpecial {
			filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: BuildMongoRegularExpExclude(SpecialTagPatterns)}}
		}
		cursor, err := AtlasClient.Database(DbName).Collection(TagCollectionName).Find(ctx, filter, op)
		if err != nil {
			return nil, err
		}
		var tags []Tag
		if err = cursor.All(ctx, &tags); err != nil {
			return nil, err
		}
		return tags, nil
	}
	if useLocalSQLite() {
		return sqliteListTags(limit, reverse)
	}
	if limit <= 0 {
		limit = math.MaxInt64
	}
	op := options.Find().SetLimit(limit)
	if reverse {
		op.SetSort(bson.M{"name": -1})
	} else {
		op.SetSort(bson.M{"name": 1})
	}
	filter := bson.M{}
	if !ListSpecial {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: BuildMongoRegularExpExclude(SpecialTagPatterns)}}
	}
	cursor, err := Client.Database(DbName).Collection(TagCollectionName).Find(ctx, filter, op)
	if err != nil {
		return nil, err
	}
	var tags []Tag
	if err = cursor.All(ctx, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}
