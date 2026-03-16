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
	"strconv"
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
	specialTagConfigname = "special.tags"

	DefaultBackendConfigName    = "re.backend"
	DefaultRemoteHostConfigName = "re.remote.host"
)

const (
	LocalBackendAuto   = "auto"
	LocalBackendMongo  = "mongo"
	LocalBackendSQLite = "sqlite"

	defaultLocalMongoURI  = "mongodb://localhost:27017"
	localMongoInitTimeout = 1500 * time.Millisecond
	remoteMongoTimeout    = 10 * time.Second
	remoteSQLiteShellPath = "$HOME/.go_tools_memo.sqlite3"
	sshConnectTimeoutSec  = 8
	remoteCmdTimeout      = 2 * time.Minute
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
	DefaultRemoteHost  = utilsw.GetConfig(DefaultRemoteHostConfigName, "")
)

var (
	ctx         context.Context
	Client      *mongo.Client
	AtlasClient *mongo.Client
)

var (
	Remote           = utilsw.NewThreadSafeVal(false)
	localBackendMode = utilsw.NewThreadSafeVal(LocalBackendAuto)
	remoteMongoURI   = utilsw.NewThreadSafeVal("")
	remoteMongoHost  = utilsw.NewThreadSafeVal("")
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
	host := strings.TrimSpace(remoteMongoHost.Get().(string))
	if host == "" {
		panic("remote host not configured, pass --host <ip[:port]> or set .configW:re.remote.host")
	}
	uri := normalizeMongoURI(host)
	if AtlasClient != nil && remoteMongoURI.Get().(string) == uri {
		Remote.Set(true)
		return
	}
	if AtlasClient != nil {
		_ = AtlasClient.Disconnect(context.Background())
		AtlasClient = nil
	}
	fmt.Printf("connecting to Remote %s...\n", host)
	connectCtx, cancel := context.WithTimeout(context.Background(), remoteMongoTimeout)
	defer cancel()
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetConnectTimeout(remoteMongoTimeout)
	clientOptions.SetServerSelectionTimeout(remoteMongoTimeout)
	client, err := mongo.Connect(connectCtx, clientOptions)
	if err == nil {
		err = client.Ping(connectCtx, readpref.Primary())
	}
	if err != nil {
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		panic(fmt.Sprintf("failed to connect to remote host %s: %v", host, err))
	}
	AtlasClient = client
	remoteMongoURI.Set(uri)
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
	// fmt.Println("init remote", uri, AtlasClient)
}

func SetRemoteHost(host string) {
	remoteMongoHost.Set(strings.TrimSpace(host))
}

func normalizeMongoURI(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.HasPrefix(host, "mongodb://") || strings.HasPrefix(host, "mongodb+srv://") {
		return host
	}
	if !strings.Contains(host, ":") {
		host += ":27017"
	}
	return "mongodb://" + host
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

func recordPrimaryTitle(title string) string {
	title = strings.ReplaceAll(title, "\r", "")
	for _, line := range strings.Split(title, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return strings.TrimSpace(title)
}

func appendUniqueRecords(dst []*Record, seen map[primitive.ObjectID]struct{}, groups ...[]*Record) []*Record {
	for _, group := range groups {
		for _, r := range group {
			if _, ok := seen[r.ID]; ok {
				continue
			}
			seen[r.ID] = struct{}{}
			dst = append(dst, r)
		}
	}
	return dst
}

func filterExactTitleMatches(records []*Record, ref string) []*Record {
	res := make([]*Record, 0, len(records))
	for _, r := range records {
		if strings.EqualFold(strings.TrimSpace(r.Title), ref) || strings.EqualFold(recordPrimaryTitle(r.Title), ref) {
			res = append(res, r)
		}
	}
	return res
}

func listRecordsByLiteralTitle(ref string, remote bool) []*Record {
	query := ref
	if remote {
		query = regexp.QuoteMeta(ref)
	}
	records, _ := listRecords(math.MaxInt64, false, true, nil, false, query, false, false)
	return records
}

func chooseRecordID(ref string, records []*Record) string {
	if len(records) == 0 {
		panic(fmt.Sprintf("record %q not found", ref))
	}
	if len(records) == 1 {
		return records[0].ID.Hex()
	}
	ids := make([]*primitive.ObjectID, 0, len(records))
	titles := make([]string, 0, len(records))
	for _, r := range records {
		ids = append(ids, &r.ID)
		titles = append(titles, r.Title)
	}
	WriteInfo(ids, titles)
	fmt.Printf("multiple records matched %q, choose one:\n", ref)
	return ReadInfo(false)
}

func resolveRecordReferenceID(ref string, remote bool) string {
	scanner := bufio.NewScanner(os.Stdin)
	ref = strings.TrimSpace(ref)
	if ref == "" {
		fmt.Print("Input the ObjectID/title/tag: ")
		scanner.Scan()
		ref = strings.TrimSpace(scanner.Text())
	}
	if IsObjectID(ref) {
		return ref
	}

	remoteBackup := Remote.Get().(bool)
	listSpecialBackup := ListSpecial
	defer func() {
		Remote.Set(remoteBackup)
		ListSpecial = listSpecialBackup
	}()
	ListSpecial = true
	if remote {
		InitRemote()
		Remote.Set(true)
	} else {
		Remote.Set(false)
	}

	titleMatches := listRecordsByLiteralTitle(ref, remote)
	exactTitleMatches := filterExactTitleMatches(titleMatches, ref)
	exactTagMatches, _ := listRecords(math.MaxInt64, false, true, []string{ref}, false, "", false, false)

	seen := make(map[primitive.ObjectID]struct{})
	exactMatches := appendUniqueRecords(nil, seen, exactTitleMatches, exactTagMatches)
	if len(exactMatches) > 0 {
		return chooseRecordID(ref, exactMatches)
	}

	seen = make(map[primitive.ObjectID]struct{})
	tagMatches, _ := listRecords(math.MaxInt64, false, true, []string{ref}, false, "", true, false)
	fuzzyMatches := appendUniqueRecords(nil, seen, titleMatches, tagMatches)
	return chooseRecordID(ref, fuzzyMatches)
}

func parseSSHHostSpec(spec string) (target, port string) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return "", ""
	}
	idx := strings.LastIndex(spec, ":")
	if idx <= strings.LastIndex(spec, "@") {
		return spec, ""
	}
	if idx < 0 || strings.Count(spec, ":") > 1 && !strings.Contains(spec, "@") {
		return spec, ""
	}
	if _, err := strconv.Atoi(spec[idx+1:]); err != nil {
		return spec, ""
	}
	return spec[:idx], spec[idx+1:]
}

