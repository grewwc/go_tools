package internal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

const (
	DbName            = "daily"
	CollectionName    = "memo"
	TagCollectionName = "tag"
)

// config key settings
// .configW
const (
	localMongoConfigName = "mongo.local"
	atlasMongoConfigName = "mongo.atlas"
	specialTagConfigname = "special.tags"

	DefaultBackendConfigName = "re.backend"
)

const (
	LocalBackendAuto   = "auto"
	LocalBackendMongo  = "mongo"
	LocalBackendSQLite = "sqlite"

	defaultLocalMongoURI  = "mongodb://localhost:27017"
	localMongoInitTimeout = 1500 * time.Millisecond
)

const (
	autoTag              = "auto"
	DefaultTxtOutputName = "output.txt"
	outputName           = "output_binary"
	Finish               = "finish"

	titleLen = 200
)

// biz vars
var (
	ListTagsAndOrderByTime       = false
	Reverse                      = false
	IncludeFinished              = false
	TxtOutputName                = "output.txt"
	ToBinary                     = false
	SpecialTagPatterns           = cw.NewSet("learn")
	Verbose                      = false
	RecordLimit            int64 = 100
	Prefix                       = false
)

var (
	DefaultBackendMode = utilsw.GetConfig(DefaultBackendConfigName, LocalBackendAuto)
)

var (
	ctx         context.Context
	Client      *mongo.Client
	AtlasClient *mongo.Client
)

var (
	Remote           = utilsw.NewThreadSafeVal(false)
	localBackendMode = utilsw.NewThreadSafeVal(LocalBackendAuto)
	mu               sync.Mutex
	localMongoMu     sync.Mutex
)

var (
	localMongoChecked bool
	localMongoErr     error
)

var (
	ListSpecial = false
	UseVsCode   = false
	OnlyTags    = false
)

func InitRemote() {
	mu.Lock()
	defer mu.Unlock()
	if AtlasClient != nil {
		Remote.Set(true)
		return
	}
	fmt.Println("connecting to Remote...")
	m := utilsw.GetAllConfig()
	var err error
	// mongo atlas init
	atlasURI := strings.TrimSpace(m.GetOrDefault(atlasMongoConfigName, "").(string))
	if atlasURI == "" {
		panic("mongo.atlas not configured")
	}
	clientOptions := options.Client().ApplyURI(atlasURI)
	AtlasClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}
	// check if tags and memo collections exists
	db := AtlasClient.Database(DbName)
	if !CollectionExists(db, ctx, TagCollectionName) {
		db.Collection(TagCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.DocElem{Name: "name", Value: "text"}}.Map(),
			Options: options.Index().SetUnique(true),
		})
	}
	fmt.Println("connected")
	Remote.Set(true)
	// fmt.Println("init Atlas", atlasURI, atlasClient)
}

func normalizeLocalBackendMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", LocalBackendAuto:
		return LocalBackendAuto
	case LocalBackendMongo, "mongodb":
		return LocalBackendMongo
	case LocalBackendSQLite, "sqlite3":
		return LocalBackendSQLite
	default:
		panic(fmt.Sprintf("unknown backend %q, use auto|mongo|sqlite", mode))
	}
}

func SetLocalBackendMode(mode string) {
	localBackendMode.Set(normalizeLocalBackendMode(mode))
}

func initLocalMongo() error {
	localMongoMu.Lock()
	defer localMongoMu.Unlock()
	if Client != nil {
		return nil
	}
	if localMongoChecked {
		return localMongoErr
	}
	m := utilsw.GetAllConfig()
	uri := strings.TrimSpace(m.GetOrDefault(localMongoConfigName, "").(string))
	if uri == "" {
		uri = defaultLocalMongoURI
	}
	connectCtx, cancel := context.WithTimeout(context.Background(), localMongoInitTimeout)
	defer cancel()
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(10)
	clientOptions.SetConnectTimeout(localMongoInitTimeout)
	clientOptions.SetServerSelectionTimeout(localMongoInitTimeout)
	client, err := mongo.Connect(connectCtx, clientOptions)
	if err == nil {
		err = client.Ping(connectCtx, readpref.Primary())
	}
	localMongoChecked = true
	localMongoErr = err
	if err != nil {
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		return err
	}
	Client = client
	return nil
}

