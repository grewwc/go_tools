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
	uri           string
	clientOptions = &options.ClientOptions{}
	Ctx           context.Context
	Client        *mongo.Client
	AtlasClient   *mongo.Client
)

var (
	Remote = utilsw.NewThreadSafeVal(false)
	mu     sync.Mutex
)

var (
	ListSpecial = false
	UseVsCode   = false
	OnlyTags    = false
)

func InitRemote() {
	mu.Lock()
	if AtlasClient != nil {
		mu.Unlock()
		return
	}
	fmt.Println("connecting to Remote...")
	m := utilsw.GetAllConfig()
	var err error
	// mongo atlas init
	atlasURI := m.GetOrDefault(atlasMongoConfigName, "").(string)
	if atlasURI != "" {
		clientOptions = options.Client().ApplyURI(atlasURI)
		AtlasClient, err = mongo.Connect(Ctx, clientOptions)
		if err != nil {
			panic(err)
		}
	}
	// check if tags and memo collections exists
	db := AtlasClient.Database(DbName)
	if !CollectionExists(db, Ctx, TagCollectionName) {
		db.Collection(TagCollectionName).Indexes().CreateOne(Ctx, mongo.IndexModel{
			Keys:    bson.D{bson.DocElem{Name: "name", Value: "text"}}.Map(),
			Options: options.Index().SetUnique(true),
		})
	}
	fmt.Println("connected")
	mu.Unlock()
	Remote.Set(true)
	// fmt.Println("init Atlas", atlasURI, atlasClient)
}

func init() {
	// var cancel context.CancelFunc
	// get the uri
	m := utilsw.GetAllConfig()
	uriFromConfig := m.GetOrDefault(localMongoConfigName, "")
	if uriFromConfig != "" {
		uri = uriFromConfig.(string)
		clientOptions.ApplyURI(uri)
	}

	// init client
	Ctx = context.Background()
	clientOptions.SetMaxPoolSize(10)
	var err error
	Client, err = mongo.Connect(Ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	// read the special tag patters from .configW
	for _, val := range strw.SplitNoEmpty(m.GetOrDefault(specialTagConfigname, "").(string), ",") {
		val = strings.TrimSpace(val)
		SpecialTagPatterns.Add(val)
	}

}

func ListRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool, title string, prefix bool) ([]*Record, bool) {
	if tags == nil {
		tags = []string{}
	}
	if limit <= 0 {
		limit = math.MaxInt64
	}
	reverseNum := 1
	if reverse {
		reverseNum = -1
	}
	var collection *mongo.Collection
	if !Remote.Get().(bool) {
		collection = Client.Database(DbName).Collection(CollectionName)
	} else {
		InitRemote()
		collection = AtlasClient.Database(DbName).Collection(CollectionName)
	}
	modifiedDataOption := options.Find()
	addDateOption := options.Find()
	modifiedDataOption.SetLimit(limit)
	addDateOption.SetLimit(limit)
	modifiedDataOption.SetSort(bson.M{"modified_date": reverseNum})
	addDateOption.SetSort(bson.M{"add_date": reverseNum})
	m := bson.M{}
	// construct search filter
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

	cursor, err := collection.Find(Ctx, m, addDateOption, modifiedDataOption)
	if err != nil {
		panic(err)
	}
	var res []*Record
	if err = cursor.All(Ctx, &res); err != nil {
		panic(err)
	}
	// filter by special tags
	// fmt.Println("here", listSpecial, tags)
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
	// fmt.Println("here", recordIDs)
	// os.Exit(0)
	written := WriteInfo(recordIDs, recordTitles)
	return res, written
}

func (r *Record) exists() bool {
	var collection *mongo.Collection
	if !Remote.Get().(bool) {
		collection = Client.Database(DbName).Collection(CollectionName)
	} else {
		collection = AtlasClient.Database(DbName).Collection(CollectionName)
	}
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

func incrementTagCount(db *mongo.Database, tags []string, val int) {
	session, err := Client.StartSession()
	if err != nil {
		panic(err)
	}
	if err = session.StartTransaction(); err != nil {
		panic(err)
	}

	for _, tag := range tags {
		_, err = db.Collection(TagCollectionName).UpdateOne(Ctx,
			bson.M{"name": tag},
			bson.M{"$inc": bson.M{"count": val}}, options.Update().SetUpsert(true))
		if err != nil {
			session.AbortTransaction(Ctx)
			panic(err)
		}
	}

	if _, err := db.Collection(TagCollectionName).DeleteMany(Ctx, bson.M{"count": bson.M{"$lt": 1}}); err != nil {
		session.AbortTransaction(Ctx)
		panic(err)
	}
	session.CommitTransaction(Ctx)
}

func Update(parser *terminalw.Parser, fromFile bool, fromEditor bool, prev bool) {
	var err error
	var changed bool
	var cli = Client
	if Remote.Get().(bool) {
		cli = AtlasClient
	}
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
			incrementTagCount(cli.Database(DbName), oldTags, -1)
			c <- nil
		}(c)
		go func(c chan interface{}) {
			incrementTagCount(cli.Database(DbName), newRecord.Tags, 1)
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
	var cli = Client
	if Remote.Get().(bool) {
		cli = AtlasClient
	}
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
		incrementTagCount(cli.Database(DbName), r.Tags, inc)
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
	var cli = Client
	if Remote.Get().(bool) {
		cli = AtlasClient
	}
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
		incrementTagCount(cli.Database(DbName), newTagSet.ToStringSlice(), incVal)
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
	remoteBackUp := Remote
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
	var r Record
	r.ID = hexID
	remoteClient := AtlasClient

	if !push {
		msg = "pull"
		remoteClient = Client
		Remote.Set(true)
		r.LoadByID()
		Remote = remoteBackUp
	} else {
		msg = "push"
		r.LoadByID()
	}

	if err = remoteClient.Database(DbName).Collection(CollectionName).FindOne(Ctx, bson.M{"_id": hexID}).Err(); err != nil && err != mongo.ErrNoDocuments {
		panic(err)
	}

	// 保存的时候，remote需要重新设置
	if push {
		Remote.Set(true)
	} else {
		Remote.Set(false)
	}
	if err == mongo.ErrNoDocuments {
		r.Save(true)
	} else {
		r.Update(false)
	}
	// 恢复remote
	Remote = remoteBackUp
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