func buildCommandLine(name string, args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, strconv.Quote(name))
	for _, arg := range args {
		parts = append(parts, strconv.Quote(arg))
	}
	return strings.Join(parts, " ")
}

func runSSHCommand(host, command string) (string, error) {
	args, _, err := sshTargetArgs(host)
	if err != nil {
		return "", err
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return "", fmt.Errorf("empty remote command")
	}
	args = append(args, command)
	return utilsw.RunCmdWithTimeout(buildCommandLine("ssh", args...), remoteCmdTimeout)
}

func buildRemoteSQLiteReplaceCommand() string {
	sidecars := sqliteCompanionPaths(remoteSQLiteShellPath)
	return fmt.Sprintf(
		"rm -f -- %s %s && mv -- %s %s",
		sidecars[0],
		sidecars[1],
		remoteSQLiteShellPath+".incoming",
		remoteSQLiteShellPath,
	)
}

func sshTargetArgs(host string) ([]string, string, error) {
	target, port := parseSSHHostSpec(host)
	if target == "" {
		return nil, "", fmt.Errorf("empty host")
	}
	args := []string{"-o", fmt.Sprintf("ConnectTimeout=%d", sshConnectTimeoutSec)}
	if port != "" {
		args = append(args, "-p", port)
	}
	args = append(args, target)
	return args, target, nil
}

func scpTargetArgs(host string) ([]string, string, error) {
	target, port := parseSSHHostSpec(host)
	if target == "" {
		return nil, "", fmt.Errorf("empty host")
	}
	args := []string{"-o", fmt.Sprintf("ConnectTimeout=%d", sshConnectTimeoutSec)}
	if port != "" {
		args = append(args, "-P", port)
	}
	return args, target, nil
}

func remoteFileExists(host, remotePath string) bool {
	_, err := runSSHCommand(host, fmt.Sprintf("test -f %s", remotePath))
	return err == nil
}

func copyRemoteFileToLocal(host, remotePath, dest string) error {
	args, target, err := scpTargetArgs(host)
	if err != nil {
		return err
	}
	args = append(args, fmt.Sprintf("%s:%s", target, remotePath), dest)
	if output, err := utilsw.RunCmdWithTimeout(buildCommandLine("scp", args...), remoteCmdTimeout); err != nil {
		if output != "" {
			return fmt.Errorf("failed to download remote sqlite: %s", output)
		}
		return fmt.Errorf("failed to download remote sqlite: %v", err)
	}
	return nil
}

func copyLocalFileToRemote(host, src, remotePath string) error {
	args, target, err := scpTargetArgs(host)
	if err != nil {
		return err
	}
	args = append(args, src, fmt.Sprintf("%s:%s", target, remotePath))
	if output, err := utilsw.RunCmdWithTimeout(buildCommandLine("scp", args...), remoteCmdTimeout); err != nil {
		if output != "" {
			return fmt.Errorf("failed to upload remote sqlite: %s", output)
		}
		return fmt.Errorf("failed to upload remote sqlite: %v", err)
	}
	return nil
}

func replaceRemoteSQLite(host string) error {
	if output, err := runSSHCommand(host, buildRemoteSQLiteReplaceCommand()); err != nil {
		if output != "" {
			return fmt.Errorf("failed to replace remote sqlite: %s", output)
		}
		return fmt.Errorf("failed to replace remote sqlite: %v", err)
	}
	return nil
}

func remoteSQLiteExists(host string) bool {
	return remoteFileExists(host, remoteSQLiteShellPath)
}