func useLocalSQLite() bool {
	if Remote.Get().(bool) {
		return false
	}
	mode := localBackendMode.Get().(string)
	switch mode {
	case LocalBackendSQLite:
		initLocalSQLite()
		return true
	case LocalBackendMongo:
		if err := initLocalMongo(); err != nil {
			panic(fmt.Sprintf("local mongodb unavailable: %v", err))
		}
		return false
	default:
		if err := initLocalMongo(); err == nil {
			return false
		}
		initLocalSQLite()
		return true
	}
}

func init() {
	ctx = context.Background()

	// read the special tag patters from .configW
	m := utilsw.GetAllConfig()
	for _, val := range strw.SplitNoEmpty(m.GetOrDefault(specialTagConfigname, "").(string), ",") {
		val = strings.TrimSpace(val)
		SpecialTagPatterns.Add(val)
	}

}

func ListRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool, title string, prefix bool) ([]*Record, bool) {
	return listRecords(limit, reverse, includeFinished, tags, useAnd, title, prefix, true)
}

func listRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool, title string, prefix bool, writeInfo bool) ([]*Record, bool) {
	if tags == nil {
		tags = []string{}
	}
	if limit <= 0 {
		limit = math.MaxInt64
	}
	var res []*Record
	if Remote.Get().(bool) {
		InitRemote()
		reverseNum := 1
		if reverse {
			reverseNum = -1
		}
		collection := AtlasClient.Database(DbName).Collection(CollectionName)
		modifiedDataOption := options.Find()
		addDateOption := options.Find()
		modifiedDataOption.SetLimit(limit)
		addDateOption.SetLimit(limit)
		modifiedDataOption.SetSort(bson.M{"modified_date": reverseNum})
		addDateOption.SetSort(bson.M{"add_date": reverseNum})
		m := bson.M{}
		if !includeFinished {
			m["finished"] = false
		}

		if len(tags) > 0 {
			if useAnd {
				m["tags"] = bson.M{"$all": tags}
			} else {
				if prefix {
					tagsReg := make([]primitive.Regex, len(tags))
					for i := range tags {
						tagsReg[i] = primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", tags[i])}
					}
					m["tags"] = bson.M{"$elemMatch": bson.M{"$in": tagsReg}}
				} else {
					m["tags"] = bson.M{"$elemMatch": bson.M{"$in": tags}}
				}
			}
		}
		if title != "" {
			m["title"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", title), Options: "i"}}
		}

		cursor, err := collection.Find(ctx, m, addDateOption, modifiedDataOption)
		if err != nil {
			panic(err)
		}
		if err = cursor.All(ctx, &res); err != nil {
			panic(err)
		}
	} else if useLocalSQLite() {
		var err error
		res, err = sqliteListRecords(limit, reverse, includeFinished, tags, useAnd, title, prefix)
		if err != nil {
			panic(err)
		}
	} else {
		reverseNum := 1
		if reverse {
			reverseNum = -1
		}
		collection := Client.Database(DbName).Collection(CollectionName)
		modifiedDataOption := options.Find()
		addDateOption := options.Find()
		modifiedDataOption.SetLimit(limit)
		addDateOption.SetLimit(limit)
		modifiedDataOption.SetSort(bson.M{"modified_date": reverseNum})
		addDateOption.SetSort(bson.M{"add_date": reverseNum})
		m := bson.M{}
		if !includeFinished {
			m["finished"] = false
		}

		if len(tags) > 0 {
			if useAnd {
				m["tags"] = bson.M{"$all": tags}
			} else {
				if prefix {
					tagsReg := make([]primitive.Regex, len(tags))
					for i := range tags {
						tagsReg[i] = primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", tags[i])}
					}
					m["tags"] = bson.M{"$elemMatch": bson.M{"$in": tagsReg}}
				} else {
					m["tags"] = bson.M{"$elemMatch": bson.M{"$in": tags}}
				}
			}
		}
		if title != "" {
			m["title"] = bson.M{"$regex": primitive.Regex{Pattern: fmt.Sprintf(".*%s.*", title), Options: "i"}}
		}

		cursor, err := collection.Find(ctx, m, addDateOption, modifiedDataOption)
		if err != nil {
			panic(err)
		}
		if err = cursor.All(ctx, &res); err != nil {
			panic(err)
		}
	}
	// filter by special tags
	// fmt.Println("here", res)
	if !ListSpecial {
		resCopy := make([]*Record, 0, len(res))
		for _, r := range res {
			trie := cw.NewTrie()
			tags := r.Tags
			for _, t := range tags {
				trie.Insert(t)
			}
			if !SearchTrie(trie, SpecialTagPatterns) {
				resCopy = append(resCopy, r)
			}
		}
		res = resCopy
	}
	recordTitles := make([]string, len(res))
	recordIDs := make([]*primitive.ObjectID, len(res))
	for i := range res {
		recordTitles[i] = res[i].Title
		recordIDs[i] = &res[i].ID
	}
	written := false
	if writeInfo {
		written = WriteInfo(recordIDs, recordTitles)
	}
	return res, written
}

func (r *Record) exists() bool {
	if Remote.Get().(bool) {
		collection := AtlasClient.Database(DbName).Collection(CollectionName)
		singleResults := collection.FindOne(context.Background(), bson.M{"_id": r.ID})
		err := singleResults.Err()
		if err == nil {
			return true
		}

		if err == mongo.ErrNoDocuments {
			return false
		}
		panic(err)
	}
	if useLocalSQLite() {
		exists, err := sqliteRecordExists(r.ID)
		if err != nil {
			panic(err)
		}
		return exists
	}
	collection := Client.Database(DbName).Collection(CollectionName)
	singleResults := collection.FindOne(context.Background(), bson.M{"_id": r.ID})
	err := singleResults.Err()
	if err == nil {
		return true
	}

	if err == mongo.ErrNoDocuments {
		return false
	}
	panic(err)
}

func incrementTagCount(tags []string, val int) {
	if Remote.Get().(bool) {
		InitRemote()
		db := AtlasClient.Database(DbName)
		var err error
		for _, tag := range tags {
			_, err = db.Collection(TagCollectionName).UpdateOne(ctx,
				bson.M{"name": tag},
				bson.M{"$inc": bson.M{"count": val}}, options.Update().SetUpsert(true))
			if err != nil {
				panic(err)
			}
		}

		if _, err := db.Collection(TagCollectionName).DeleteMany(ctx, bson.M{"count": bson.M{"$lt": 1}}); err != nil {
			panic(err)
		}
		return
	}
	if useLocalSQLite() {
		if err := sqliteIncrementTagCount(tags, val); err != nil {
			panic(err)
		}
		return
	}
	db := Client.Database(DbName)
	var err error
	for _, tag := range tags {
		_, err = db.Collection(TagCollectionName).UpdateOne(ctx,
			bson.M{"name": tag},
			bson.M{"$inc": bson.M{"count": val}}, options.Update().SetUpsert(true))
		if err != nil {
			panic(err)
		}
	}

	if _, err := db.Collection(TagCollectionName).DeleteMany(ctx, bson.M{"count": bson.M{"$lt": 1}}); err != nil {
		panic(err)
	}
}

func Update(parser *terminalw.Parser, fromFile bool, fromEditor bool, prev bool) {
	var err error
	var changed bool
	scanner := bufio.NewScanner(os.Stdin)
	id := parser.GetFlagValueDefault("u", "")
	if prev {
		id = ReadInfo(false)
		if !fromFile {
			fromEditor = true
		}
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = scanner.Text()
	}
	newRecord := Record{}
	if newRecord.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}

	newRecord.LoadByID()
	oldTitle := newRecord.Title
	oldTags := newRecord.Tags
	fmt.Print("input the title: ")
	var title string
	if fromEditor {
		newRecord.Title = utilsw.InputWithEditor(oldTitle, UseVsCode)
		if newRecord.Title != oldTitle {
			changed = true
		}
		fmt.Println()
	} else {
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())
		if fromFile {
			title = utilsw.ReadString(title)
		}
		if title != "" {
			changed = true
			newRecord.Title = title
		}
	}
	tags := strings.TrimSpace(utilsw.UserInput("input the tags: ", false))
	tags = strings.ReplaceAll(tags, ",", " ")
	var tagsRunes []rune
	for _, r := range tags {
		if unicode.IsPrint(r) {
			tagsRunes = append(tagsRunes, r)
		}
	}
	tags = string(tagsRunes)
	if tags != "" {
		changed = true
		newRecord.Tags = strw.SplitNoEmpty(tags, " ")
		c := make(chan interface{}, 1)
		defer close(c)
		go func(c chan interface{}) {
			incrementTagCount(oldTags, -1)
			c <- nil
		}(c)
		go func(c chan interface{}) {
			incrementTagCount(newRecord.Tags, 1)
			c <- nil
		}(c)
		<-c
		<-c
	}
	if !changed {
		return
	}
	newRecord.Update(true)
}

