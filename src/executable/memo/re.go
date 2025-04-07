package main

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
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/sortW"

	"github.com/grewwc/go_tools/src/executable/memo/_helpers"
	"github.com/grewwc/go_tools/src/strW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	dbName            = "daily"
	collectionName    = "memo"
	tagCollectionName = "tag"
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
	defaultTxtOutputName = "output.txt"
	outputName           = "output_binary"
	finish               = "finish"
	hold                 = "hold"
	myproblem            = "myproblem"
	titleLen             = 200
)

var (
	txtOutputName      = "output.txt"
	specialTagPatterns = containerW.NewSet("learn")
)

var (
	uri           string
	clientOptions = &options.ClientOptions{}
	ctx           context.Context
	client        *mongo.Client
	atlasClient   *mongo.Client
)

var (
	remote = utilsW.NewThreadSafeVal(false)
	mu     sync.Mutex
)

var (
	listSpecial = false
	useVsCode   = false
)

type record struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty" json:"id,ignoreempty"`
	Tags         []string           `bson:"tags,ignoreempty" json:"tags,ignoreempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty" json:"add_date,ignoreempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty" json:"modified_date,ignoreempty"`
	MyProblem    bool               `bson:"my_problem,ignoreempty" json:"my_problem,ignoreempty"`
	Finished     bool               `bson:"finished,ignoreempty" json:"finished,ignoreempty"`
	Hold         bool               `bson:"hold,ignoreempty" json:"hold,ignoreempty"`
	Title        string             `bson:"title,ignoreempty" json:"title,ignoreempty"`
}

type tag struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty"`
	Name         string             `bson:"name,ignoreempty"`
	Count        int64              `bson:"count,ignoreempty"`
	modifiedDate time.Time
}

func (t tag) String() string {
	return utilsW.ToString(t, "ID", "Name", "Count")
}

func newRecord(title string, tags ...string) *record {
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
	r := &record{Title: title, Tags: tags, Finished: false, MyProblem: true}
	t := time.Now()
	r.AddDate = t
	r.ModifiedDate = t
	r.ID = primitive.NewObjectID()
	return r
}

func init() {
	// var cancel context.CancelFunc
	// get the uri
	m := utilsW.GetAllConfig()
	uriFromConfig := m.GetOrDefault(localMongoConfigName, "")
	if uriFromConfig != "" {
		uri = uriFromConfig.(string)
	}

	// init client
	ctx = context.Background()
	clientOptions.SetMaxPoolSize(10)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(err)
	}

	// read the special tag patters from .configW
	for _, val := range strW.SplitNoEmpty(m.GetOrDefault(specialTagConfigname, "").(string), ",") {
		specialTagPatterns.Add(val)
	}

}

func initAtlas() {
	mu.Lock()
	if atlasClient != nil {
		mu.Unlock()
		return
	}
	fmt.Println("connecting to Mongo Atlas...")
	m := utilsW.GetAllConfig()
	var err error
	// mongo atlas init
	atlasURI := m.GetOrDefault(atlasMongoConfigName, "").(string)
	if atlasURI != "" {
		clientOptions = options.Client().ApplyURI(atlasURI)
		atlasClient, err = mongo.Connect(ctx, clientOptions)
		if err != nil {
			panic(err)
		}
	}
	// check if tags and memo collections exists
	db := atlasClient.Database(dbName)
	if !_helpers.CollectionExists(db, ctx, tagCollectionName) {
		db.Collection(tagCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.DocElem{Name: "name", Value: "text"}}.Map(),
			Options: options.Index().SetUnique(true),
		})
	}
	fmt.Println("connected")
	mu.Unlock()
	// fmt.Println("init Atlas", atlasURI, atlasClient)
}

func (r record) String() string {
	return utilsW.ToString(r, "AddDate", "ModifiedDate")
}

func incrementTagCount(db *mongo.Database, tags []string, val int) {
	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	if err = session.StartTransaction(); err != nil {
		panic(err)
	}

	for _, tag := range tags {
		_, err = db.Collection(tagCollectionName).UpdateOne(ctx,
			bson.M{"name": tag},
			bson.M{"$inc": bson.M{"count": val}}, options.Update().SetUpsert(true))
		if err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
	}

	if _, err := db.Collection(tagCollectionName).DeleteMany(ctx, bson.M{"count": bson.M{"$lt": 1}}); err != nil {
		session.AbortTransaction(ctx)
		panic(err)
	}
	session.CommitTransaction(ctx)
}