func pullRemoteSQLiteToTemp(host, dest string, required bool) error {
	if !remoteSQLiteExists(host) {
		if required {
			return fmt.Errorf("remote sqlite not found on %s (%s)", host, defaultLocalSQLite)
		}
		return nil
	}
	if err := copyRemoteFileToLocal(host, defaultLocalSQLite, dest); err != nil {
		return err
	}
	remoteSidecars := sqliteCompanionPaths(defaultLocalSQLite)
	remoteShellSidecars := sqliteCompanionPaths(remoteSQLiteShellPath)
	localSidecars := sqliteSidecarPaths(dest)
	for i, remotePath := range remoteSidecars {
		if !remoteFileExists(host, remoteShellSidecars[i]) {
			continue
		}
		if err := copyRemoteFileToLocal(host, remotePath, localSidecars[i]); err != nil {
			return err
		}
	}
	return nil
}

func pushTempSQLiteToRemote(host, src string) error {
	if err := prepareSQLitePathForTransfer(src); err != nil {
		return err
	}
	if err := copyLocalFileToRemote(host, src, defaultLocalSQLite+".incoming"); err != nil {
		return err
	}
	if err := replaceRemoteSQLite(host); err != nil {
		return err
	}
	return nil
}

// loadRecordFromCurrentLocal reads from the active local backend (mongo or sqlite)
// before the remote sqlite staging path temporarily switches the process into sqlite mode.
func loadRecordFromCurrentLocal(ref string) *Record {
	resolvedID := resolveRecordReferenceID(ref, false)
	hexID, err := primitive.ObjectIDFromHex(resolvedID)
	if err != nil {
		panic(err)
	}
	remoteBackup := Remote.Get().(bool)
	defer Remote.Set(remoteBackup)
	Remote.Set(false)
	r := &Record{ID: hexID}
	r.LoadByID()
	if r.Invalid {
		panic(fmt.Sprintf("record %q not found", ref))
	}
	return r
}

func loadRecordFromSQLitePath(ref, sqlitePath string) (*Record, error) {
	var loaded *Record
	err := withSQLitePath(sqlitePath, func() error {
		resolvedID := resolveRecordReferenceID(ref, false)
		hexID, err := primitive.ObjectIDFromHex(resolvedID)
		if err != nil {
			return err
		}
		loaded, err = sqliteLoadRecord(hexID)
		if err != nil {
			return err
		}
		if loaded == nil {
			return fmt.Errorf("record %q not found in remote sqlite", ref)
		}
		return nil
	})
	return loaded, err
}

func saveRecordToSQLitePath(r *Record, sqlitePath string) error {
	return withSQLitePath(sqlitePath, func() error {
		return sqliteUpsertRecord(r)
	})
}

func upsertRecordToCurrentLocal(r *Record) error {
	remoteBackup := Remote.Get().(bool)
	defer Remote.Set(remoteBackup)
	Remote.Set(false)
	if useLocalSQLite() {
		return sqliteUpsertRecord(r)
	}
	if err := initLocalMongo(); err != nil {
		return err
	}
	collection := Client.Database(DbName).Collection(CollectionName)
	err := collection.FindOne(ctx, bson.M{"_id": r.ID}).Err()
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}
	if err == mongo.ErrNoDocuments {
		r.Save(true)
		return nil
	}
	r.Update(false)
	return nil
}

func SyncByHost(ref, host string, push, quiet bool) {
	host = strings.TrimSpace(host)
	if host == "" {
		panic("host is required")
	}
	tmpDir, err := os.MkdirTemp("", "go_tools_remote_sqlite_")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)
	tmpSQLite := filepath.Join(tmpDir, "memo.sqlite3")
	if err = pullRemoteSQLiteToTemp(host, tmpSQLite, !push); err != nil {
		panic(err)
	}

	var msg string
	var record *Record
	if push {
		msg = "push"
		record = loadRecordFromCurrentLocal(ref)
		if err = saveRecordToSQLitePath(record, tmpSQLite); err != nil {
			panic(err)
		}
		if err = pushTempSQLiteToRemote(host, tmpSQLite); err != nil {
			panic(err)
		}
	} else {
		msg = "pull"
		record, err = loadRecordFromSQLitePath(ref, tmpSQLite)
		if err != nil {
			panic(err)
		}
		if err = upsertRecordToCurrentLocal(record); err != nil {
			panic(err)
		}
	}
	n := 70
	if quiet {
		n = 20
	}
	fmt.Printf("finished %s %s: \n", msg, color.GreenString(strw.SubStringQuiet(record.Title, 0, n)))
}

func SyncByID(id string, push, quiet bool) {
	var msg string
	id = resolveRecordReferenceID(id, !push)
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		panic(err)
	}
	InitRemote()
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

// SyncByIDToHost pushes a local record to the managed SQLite database on a remote host.
// The remote database path is managed by re and defaults to ~/.go_tools_memo.sqlite3 on that host.
func SyncByIDToHost(id, targetHost string, quiet bool) {
	SyncByHost(id, targetHost, true, quiet)
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