func toggleByName(r *Record, fieldName string) {
	rr := reflect.ValueOf(r)
	val := reflect.Indirect(rr).FieldByName(fieldName)
	if val.Bool() {
		val.SetBool(false)
	} else {
		val.SetBool(true)
	}
}

func setValByFielName(r *Record, fieldName string, val bool) {
	rr := reflect.ValueOf(r)
	fieldVal := reflect.Indirect(rr).FieldByName(fieldName)
	fieldVal.SetBool(val)
}

func GetAllTagsModifiedDate(records []*Record) map[string]time.Time {
	m := make(map[string]time.Time)
	for _, r := range records {
		for _, t := range r.Tags {
			if mt, ok := m[t]; !ok || r.ModifiedDate.After(mt) {
				m[t] = r.ModifiedDate
			}
		}
	}
	return m
}

func Insert(fromEditor bool, filename, tagName string) {
	var title string
	var tagsStr string
	var titleSlice []string
	var err error
	tagName = strings.TrimSpace(tagName)
	if filename != "" {
		title = strings.ReplaceAll(filename, ",", " ")
		titleSlice = strw.SplitNoEmpty(title, " ")
		if len(titleSlice) == 1 {
			titleSlice, err = filepath.Glob(title)
			// fmt.Println("here", titleSlice)
			if err != nil {
				panic(err)
			}
		}
		for i := range titleSlice {
			titleSlice[i] = filepath.Base(titleSlice[i]) + "\n" + utilsw.ReadString(titleSlice[i])
		}
	} else if fromEditor {
		fmt.Print("input the title: ")
		title = utilsw.InputWithEditor("", UseVsCode)
		fmt.Println()
	} else {
		title = strings.TrimSpace(utilsw.UserInput("input the title: ", false))
	}
	if len(tagName) == 0 {
		tagsStr = strings.TrimSpace(utilsw.UserInput("input the tags: ", false))
	} else {
		tagsStr = tagName
	}
	tagsStr = strings.ReplaceAll(tagsStr, ",", " ")
	tags := strw.SplitNoEmpty(tagsStr, " ")
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
	if titleSlice == nil {
		titleSlice = []string{title}
	}
	for _, title = range titleSlice {
		r := NewRecord(title, tags...)
		c := make(chan interface{})
		defer close(c)
		go func(chan interface{}) {
			defer func() {
				c <- nil
			}()
			r.Save(true)
		}(c)
		<-c
		fmt.Println("Inserted: ")
		fmt.Println("\tTags:", r.Tags)
		fmt.Println("\tTitle:", strw.SubStringQuiet(r.Title, 0, titleLen))
	}
}