func (r *record) exists() bool {
	var collection *mongo.Collection
	if !remote.Get().(bool) {
		collection = client.Database(dbName).Collection(collectionName)
	} else {
		collection = atlasClient.Database(dbName).Collection(collectionName)
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

func (r *record) do(action string, options ...string) {
	var err error
	var db *mongo.Database
	if !remote.Get().(bool) {
		db = client.Database(dbName)
	} else {
		db = atlasClient.Database(dbName)
	}
	collection := db.Collection(collectionName)
	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	if err = session.StartTransaction(); err != nil {
		panic(err)
	}
	switch action {
	case "load":
		if err = collection.FindOne(ctx, bson.M{"_id": r.ID}).Decode(r); err != nil {
			panic(err)
		}

	case "save":
		if r.exists() {
			return
		}
		noUpdateModifiedDate := false
		if len(options) > 0 {
			s := containerW.NewSet()
			for _, option := range options {
				s.Add(option)
			}
			if s.Contains("noUpdateModifiedDate") {
				noUpdateModifiedDate = true
			}
		}
		if !noUpdateModifiedDate {
			r.ModifiedDate = time.Now()
		}
		if _, err = collection.InsertOne(context.Background(), r); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
		incrementTagCount(db, r.Tags, 1)
	case "delete":
		if _, err = collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
		incrementTagCount(db, r.Tags, -1)
	case "deleteByID":
		if _, err = collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
		incrementTagCount(db, r.Tags, -1)
	case "update":
		if _, err = collection.UpdateOne(ctx, bson.M{"_id": r.ID}, bson.M{"$set": r}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}

	default:
		panic("unknow action " + action)
	}

	if err = session.CommitTransaction(ctx); err != nil {
		panic(err)
	}
}

func (r *record) save(noUpdateModifiedDate bool) {
	if noUpdateModifiedDate {
		r.do("save", "noUpdateModifiedDate")
	} else {
		r.do("save")
	}
}

func (r *record) delete() {
	r.do("delete")
}

func (r *record) deleteByID() {
	r.do("deleteByID")
}

func (r *record) update(changeModifiedDate bool) {
	if changeModifiedDate {
		r.ModifiedDate = time.Now()
	}
	r.do("update")
}

func (r *record) loadByID() {
	r.do("load")
}

func toggleByName(r *record, fieldName string) {
	rr := reflect.ValueOf(r)
	val := reflect.Indirect(rr).FieldByName(fieldName)
	if val.Bool() {
		val.SetBool(false)
	} else {
		val.SetBool(true)
	}
}

func setValByFielName(r *record, fieldName string, val bool) {
	rr := reflect.ValueOf(r)
	fieldVal := reflect.Indirect(rr).FieldByName(fieldName)
	fieldVal.SetBool(val)
}

func listRecords(limit int64, reverse, includeFinished bool, includeHold bool, tags []string, useAnd bool, title string,
	onlyMyproblem, onlyHold bool, prefix bool) ([]*record, bool) {
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
	if !remote.Get().(bool) {
		collection = client.Database(dbName).Collection(collectionName)
	} else {
		initAtlas()
		collection = atlasClient.Database(dbName).Collection(collectionName)
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
	if !includeHold {
		m["$or"] = []interface{}{bson.M{"hold": false}, bson.M{"hold": nil}}
	}
	if onlyMyproblem {
		m["my_problem"] = true
	}
	if onlyHold {
		m["hold"] = true
		delete(m, "$or")
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
	var res []*record
	if err = cursor.All(ctx, &res); err != nil {
		panic(err)
	}
	// filter by special tags
	if !listSpecial {
		resCopy := make([]*record, 0, len(res))
		for _, r := range res {
			trie := containerW.NewTrie()
			tags := r.Tags
			for _, t := range tags {
				trie.Insert(t)
			}
			if !_helpers.SearchTrie(trie, specialTagPatterns) {
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
	written := _helpers.WriteInfo(recordIDs, recordTitles)
	return res, written
}

func update(parser *terminalW.Parser, fromFile bool, fromEditor bool, prev bool) {
	var err error
	var changed bool
	var cli = client
	if remote.Get().(bool) {
		cli = atlasClient
	}
	scanner := bufio.NewScanner(os.Stdin)
	id := parser.GetFlagValueDefault("u", "")
	if prev {
		id = _helpers.ReadInfo(false)
		if !fromFile {
			fromEditor = true
		}
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = scanner.Text()
	}
	newRecord := record{}
	if newRecord.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}

	newRecord.loadByID()
	oldTitle := newRecord.Title
	oldTags := newRecord.Tags
	fmt.Print("input the title: ")
	var title string
	if fromEditor {
		newRecord.Title = utilsW.InputWithEditor(oldTitle, useVsCode)
		if newRecord.Title != oldTitle {
			changed = true
		}
		fmt.Println()
	} else {
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())
		if fromFile {
			title = utilsW.ReadString(title)
		}
		if title != "" {
			changed = true
			newRecord.Title = title
		}
	}
	tags := strings.TrimSpace(utilsW.UserInput("input the tags: ", false))
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
		newRecord.Tags = strW.SplitNoEmpty(tags, " ")
		c := make(chan interface{}, 1)
		go func(c chan interface{}) {
			incrementTagCount(cli.Database(dbName), oldTags, -1)
			c <- nil
		}(c)
		go func(c chan interface{}) {
			incrementTagCount(cli.Database(dbName), newRecord.Tags, 1)
			c <- nil
		}(c)
		<-c
		<-c
	}
	if !changed {
		return
	}
	newRecord.update(true)
}

func getAllTagsModifiedDate(records []*record) map[string]time.Time {
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

func insert(fromEditor bool, filename, tagName string) {
	var title string
	var tagsStr string
	var titleSlice []string
	var err error
	tagName = strings.TrimSpace(tagName)
	if filename != "" {
		title = strings.ReplaceAll(filename, ",", " ")
		titleSlice = strW.SplitNoEmpty(title, " ")
		if len(titleSlice) == 1 {
			titleSlice, err = filepath.Glob(title)
			// fmt.Println("here", titleSlice)
			if err != nil {
				panic(err)
			}
		}
		for i := range titleSlice {
			titleSlice[i] = filepath.Base(titleSlice[i]) + "\n" + utilsW.ReadString(titleSlice[i])
		}
	} else if fromEditor {
		fmt.Print("input the title: ")
		title = utilsW.InputWithEditor("", useVsCode)
		fmt.Println()
	} else {
		title = strings.TrimSpace(utilsW.UserInput("input the title: ", false))
	}
	if len(tagName) == 0 {
		tagsStr = strings.TrimSpace(utilsW.UserInput("input the tags: ", false))
	} else {
		tagsStr = tagName
	}
	tagsStr = strings.ReplaceAll(tagsStr, ",", " ")
	tags := strW.SplitNoEmpty(tagsStr, " ")
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
	if titleSlice == nil {
		titleSlice = []string{title}
	}
	for _, title = range titleSlice {
		r := newRecord(title, tags...)
		c := make(chan interface{})
		go func(chan interface{}) {
			defer func() {
				c <- nil
			}()
			r.save(true)
		}(c)
		<-c
		fmt.Println("Inserted: ")
		fmt.Println("\tTags:", r.Tags)
		fmt.Println("\tTitle:", strW.SubStringQuiet(r.Title, 0, titleLen))
	}
}

func toggle(val bool, id string, name string, prev bool) {
	var err error
	var r record
	var cli = client
	if remote.Get().(bool) {
		cli = atlasClient
	}
	id = strings.TrimSpace(id)
	if prev {
		id = _helpers.ReadInfo(false)
	} else if id == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	r.loadByID()
	c := make(chan interface{})

	var changed bool
	inc := 0
	switch name {
	case finish:
		if r.Finished != val {
			r.Finished = val
			changed = true
			if val {
				inc = -1
			} else {
				inc = 1
			}
		}
	case hold:
		if r.Hold != val {
			r.Hold = val
			changed = true
			if val {
				inc = -1
			} else {
				inc = 1
			}
		}
	case myproblem:
		if r.MyProblem != val {
			changed = true
			r.MyProblem = val
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
		incrementTagCount(cli.Database(dbName), r.Tags, inc)
	}(c, inc)
	<-c
	r.update(false)
}

func deleteRecord(id string, prev bool) {
	var err error
	r := record{}
	id = strings.TrimSpace(id)
	if prev {
		id = _helpers.ReadInfo(false)
	} else if id == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	r.loadByID()
	r.delete()
}

func changeTitle(fromFile, fromEditor bool, id string, prev bool) {
	var err error
	id = strings.TrimSpace(id)
	r := record{}
	scanner := bufio.NewScanner(os.Stdin)
	if prev {
		id = _helpers.ReadInfo(false)
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	c := make(chan interface{})
	go func(chan interface{}) {
		r.loadByID()
		c <- nil
	}(c)
	<-c
	fmt.Print("input the New Title: ")
	if fromEditor {
		newTitle := utilsW.InputWithEditor(r.Title, useVsCode)
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
			newTitle = utilsW.ReadString(r.Title)
			if newTitle == r.Title {
				fmt.Println("content not changed")
				return
			}
			r.Title = newTitle
		}
	}
	go func(c chan interface{}) {
		r.update(true)
		c <- nil
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println("\tTags:", r.Tags)
	fmt.Println("\tTitle:", strW.SubStringQuiet(r.Title, 0, titleLen))
}

func addTag(add bool, id string, prev bool) {
	var err error
	var cli = client
	if remote.Get().(bool) {
		cli = atlasClient
	}
	id = strings.TrimSpace(id)
	scanner := bufio.NewScanner(os.Stdin)
	if prev {
		id = _helpers.ReadInfo(false)
	} else if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = strings.TrimSpace(scanner.Text())
	}
	r := record{}
	if r.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}
	c := make(chan interface{})
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		r.loadByID()
	}(c)
	<-c
	go func(c chan interface{}) {
		s := containerW.NewSet()
		for _, tag := range r.Tags {
			s.Add(tag)
		}
		c <- s
	}(c)
	fmt.Print("input the Tag: ")
	scanner.Scan()
	newTags := strW.SplitNoEmpty(strings.ReplaceAll(strings.TrimSpace(scanner.Text()), ",", " "), " ")

	s := (<-c).(*containerW.Set)
	if s.Size() == 1 && !add {
		panic("can't delete the tag, because it's the only tag")
	}
	newTagSet := containerW.NewSet()
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
		incrementTagCount(cli.Database(dbName), newTagSet.ToStringSlice(), incVal)
	}(c)
	<-c
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		r.Tags = s.ToStringSlice()
		r.update(false)
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println("\tTags:", r.Tags)
	fmt.Println("\tTitle:", strW.SubStringQuiet(r.Title, 0, titleLen))
}

func printSeperator() {
	fmt.Println(color.BlueString(strings.Repeat("~", 20)))
}

func coloringRecord(r *record, p *regexp.Regexp) {
	if p != nil {
		all := bytes.NewBufferString("")
		indices := p.FindAllStringIndex(r.Title, -1)
		beg := containerW.NewQueue()
		end := containerW.NewQueue()
		bt := []byte(r.Title)
		for _, idx := range indices {
			i, j := idx[0], idx[1]
			beg.Enqueue(i)
			end.Enqueue(j)
		}
		idx := 0
		for !beg.Empty() {
			i := beg.Dequeue().(int)
			j := end.Dequeue().(int)
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

func syncByID(id string, push, quiet bool) {
	initAtlas()
	remoteBackUp := remote
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
	var r record
	r.ID = hexID
	remoteClient := atlasClient

	if !push {
		msg = "pull"
		remoteClient = client
		remote.Set(true)
		r.loadByID()
		remote = remoteBackUp
	} else {
		msg = "push"
		r.loadByID()
	}

	if err = remoteClient.Database(dbName).Collection(collectionName).FindOne(ctx, bson.M{"_id": hexID}).Err(); err != nil && err != mongo.ErrNoDocuments {
		panic(err)
	}

	// 保存的时候，remote需要重新设置
	if push {
		remote.Set(true)
	} else {
		remote.Set(false)
	}
	if err == mongo.ErrNoDocuments {
		r.save(true)
	} else {
		r.update(false)
	}
	// 恢复remote
	remote = remoteBackUp
	n := 70
	if quiet {
		n = 20
	}
	fmt.Printf("finished %s %s: \n", msg, color.GreenString(strW.SubStringQuiet(r.Title, 0, n)))
	// printSeperator()
	// fmt.Println(r)
	// printSeperator()
}

func getObjectIdByTags(tags []string) string {
	// check if the tags are objectid
	if len(tags) == 1 {
		tag := tags[0]
		if bson.IsObjectIdHex(tag) {
			return tag
		}
	}
	if len(tags) > 0 {
		listRecords(-1, true, false, false, tags, false, "", true, false, false)
	}
	id := _helpers.ReadInfo(false)
	return id
}

func finishRecordsByTags(tags []string) {
	doRecordsByTagsByAction(tags, "Finished")
}

func holdRecordsByTags(tags []string) {
	doRecordsByTagsByAction(tags, "Hold")
}

func doRecordsByTagsByAction(tags []string, name string) {
	rs, _ := listRecords(-1, false, true, true, tags, false, "", false, false, true)
	for _, r := range rs {
		// r.Finished = true
		setValByFielName(r, name, true)
		r.update(false)
	}
}

func filterTags(tags []tag, prefix []string) []tag {
	if len(prefix) == 0 {
		return tags
	}
	res := make([]tag, 0, len(tags))
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

func main() {
	var n int64 = 100
	parser := terminalW.NewParser()
	parser.Bool("i", false, "insert a record")
	parser.String("ct", "", "change a record title")
	parser.String("u", "", "update a record")
	parser.String("d", "", "delete a record")
	parser.Int("n", 100, "# of records to list")
	parser.Bool("h", false, "print help information")
	parser.String("push", "objectID to push", "push to remote db (may take a while)")
	parser.String("pull", "objectID to pull", "pull from remote db (may take a while)")
	parser.Bool("r", false, "reverse sort")
	parser.Bool("all", false, "including all record")
	parser.Bool("a", false, "shortcut for -all")
	parser.String("f", "", "finish a record")
	parser.String("nf", "", "set a record UNFINISHED")
	parser.String("hold", "", "hold a record for later finish")
	parser.String("unhold", "", "unhold a record (reverse operation for the -hold)")
	parser.String("p", "", "set a record my problem")
	parser.String("np", "", "set a record NOT my problem")
	parser.String("t", "", "search by tags")
	parser.Bool("include-finished", false, "include finished record")
	parser.Bool("include-hold", false, "include held record")
	parser.String("add-tag", "", "add tags for a record")
	parser.String("del-tag", "", "delete tags for a record")
	parser.String("clean-tag", "", "clean all the records having the tag")
	parser.Bool("tags", false, "list all tags")
	parser.Bool("and", false, "use and logic to match tags")
	parser.Bool("v", false, "verbose (show modify/add time, verbose)")
	parser.String("file", "", "read title from a file, for '-u' & '-ct', file serve as bool, for '-i', needs to pass filename")
	parser.Bool("e", false, "read from editor")
	parser.String("title", "", "search by title")
	parser.String("c", "", "content (alias for title)")
	parser.String("out", "", fmt.Sprintf("output to text file (default is %s)", defaultTxtOutputName))
	parser.Bool("my", false, "only list my problem")
	parser.Bool("remote", false, "operate on the remote server")
	parser.Bool("prev", false, "operate based on the previous ObjectIDs")
	parser.Bool("count", false, "only print the count, not the result")
	parser.Bool("prefix", false, "tag prefix")
	parser.Bool("pre", false, "tag prefix (short for -prefix)")
	parser.Bool("binary", false, "if the title is binary file")
	parser.Bool("b", false, "shortcut for -binary")
	parser.Bool("force", false, "force overwrite")
	parser.Bool("sp", false, fmt.Sprintf("if list tags started with special: %v (config in .configW->special.tags)", specialTagPatterns.ToSlice()))
	parser.Bool("onlyhold", false, "list only the hold rerods")
	parser.String("ex", "", "exclude some tag prefix when list tags")
	parser.Bool("code", false, "if use vscode as input editor (default false)")

	parser.ParseArgsCmd("h", "r", "all", "a",
		"i", "include-finished", "tags", "and", "v", "e", "my", "remote", "prev", "count", "prefix", "binary", "b",
		"sp", "include-held", "onlyhold", "p", "code", "pre", "force")
	// fmt.Println(parser.Optional)
	// default behavior
	// re
	if parser.Empty() {
		records, _ := listRecords(n, false, false, false, []string{"todo", "urgent"}, false, "", true, false, true)
		for _, record := range records {
			printSeperator()
			coloringRecord(record, nil)
			fmt.Println(record)
			fmt.Println(color.HiRedString(record.ID.String()))
		}
		return
	}

	positional := parser.Positional
	prefix := parser.ContainsAnyFlagStrict("prefix", "pre")
	isWindows := utilsW.WINDOWS == utilsW.GetPlatform()
	onlyHold := parser.ContainsFlagStrict("onlyhold") ||
		(parser.ContainsFlagStrict("hold") && parser.GetFlagValueDefault("hold", "") == "")

	if parser.ContainsFlagStrict("remote") {
		initAtlas()
		remote.Set(true)
	}

	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	if parser.ContainsAllFlagStrict("code") {
		useVsCode = true
	}

	// finish and unfihish
	if parser.ContainsFlagStrict("f") {
		if prefix {
			finishRecordsByTags([]string{parser.GetFlagValueDefault("f", "")})
			return
		}
		toggle(true, getObjectIdByTags([]string{parser.GetFlagValueDefault("f", "")}), finish, parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("nf") {
		toggle(false, getObjectIdByTags([]string{parser.GetFlagValueDefault("nf", "")}), finish, parser.ContainsFlagStrict("prev"))
		return
	}

	// hold and unhold
	if parser.GetFlagValueDefault("hold", "") != "" {
		if prefix {
			holdRecordsByTags([]string{parser.GetFlagValueDefault("hold", "")})
			return
		}
		toggle(true, getObjectIdByTags([]string{parser.GetFlagValueDefault("hold", "")}), hold, parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("unhold") {
		toggle(false, getObjectIdByTags([]string{parser.GetFlagValueDefault("unhold", "")}), hold, parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("p") {
		toggle(true, getObjectIdByTags([]string{parser.GetFlagValueDefault("p", "")}), myproblem, parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("np") {
		toggle(false, getObjectIdByTags([]string{parser.GetFlagValueDefault("np", "")}), myproblem, parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.GetNumArgs() != -1 {
		n = int64(parser.GetNumArgs())
	}
	if parser.ContainsFlagStrict("n") {
		n = parser.MustGetFlagValAsInt64("n")
	}

	all := parser.ContainsFlagStrict("all") || (parser.ContainsFlag("a") &&
		!parser.ContainsFlagStrict("add-tag") && !parser.ContainsFlagStrict("del-tag") &&
		!parser.ContainsFlagStrict("tags")) && !parser.ContainsFlagStrict("binary")
	if all {
		n = math.MaxInt64
	}
	listSpecial = parser.ContainsFlagStrict("sp") || all
	reverse := parser.ContainsFlag("r") && !parser.ContainsAnyFlagStrict("prev", "remote", "prefix", "pre")
	includeFinished := parser.ContainsFlagStrict("include-finished") || all
	includeHeld := parser.ContainsFlagStrict("include-held") || all

	verbose := parser.ContainsFlagStrict("v")
	tags := []string{}
	listTagsAndOrderByTime := _helpers.OrderByTime(parser)
	if parser.ContainsFlagStrict("out") {
		txtOutputName, _ = parser.GetFlagVal("out")
		if txtOutputName == "" {
			txtOutputName = defaultTxtOutputName
		}
	}
	toBinary := parser.ContainsAnyFlagStrict("binary", "b")

	if (parser.ContainsFlagStrict("t") || parser.CoExists("t", "a")) && !listTagsAndOrderByTime {
		tags = strW.SplitNoEmpty(strings.TrimSpace(parser.GetMultiFlagValDefault([]string{"t", "ta", "at"}, "")), " ")
	}

	if parser.ContainsFlagStrict("clean-tag") {
		t := parser.GetFlagValueDefault("clean-tag", "")
		t = strings.ReplaceAll(t, ",", " ")
		tags = strW.SplitNoEmpty(t, " ")
		coloredTags := make([]string, len(tags))
		if len(tags) == 0 {
			fmt.Println("empty tags")
			return
		}
		for i := range tags {
			coloredTags[i] = color.HiRedString(tags[i])
		}
		fmt.Println("cleaning tags:", coloredTags)
		records, _ := listRecords(-1, reverse, true, true, tags, true, "", false, onlyHold, false)
		// fmt.Println("here", records)
		for _, record := range records {
			record.delete()
		}
		return
	}

	// list by tag name
	if (parser.ContainsFlagStrict("t") || parser.CoExists("t", "a")) && !listTagsAndOrderByTime {
		if parser.ContainsFlagStrict("pull") {
			remote.Set(true)
		}
		var records []*record
		// 如果是 id，特殊处理
		if _helpers.IsObjectID(parser.GetFlagValueDefault("t", "")) {
			id, err := primitive.ObjectIDFromHex(parser.GetFlagValueDefault("t", ""))
			if err != nil {
				panic(err)
			}
			r := &record{ID: id}
			r.loadByID()
			records = []*record{r}
		} else {
			records, _ = listRecords(n, reverse, includeFinished, includeHeld,
				tags, parser.ContainsFlagStrict("and"), "", parser.ContainsFlag("my") && !all, onlyHold,
				parser.ContainsAnyFlagStrict("prefix", "pre"))
		}
		if parser.ContainsFlagStrict("count") {
			fmt.Printf("%d records found\n", len(records))
			return
		}
		if !parser.ContainsAnyFlagStrict("pull", "push") {
			ignoreFields := []string{"AddDate", "ModifiedDate"}
			if verbose {
				ignoreFields = []string{}
			}
			// to stdout
			if !parser.ContainsFlagStrict("out") && !toBinary {
				for _, record := range records {
					printSeperator()
					coloringRecord(record, nil)
					if !utilsW.IsText([]byte(record.Title)) {
						record.Title = color.HiYellowString("<binary>")
					}
					fmt.Println(utilsW.ToString(record, ignoreFields...))
					fmt.Println(color.HiRedString(record.ID.String()))
				}
			} else if toBinary {
				for i := range records {
					content := records[i].Title
					idx := strings.IndexByte(content, '\n')
					filename := content[:idx]
					title := content[idx+1:]
					if !utilsW.IsExist(filename) ||
						(utilsW.IsExist(filename) && parser.ContainsFlagStrict("force")) ||
						(utilsW.IsExist(filename) && utilsW.PromptYesOrNo((fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", filename)))) {
						if err := os.WriteFile(filename, []byte(title), 0666); err != nil {
							panic(err)
						}
					}
				}
			} else { // to file
				var err error
				if (utilsW.IsExist(txtOutputName) && utilsW.PromptYesOrNo(fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", txtOutputName))) ||
					!utilsW.IsExist(txtOutputName) {
					buf := bytes.NewBufferString("")
					for _, r := range records {
						buf.WriteString(fmt.Sprintf("%s %v %s\n", strings.Repeat("-", 10), r.Tags, strings.Repeat("-", 10)))
						buf.WriteString(r.Title)
						buf.WriteString("\n")
					}
					if err = os.WriteFile(txtOutputName, buf.Bytes(), 0666); err != nil {
						panic(err)
					}
				}
			}
		} else {
			wg := sync.WaitGroup{}
			wg.Add(len(records))
			for _, r := range records {
				go func(r *record) {
					fmt.Printf("begin to sync %s...\n", r.ID.Hex())
					syncByID(r.ID.Hex(), parser.ContainsFlagStrict("push"), true)
					fmt.Println("finished syncing")
					wg.Done()
				}(r)
			}
			utilsW.TimeoutWait(&wg, 30*time.Second)
		}
		return
	}

	if parser.ContainsFlagStrict("u") || positional.Contains("u") {
		positional.Delete("u")
		var id string
		tags := positional.ToStringSlice()
		isObjectID := false
		if positional.Len() > 0 {
			isObjectID = _helpers.IsObjectID(tags[0])
		}
		// tags 里面可能是 objectid
		if len(tags) == 1 && isObjectID {
			id = tags[0]
			goto tagIsId
		}

		if len(tags) > 0 {
			if r, _ := listRecords(-1, true, false, false, tags, false, "", true, onlyHold, prefix); len(r) < 1 {
				fmt.Println(color.YellowString("no records associated with the tags (%v: prefix: %v) found", tags, prefix))
				return
			}
		}
		id = _helpers.ReadInfo(false)
	tagIsId:
		parser.Optional["-u"] = id
		if id != "" {
			parser.Optional["-e"] = ""
		}
		update(parser, parser.ContainsFlagStrict("file"), parser.ContainsFlagStrict("e"), id == "")
		return
	}

	if parser.ContainsFlagStrict("i") || parser.CoExists("i", "e") {
		insert(parser.CoExists("i", "e"), parser.GetFlagValueDefault("file", ""), "")
		return
	}

	if parser.ContainsFlagStrict("ct") || parser.CoExists("ct", "e") {
		changeTitle(parser.ContainsFlagStrict("file"),
			parser.CoExists("ct", "e"),
			parser.GetMultiFlagValDefault([]string{"ct", "cte", "ect"}, ""),
			parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("d") {
		deleteRecord(parser.GetFlagValueDefault("d", ""), parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("add-tag") {
		addTag(true, parser.GetFlagValueDefault("add-tag", ""), parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("del-tag") {
		addTag(false, parser.GetFlagValueDefault("del-tag", ""), parser.ContainsFlagStrict("prev"))
		return
	}

	if parser.ContainsFlagStrict("push") {
		fmt.Println("pushing...")
		syncByID(parser.GetFlagValueDefault("push", ""), true, true)
		return
	}

	if parser.ContainsFlagStrict("pull") {
		fmt.Println("pulling...")
		syncByID(parser.GetFlagValueDefault("pull", ""), false, true)
		return
	}
	// list tags, i stands for 'information'
	if listTagsAndOrderByTime || parser.ContainsFlagStrict("tags") || positional.Contains("tags") || positional.Contains("i") || positional.Contains("t") {
		all = parser.ContainsAnyFlagStrict("a", "all")
		var tags []tag
		var w int
		var err error
		buf := bytes.NewBufferString("")
		var cursor *mongo.Cursor
		var cli *mongo.Client
		var sortBy = "name"
		op1 := options.FindOptions{}
		var m bson.M = bson.M{}

		if all || listTagsAndOrderByTime {
			allRecords, _ := listRecords(-1, false, !listTagsAndOrderByTime || all, !listTagsAndOrderByTime || all,
				nil, false, "", false, onlyHold, false)

			// modified date map
			mtMap := getAllTagsModifiedDate(allRecords)
			testTags := containerW.NewOrderedMap()
			for _, r := range allRecords {
				for _, t := range r.Tags {
					testTags.Put(t, testTags.GetOrDefault(t, 0).(int)+1)
				}
			}
			for it := range testTags.Iterate() {
				v := it.Val().(int)
				t := tag{Name: it.Key().(string), Count: int64(v), modifiedDate: mtMap[it.Key().(string)]}
				// fmt.Println("here", it.Key().(string), mtMap[it.Key().(string)])
				tags = append(tags, t)
			}
			if listTagsAndOrderByTime {
				// sort.Sort(tagSlice(tags))
				sortW.Sort(tags, func(t1, t2 tag) int {
					if t1.modifiedDate.Before(t2.modifiedDate) {
						return -1
					}
					if t1.modifiedDate.Equal(t2.modifiedDate) {
						return 0
					}
					return 1
				})
			}
			// fmt.Println("tags", tags)
			goto print
		}
		if parser.GetNumArgs() != -1 {
			n = int64(parser.GetNumArgs())
		} else {
			n = 100
		}
		op1.SetLimit(n)
		if reverse {
			op1.SetSort(bson.M{sortBy: -1})
		} else {
			op1.SetSort(bson.M{sortBy: 1})
		}
		cli = client
		if remote.Get().(bool) {
			cli = atlasClient
		}
		if !listSpecial {
			m["name"] = bson.M{"$regex": primitive.Regex{Pattern: _helpers.BuildMongoRegularExpExclude(specialTagPatterns)}}
		}
		cursor, err = cli.Database(dbName).Collection(tagCollectionName).Find(ctx, m, &op1)
		if err != nil {
			panic(err)
		}
		cursor.All(ctx, &tags)
	print:
		_, w, err = utilsW.GetTerminalSize()
		// filter records
		if parser.GetFlagValueDefault("ex", "") != "" {
			tags = filterTags(tags, utilsW.GetCommandList(parser.MustGetFlagVal("ex")))
		}
		for _, tag := range tags {
			if verbose {
				tag.Name = color.HiGreenString(tag.Name)
				printSeperator()
				fmt.Println(utilsW.ToString(tag))
			} else {
				fmt.Fprintf(buf, `%s[%d]  `, tag.Name, tag.Count)
			}
		}
		if !verbose {
			if err == nil {
				terminalIndent := 2
				delimiter := "   "
				raw := strW.Wrap(buf.String(), w-terminalIndent, terminalIndent, delimiter)
				for _, line := range strW.SplitNoEmpty(raw, "\n") {
					arr := strW.SplitNoEmpty(line, " ")
					changedArr := make([]string, len(arr))
					for i := range arr {
						idx := strings.Index(arr[i], "[")
						if !isWindows {
							changedArr[i] = fmt.Sprintf("%s%s", color.HiGreenString(arr[i][:idx]), arr[i][idx:])
						} else { //windows color不能用
							changedArr[i] = arr[i]
						}
					}
					fmt.Fprintf(color.Output, "%s%s\n", strings.Repeat(" ", terminalIndent), strings.Join(changedArr, delimiter))
				}
			} else {
				panic(err)
			}
		}
		return
	}

	// list by title search
	if parser.ContainsFlagStrict("title") || parser.ContainsFlagStrict("c") {
		title := parser.GetFlagValueDefault("title", "")
		if title == "" {
			title = parser.GetFlagValueDefault("c", "")
		}
		records, _ := listRecords(n, reverse, includeFinished, includeHeld,
			tags, parser.ContainsFlagStrict("and"), title, parser.ContainsFlag("my") || all, onlyHold, prefix)

		if parser.ContainsFlagStrict("count") {
			fmt.Printf("%d records found\n", len(records))
			return
		}
		if !parser.ContainsFlagStrict("out") && !toBinary {
			for _, record := range records {
				printSeperator()
				p := regexp.MustCompile(`(?i)` + title)
				if !verbose {
					record.Title = "<hidden>"
				}
				coloringRecord(record, p)
				fmt.Println(record)
				fmt.Println(color.HiRedString(record.ID.String()))
			}
		} else if toBinary {
			panic("not supported")
		} else {
			var err error
			if (utilsW.IsExist(txtOutputName) && utilsW.PromptYesOrNo(fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", txtOutputName))) ||
				!utilsW.IsExist(txtOutputName) {
				buf := bytes.NewBufferString("")
				for _, r := range records {
					buf.WriteString(fmt.Sprintf("%s %v %s\n", strings.Repeat("-", 10), r.Tags, strings.Repeat("-", 10)))
					buf.WriteString(r.Title)
					buf.WriteString("\n")
				}
				if err = os.WriteFile(txtOutputName, buf.Bytes(), 0666); err != nil {
					panic(err)
				}
			}
		}
		return
	}
	if positional.Contains("open") {
		positional.Delete("open")
		listSpecial = true
		tags := positional.ToStringSlice()
		isObjectID := false
		if positional.Len() > 0 {
			isObjectID = _helpers.IsObjectID(tags[0])
		}
		// tags 里面可能是 objectid
		if len(tags) == 1 && isObjectID {
			objectID, _ := primitive.ObjectIDFromHex(tags[0])
			r := &record{ID: objectID}
			r.loadByID()
			_helpers.WriteInfo([]*primitive.ObjectID{&r.ID}, []string{r.Title})
		}
		if !isObjectID && len(tags) > 0 {
			if _, written := listRecords(-1, true, true, true, tags, false, "", false, onlyHold, prefix); !written {
				fmt.Printf("there are NO urls associated with tags: %v (prefix: %v)\n", tags, prefix)
				return
			}
		}

		_helpers.ReadInfo(true)
		return
	}

	// log everyday work
	if positional.Contains("log") {
		positional.Delete("log")
		nextDay := 0
		var err error
		if positional.Len() == 1 {
			if nextDay, err = strconv.Atoi(positional.ToStringSlice()[0]); err != nil {
				nextDay = 0
			}
		}

		tag := time.Now().Add(time.Duration(nextDay * int(time.Hour) * 24)).Format(fmt.Sprintf("%s.2006-01-02", "log"))
		rs, _ := listRecords(-1, true, true, true, []string{tag}, false, "", false, false, false)
		if len(rs) > 1 {
			panic("log failed: ")
		}
		if len(rs) == 0 {
			insert(true, "", tag)
		} else {
			parser.Optional["-u"] = rs[0].ID.Hex()
			update(parser, false, true, false)
		}
		return
	}

	// log week work
	if positional.Contains("week") {
		// merge from log.yyyy-MM-dd
		firstDay := utilsW.GetFirstDayOfThisWeek()
		now := time.Now()
		tag := firstDay.Format(fmt.Sprintf("%s.%s", "week", utilsW.DateFormat))
		rs, _ := listRecords(-1, true, true, true, []string{tag}, false, "", false, false, false)
		title := bytes.NewBufferString("")
		newWeekRecord := false
		if len(rs) > 1 {
			panic("too many week tags ")
		}
		if len(rs) == 0 {
			rs = []*record{newRecord("", tag)}
			newWeekRecord = true
		}
		for firstDay.Before(now) {
			dayTag := firstDay.Format(fmt.Sprintf("%s.%s", "log", utilsW.DateFormat))
			r, _ := listRecords(-1, true, true, true, []string{dayTag}, false, "", false, false, false)
			if len(r) > 1 {
				panic("log failed")
			}
			if len(r) == 1 {
				title.WriteString(fmt.Sprintf("-- %s --", firstDay.Format(utilsW.DateFormat)))
				title.WriteString("\n")
				title.WriteString(r[0].Title)
				title.WriteString("\n\n")
			}
			firstDay = firstDay.AddDate(0, 0, 1)
		}
		rs[0].Title = title.String()
		if newWeekRecord {
			rs[0].save(true)
		} else {
			rs[0].update(true)
		}
		return
	}

	// clean (move) images
	if positional.Contains("move") {
		s := positional.ToStringSlice()
		if len(s) != 3 {
			fmt.Println(">> re move absFileName type")
			return
		}
		type_, filename := s[2], s[1]
		logMsg := _helpers.LogMoveImages(type_, strings.ReplaceAll(filename, "\\\\", "\\"))
		tag := "move_" + type_
		rs, _ := listRecords(-1, true, true, true, []string{tag}, false, "", false, false, false)
		if len(rs) == 0 {
			newRecord(logMsg, tag).save(false)
		} else {
			s := containerW.NewOrderedSet()
			for _, title := range strings.Split(rs[0].Title, "\n") {
				s.Add(title)
			}
			for _, title := range strings.Split(logMsg, "\n") {
				s.Add(title)
			}
			rs[0].Title = strings.Join(s.ToStringSlice(), "\n")
			rs[0].update(false)
		}

	}
}
