package internal

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"

	_ "modernc.org/sqlite"
)

const (
	localSQLiteConfigName = "memo.sqlite"
	defaultLocalSQLite    = "~/.go_tools_memo.sqlite3"
)

func sqliteCompanionPaths(base string) []string {
	return []string{base + "-wal", base + "-shm"}
}

var (
	localSQLitePath   string
	localSQLiteReady  bool
	localSQLiteDB     *sql.DB
	localSQLiteDBPath string
	localSQLiteMu     sync.Mutex
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
	db, err := ensureLocalSQLiteDBLocked()
	if err != nil {
		panic(err)
	}
	if localSQLiteReady {
		return
	}
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`CREATE TABLE IF NOT EXISTS records (
			id TEXT PRIMARY KEY,
			add_date INTEGER NOT NULL,
			modified_date INTEGER NOT NULL,
			my_problem INTEGER NOT NULL,
			finished INTEGER NOT NULL,
			hold INTEGER NOT NULL,
			title TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS record_tags (
			record_id TEXT NOT NULL,
			tag TEXT NOT NULL,
			position INTEGER NOT NULL,
			PRIMARY KEY(record_id, position),
			FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			count INTEGER NOT NULL,
			modified_date INTEGER NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_records_modified_date ON records(modified_date);`,
		`CREATE INDEX IF NOT EXISTS idx_records_add_date ON records(add_date);`,
		`CREATE INDEX IF NOT EXISTS idx_records_finished ON records(finished);`,
		`CREATE INDEX IF NOT EXISTS idx_record_tags_record_id ON record_tags(record_id);`,
		`CREATE INDEX IF NOT EXISTS idx_record_tags_tag ON record_tags(tag);`,
	}
	if err := withSQLiteConn(db, func(ctx context.Context, conn *sql.Conn) error {
		for _, stmt := range statements {
			if _, err := conn.ExecContext(ctx, stmt); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	localSQLiteReady = true
}

func ensureLocalSQLiteDBLocked() (*sql.DB, error) {
	if localSQLiteDB != nil && localSQLiteDBPath == localSQLitePath {
		return localSQLiteDB, nil
	}
	if localSQLiteDB != nil {
		if err := localSQLiteDB.Close(); err != nil {
			return nil, err
		}
		localSQLiteDB = nil
		localSQLiteDBPath = ""
		localSQLiteReady = false
	}
	db, err := sql.Open("sqlite", localSQLitePath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := withSQLiteConn(db, func(context.Context, *sql.Conn) error {
		return nil
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	localSQLiteDB = db
	localSQLiteDBPath = localSQLitePath
	return db, nil
}

func withSQLitePath(path string, fn func() error) error {
	path = utilsw.ExpandUser(path)
	localSQLiteMu.Lock()
	oldPath := localSQLitePath
	oldReady := localSQLiteReady
	oldDB := localSQLiteDB
	oldDBPath := localSQLiteDBPath
	oldBackendMode := localBackendMode.Get().(string)
	localSQLitePath = path
	localSQLiteReady = false
	localSQLiteDB = nil
	localSQLiteDBPath = ""
	localSQLiteMu.Unlock()

	SetLocalBackendMode(LocalBackendSQLite)
	defer SetLocalBackendMode(oldBackendMode)
	defer func() {
		localSQLiteMu.Lock()
		tempDB := localSQLiteDB
		localSQLitePath = oldPath
		localSQLiteReady = oldReady
		localSQLiteDB = oldDB
		localSQLiteDBPath = oldDBPath
		localSQLiteMu.Unlock()
		if tempDB != nil && tempDB != oldDB {
			_ = tempDB.Close()
		}
	}()

	return fn()
}

func sqliteSidecarPaths(path string) []string {
	return sqliteCompanionPaths(utilsw.ExpandUser(path))
}

func prepareSQLitePathForTransfer(path string) error {
	return withSQLitePath(path, func() error {
		return withSQLiteConn(sqliteDB(), func(ctx context.Context, conn *sql.Conn) error {
			_, err := conn.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
			return err
		})
	})
}

func withSQLiteConn(db *sql.DB, fn func(context.Context, *sql.Conn) error) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return err
	}
	return fn(ctx, conn)
}

func withSQLiteTx(db *sql.DB, fn func(context.Context, *sql.Tx) error) error {
	return withSQLiteConn(db, func(ctx context.Context, conn *sql.Conn) error {
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := fn(ctx, tx); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	})
}

func sqliteDB() *sql.DB {
	initLocalSQLite()
	localSQLiteMu.Lock()
	defer localSQLiteMu.Unlock()
	return localSQLiteDB
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

func sqlitePlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	parts := make([]string, n)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
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

func sqliteStringArgs(values []string) []any {
	args := make([]any, len(values))
	for i, value := range values {
		args[i] = value
	}
	return args
}

func sqliteScanRecordRows(rows *sql.Rows) ([]sqliteRecordRow, error) {
	defer rows.Close()
	res := make([]sqliteRecordRow, 0)
	for rows.Next() {
		var row sqliteRecordRow
		if err := rows.Scan(&row.ID, &row.AddDate, &row.ModifiedDate, &row.MyProblem, &row.Finished, &row.Hold, &row.Title); err != nil {
			return nil, err
		}
		res = append(res, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func sqliteRecordTagRowsByIDs(ctx context.Context, conn *sql.Conn, ids []string) ([]sqliteRecordTagRow, error) {
	if len(ids) == 0 {
		return []sqliteRecordTagRow{}, nil
	}
	query := fmt.Sprintf(`
SELECT record_id, tag, position
FROM record_tags
WHERE record_id IN (%s)
ORDER BY record_id, position;
`, sqlitePlaceholders(len(ids)))
	rows, err := conn.QueryContext(ctx, query, sqliteStringArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make([]sqliteRecordTagRow, 0)
	for rows.Next() {
		var row sqliteRecordTagRow
		if err := rows.Scan(&row.RecordID, &row.Tag, &row.Position); err != nil {
			return nil, err
		}
		res = append(res, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func sqliteApplyTagCountDeltaTx(ctx context.Context, tx *sql.Tx, tags []string, val int) error {
	if len(tags) == 0 || val == 0 {
		return nil
	}
	now := time.Now().UnixNano()
	for _, tag := range tags {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO tags(id, name, count, modified_date)
VALUES (?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	count = count + excluded.count,
	modified_date = excluded.modified_date;
`, primitive.NewObjectID().Hex(), tag, val, now); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM tags WHERE count < 1;")
	return err
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
	var row sqliteCountRow
	err := withSQLiteConn(sqliteDB(), func(ctx context.Context, conn *sql.Conn) error {
		return conn.QueryRowContext(ctx, "SELECT COUNT(*) AS count FROM records WHERE id = ?;", id.Hex()).Scan(&row.Count)
	})
	if err != nil {
		return false, err
	}
	return row.Count > 0, nil
}

func sqliteLoadRecord(id primitive.ObjectID) (*Record, error) {
	var rows []sqliteRecordRow
	var tagRows []sqliteRecordTagRow
	err := withSQLiteConn(sqliteDB(), func(ctx context.Context, conn *sql.Conn) error {
		queryRows, err := conn.QueryContext(ctx, `
SELECT id, add_date, modified_date, my_problem, finished, hold, title
FROM records
WHERE id = ?;
`, id.Hex())
		if err != nil {
			return err
		}
		rows, err = sqliteScanRecordRows(queryRows)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		tagRows, err = sqliteRecordTagRowsByIDs(ctx, conn, []string{id.Hex()})
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
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
	args := make([]any, 0, len(tags)+2)
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
	HAVING COUNT(DISTINCT tag) = ?
)`, sqlitePlaceholders(len(uniqueTags))))
			args = append(args, sqliteStringArgs(uniqueTags)...)
			args = append(args, len(uniqueTags))
		} else if prefix {
			parts := make([]string, 0, len(tags))
			for _, tag := range tags {
				parts = append(parts, "instr(rt.tag, ?) > 0")
				args = append(args, tag)
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
)`, sqlitePlaceholders(len(tags))))
			args = append(args, sqliteStringArgs(tags)...)
		}
	}
	if title != "" {
		where = append(where, "instr(lower(r.title), lower(?)) > 0")
		args = append(args, title)
	}
	query := fmt.Sprintf(`
SELECT r.id, r.add_date, r.modified_date, r.my_problem, r.finished, r.hold, r.title
FROM records r
%s
ORDER BY r.modified_date %s, r.add_date %s
LIMIT ?;
`, sqliteWhereClause(where), direction, direction)
	args = append(args, limit)
	var rows []sqliteRecordRow
	var tagRows []sqliteRecordTagRow
	err := withSQLiteConn(sqliteDB(), func(ctx context.Context, conn *sql.Conn) error {
		queryRows, err := conn.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		rows, err = sqliteScanRecordRows(queryRows)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			tagRows = []sqliteRecordTagRow{}
			return nil
		}
		ids := make([]string, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}
		tagRows, err = sqliteRecordTagRowsByIDs(ctx, conn, ids)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []*Record{}, nil
	}
	return sqliteParseRecordRows(rows, tagRows)
}

func sqliteSaveRecord(r *Record, noUpdateModifiedDate bool) error {
	modifiedDate := r.ModifiedDate
	if !noUpdateModifiedDate {
		modifiedDate = time.Now()
	}
	return withSQLiteTx(sqliteDB(), func(ctx context.Context, tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO records(id, add_date, modified_date, my_problem, finished, hold, title)
VALUES (?, ?, ?, ?, ?, ?, ?);
`, r.ID.Hex(), r.AddDate.UnixNano(), modifiedDate.UnixNano(), sqliteBoolNum(r.MyProblem), sqliteBoolNum(r.Finished), sqliteBoolNum(r.Hold), r.Title)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return nil
		}
		for i, tag := range r.Tags {
			if _, err := tx.ExecContext(ctx, "INSERT INTO record_tags(record_id, tag, position) VALUES (?, ?, ?);", r.ID.Hex(), tag, i); err != nil {
				return err
			}
		}
		if err := sqliteApplyTagCountDeltaTx(ctx, tx, r.Tags, 1); err != nil {
			return err
		}
		r.ModifiedDate = modifiedDate
		return nil
	})
}

func sqliteUpsertRecord(r *Record) error {
	exists, err := sqliteRecordExists(r.ID)
	if err != nil {
		return err
	}
	if !exists {
		return sqliteSaveRecord(r, true)
	}
	return sqliteUpdateRecord(r)
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
	return withSQLiteTx(sqliteDB(), func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "DELETE FROM records WHERE id = ?;", r.ID.Hex()); err != nil {
			return err
		}
		return sqliteApplyTagCountDeltaTx(ctx, tx, r.Tags, -1)
	})
}

func sqliteUpdateRecord(r *Record) error {
	existing, err := sqliteLoadRecord(r.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return sqliteSaveRecord(r, true)
	}
	oldTags := append([]string(nil), existing.Tags...)
	return withSQLiteTx(sqliteDB(), func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
UPDATE records
SET add_date = ?, modified_date = ?, my_problem = ?, finished = ?, hold = ?, title = ?
WHERE id = ?;
`, r.AddDate.UnixNano(), r.ModifiedDate.UnixNano(), sqliteBoolNum(r.MyProblem), sqliteBoolNum(r.Finished), sqliteBoolNum(r.Hold), r.Title, r.ID.Hex()); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM record_tags WHERE record_id = ?;", r.ID.Hex()); err != nil {
			return err
		}
		for i, tag := range r.Tags {
			if _, err := tx.ExecContext(ctx, "INSERT INTO record_tags(record_id, tag, position) VALUES (?, ?, ?);", r.ID.Hex(), tag, i); err != nil {
				return err
			}
		}
		if err := sqliteApplyTagCountDeltaTx(ctx, tx, oldTags, -1); err != nil {
			return err
		}
		if err := sqliteApplyTagCountDeltaTx(ctx, tx, r.Tags, 1); err != nil {
			return err
		}
		return nil
	})
}

func sqliteIncrementTagCount(tags []string, val int) error {
	if len(tags) == 0 || val == 0 {
		return nil
	}
	return withSQLiteTx(sqliteDB(), func(ctx context.Context, tx *sql.Tx) error {
		return sqliteApplyTagCountDeltaTx(ctx, tx, tags, val)
	})
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
		args := make([]any, 0, SpecialTagPatterns.Size())
		for val := range SpecialTagPatterns.Iter().Iterate() {
			parts = append(parts, "instr(name, ?) != 1")
			args = append(args, val.(string))
		}
		if len(parts) > 0 {
			where = append(where, strings.Join(parts, " AND "))
		}
		query := fmt.Sprintf(`
SELECT id, name, count, modified_date
FROM tags
%s
ORDER BY name %s
LIMIT ?;
`, sqliteWhereClause(where), direction)
		args = append(args, limit)
		rows, err := sqliteListTagRows(query, args)
		if err != nil {
			return nil, err
		}
		return sqliteTagRowsToTags(rows), nil
	}
	query := fmt.Sprintf(`
SELECT id, name, count, modified_date
FROM tags
%s
ORDER BY name %s
LIMIT ?;
`, sqliteWhereClause(where), direction)
	rows, err := sqliteListTagRows(query, []any{limit})
	if err != nil {
		return nil, err
	}
	return sqliteTagRowsToTags(rows), nil
}

func sqliteListTagRows(query string, args []any) ([]sqliteTagRow, error) {
	var rows []sqliteTagRow
	err := withSQLiteConn(sqliteDB(), func(ctx context.Context, conn *sql.Conn) error {
		queryRows, err := conn.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer queryRows.Close()
		for queryRows.Next() {
			var row sqliteTagRow
			if err := queryRows.Scan(&row.ID, &row.Name, &row.Count, &row.ModifiedDate); err != nil {
				return err
			}
			rows = append(rows, row)
		}
		return queryRows.Err()
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func sqliteTagRowsToTags(rows []sqliteTagRow) []Tag {
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
	return res
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