func Toggle(val bool, id string, name string, prev bool) {
	var err error
	var r Record
	id = strings.TrimSpace(id)
	if prev {
		id = ReadInfo(false)
	} else if id == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	r.LoadByID()
	c := make(chan interface{})
	defer close(c)

	var changed bool
	inc := 0
	switch name {
	case Finish:
		if r.Finished != val {
			r.Finished = val
			changed = true
			if val {
				inc = -1
			} else {
				inc = 1
			}
		}
	default:
		panic("unknown name")
	}
	go func(c chan interface{}, inc int) {
		defer func() {
			c <- nil
		}()
		if !changed {
			return
		}
		incrementTagCount(r.Tags, inc)
	}(c, inc)
	<-c
	r.Update(false)
}

func DeleteRecord(id string, prev bool) {
	var err error
	r := Record{}
	id = strings.TrimSpace(id)
	if prev {
		id = ReadInfo(false)
	} else if id == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	r.LoadByID()
	r.Delete()
}

func DeleteRecordByTag(tag string) bool {
	records, _ := ListRecords(-1, false, true, []string{tag}, false, "", true)
	for _, record := range records {
		fmt.Printf("deleting record. id:%s, tag:%v\n", record.ID.String(), record.Tags)
		record.Delete()
	}
	return len(records) > 0
}

func ChangeTitle(fromFile, fromEditor bool, id string, prev bool) {
	var err error
	id = strings.TrimSpace(id)
	r := Record{}
	scanner := bufio.NewScanner(os.Stdin)
	if prev {
		id = ReadInfo(false)
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	c := make(chan interface{})
	defer close(c)
	go func(chan interface{}) {
		r.LoadByID()
		c <- nil
	}(c)
	<-c
	fmt.Print("input the New Title: ")
	if fromEditor {
		newTitle := utilsw.InputWithEditor(r.Title, UseVsCode)
		if newTitle == r.Title {
			fmt.Println("content not changed ")
			return
		}
		r.Title = newTitle
		fmt.Println()
	} else {
		scanner.Scan()
		newTitle := strings.TrimSpace(scanner.Text())
		if newTitle == r.Title {
			fmt.Println("content not changed")
			return
		}
		r.Title = newTitle
		if fromFile {
			newTitle = utilsw.ReadString(r.Title)
			if newTitle == r.Title {
				fmt.Println("content not changed")
				return
			}
			r.Title = newTitle
		}
	}
	go func(c chan interface{}) {
		r.Update(true)
		c <- nil
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println("\tTags:", r.Tags)
	fmt.Println("\tTitle:", strw.SubStringQuiet(r.Title, 0, titleLen))
}

func AddTag(add bool, id string, prev bool) {
	var err error
	id = strings.TrimSpace(id)
	scanner := bufio.NewScanner(os.Stdin)
	if prev {
		id = ReadInfo(false)
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	r := Record{}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	c := make(chan interface{})
	defer close(c)
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		r.LoadByID()
	}(c)
	<-c
	go func(c chan interface{}) {
		s := cw.NewSet()
		for _, tag := range r.Tags {
			s.Add(tag)
		}
		c <- s
	}(c)
	fmt.Print("input the Tag: ")
	scanner.Scan()
	newTags := strw.SplitNoEmpty(strings.ReplaceAll(strings.TrimSpace(scanner.Text()), ",", " "), " ")

	s := (<-c).(*cw.Set)
	if s.Size() == 1 && !add {
		panic("can't delete the tag, because it's the only tag")
	}
	newTagSet := cw.NewSet()
	for _, newTag := range newTags {
		if strings.TrimSpace(newTag) == "" {
			continue
		}
		newTagSet.Add(newTag)
		if add {
			s.Add(newTag)
		} else {
			s.Delete(newTag)
		}
	}
	var incVal int
	if add {
		incVal = 1
	} else {
		incVal = -1
	}
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		// fmt.Println("here", incVal, newTagSet.ToStringSlice())
		incrementTagCount(newTagSet.ToStringSlice(), incVal)
	}(c)
	<-c
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		r.Tags = s.ToStringSlice()
		r.Update(false)
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println("\tTags:", r.Tags)
	fmt.Println("\tTitle:", strw.SubStringQuiet(r.Title, 0, titleLen))
}

func ColoringRecord(r *Record, p *regexp.Regexp) {
	if p != nil {
		all := bytes.NewBufferString("")
		indices := p.FindAllStringIndex(r.Title, -1)
		beg := cw.NewQueue[int]()
		end := cw.NewQueue[int]()
		bt := []byte(r.Title)
		for _, idx := range indices {
			i, j := idx[0], idx[1]
			beg.Enqueue(i)
			end.Enqueue(j)
		}
		idx := 0
		for !beg.Empty() {
			i := beg.Dequeue()
			j := end.Dequeue()
			all.WriteString(color.HiWhiteString(string(bt[idx:i])))
			all.WriteString(color.RedString(string(bt[i:j])))
			idx = j
		}
		all.WriteString(color.HiWhiteString(string(bt[idx:])))
		r.Title = all.String()
	} else {
		r.Title = color.HiWhiteString(r.Title)
	}
	r.Title = "\n" + r.Title
	for i := range r.Tags {
		r.Tags[i] = color.HiGreenString(r.Tags[i])
	}
}

func SyncByID(id string, push, quiet bool) {
	InitRemote()
	scanner := bufio.NewScanner(os.Stdin)
	var msg string
	if id == "" {
		fmt.Print("Input the ObjectID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	remoteBackUp := Remote.Get().(bool)
	defer Remote.Set(remoteBackUp)
	var r Record
	r.ID = hexID

	if push {
		msg = "push"
		Remote.Set(false)
	} else {
		msg = "pull"
		Remote.Set(true)
	}
	r.LoadByID()
	if r.Invalid {
		panic(fmt.Sprintf("record %s not found", id))
	}

	if push {
		if err = AtlasClient.Database(DbName).Collection(CollectionName).FindOne(ctx, bson.M{"_id": hexID}).Err(); err != nil && err != mongo.ErrNoDocuments {
			panic(err)
		}
		Remote.Set(true)
		if err == mongo.ErrNoDocuments {
			r.Save(true)
		} else {
			r.Update(false)
		}
	} else {
		exists, localErr := sqliteRecordExists(hexID)
		if localErr != nil {
			panic(localErr)
		}
		Remote.Set(false)
		if !exists {
			r.Save(true)
		} else {
			r.Update(false)
		}
	}
	n := 70
	if quiet {
		n = 20
	}
	fmt.Printf("finished %s %s: \n", msg, color.GreenString(strw.SubStringQuiet(r.Title, 0, n)))
	// printSeperator()
	// fmt.Println(r)
	// printSeperator()
}

func GetObjectIdByTags(tags []string, includeFinished bool) string {
	// check if the tags are objectid
	if len(tags) == 1 {
		tag := tags[0]
		if bson.IsObjectIdHex(tag) {
			return tag
		}
	}
	if len(tags) > 0 {
		ListRecords(-1, true, includeFinished, tags, false, "", false)
	}
	id := ReadInfo(false)
	return id
}

func doRecordsByTagsByAction(tags []string, name string) {
	rs, _ := ListRecords(-1, false, true, tags, false, "", true)
	for _, r := range rs {
		// r.Finished = true
		setValByFielName(r, name, true)
		r.Update(false)
	}
}

func FinishRecordsByTags(tags []string) {
	doRecordsByTagsByAction(tags, "Finished")
}

func FilterTags(tags []Tag, prefix []string) []Tag {
	if len(prefix) == 0 {
		return tags
	}
	res := make([]Tag, 0, len(tags))
	for _, t := range tags {
		exclude := false
		for _, p := range prefix {
			if strings.HasPrefix(t.Name, p) {
				exclude = true
				break
			}
		}
		if !exclude {
			res = append(res, t)
		}

	}
	return res
}

func PrintSeperator() {
	fmt.Println(color.BlueString(strings.Repeat("~", 20)))
}
