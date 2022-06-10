package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"

	"github.com/grewwc/go_tools/src/executable/memo/_helpers"
	"github.com/grewwc/go_tools/src/stringsW"
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

	localMongoConfigName = "mongo.local"
	atlasMongoConfigName = "mongo.atlas"
)

const (
	autoTag              = "auto"
	defaultTxtOutputName = "output.txt"
	outputName           = "output_binary"
	finish               = "finish"
	myproblem            = "myproblem"
	titleLen             = 200
)

var (
	txtOutputName = "output.txt"
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

type record struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty" json:"id,ignoreempty"`
	Tags         []string           `bson:"tags,ignoreempty" json:"tags,ignoreempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty" json:"add_date,ignoreempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty" json:"modified_date,ignoreempty"`
	MyProblem    bool               `bson:"my_problem,ignoreempty" json:"my_problem,ignoreempty"`
	Finished     bool               `bson:"finished,ignoreempty" json:"finished,ignoreempty"`
	Title        string             `bson:"title,ignoreempty" json:"title,ignoreempty"`
}

type tag struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty"`
	Name         string             `bson:"name,ignoreempty"`
	Count        int64              `bson:"count,ignoreempty"`
	modifiedDate time.Time
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

	// check if tags and memo collections exists
	// db := client.Database(dbName)
	// if !_helpers.CollectionExists(db, ctx, tagCollectionName) {
	// 	db.Collection(tagCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
	// 		Keys:    bson.D{bson.DocElem{Name: "name", Value: "text"}}.Map(),
	// 		Options: options.Index().SetUnique(true),
	// 	})
	// }

	// if !helpers.CollectionExists(db, ctx, collectionName) {
	// 	db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
	// 		Keys:    bson.D{bson.DocElem{Name: "title", Value: "text"}}.Map(),
	// 		Options: options.Index().SetUnique(true),
	// 	})
	// }
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

func (r *record) do(action string) {
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
		r.ModifiedDate = time.Now()
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

func (r *record) save() {
	r.do("save")
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

func listRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool, title string,
	onlyMyproblem bool, prefix bool) ([]*record, bool) {
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
	if onlyMyproblem {
		m["my_problem"] = true
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

func update(parsed *terminalW.ParsedResults, fromFile bool, fromEditor bool, prev bool) {
	var err error
	var changed bool
	var cli = client
	if remote.Get().(bool) {
		cli = atlasClient
	}
	scanner := bufio.NewScanner(os.Stdin)
	id := parsed.GetFlagValueDefault("u", "")
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
		newRecord.Title = utilsW.InputWithEditor(oldTitle)
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
	fmt.Print("input the tags: ")
	scanner.Scan()
	tags := strings.TrimSpace(scanner.Text())
	tags = strings.ReplaceAll(tags, ",", " ")
	if tags != "" {
		changed = true
		newRecord.Tags = stringsW.SplitNoEmpty(tags, " ")
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

func insert(fromEditor bool, filename string) {
	scanner := bufio.NewScanner(os.Stdin)
	var title string
	var titleSlice []string
	var err error
	if filename != "" {
		title = strings.ReplaceAll(filename, ",", " ")
		titleSlice = stringsW.SplitNoEmpty(title, " ")
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
		title = utilsW.InputWithEditor("")
		fmt.Println()
	} else {
		fmt.Print("input the title: ")
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())
	}
	fmt.Print("input the tags: ")
	scanner.Scan()
	tagsStr := strings.TrimSpace(scanner.Text())
	tagsStr = strings.ReplaceAll(tagsStr, ",", " ")
	tags := stringsW.SplitNoEmpty(tagsStr, " ")
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
			r.save()
		}(c)
		<-c
		fmt.Println("Inserted: ")
		fmt.Println("\tTags:", r.Tags)
		fmt.Println("\tTitle:", stringsW.SubStringQuiet(r.Title, 0, titleLen))
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

func delete(id string, prev bool) {
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
		newTitle := utilsW.InputWithEditor(r.Title)
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
	fmt.Println("\tTitle:", stringsW.SubStringQuiet(r.Title, 0, titleLen))
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
	newTags := stringsW.SplitNoEmpty(strings.ReplaceAll(strings.TrimSpace(scanner.Text()), ",", " "), " ")

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
	fmt.Println("\tTitle:", stringsW.SubStringQuiet(r.Title, 0, titleLen))
}

func printSeperator() {
	fmt.Println(color.GreenString(strings.Repeat("~", 10)))
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
			all.WriteString(color.HiYellowString(string(bt[idx:i])))
			all.WriteString(color.RedString(string(bt[i:j])))
			idx = j
		}
		all.WriteString(color.HiYellowString(string(bt[idx:])))
		r.Title = all.String()
	} else {
		r.Title = color.HiYellowString(r.Title)
	}
	r.Title = "\n" + r.Title
	for i := range r.Tags {
		r.Tags[i] = color.HiBlueString(r.Tags[i])
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
		r.save()
	} else {
		r.update(false)
	}
	// 恢复remote
	remote = remoteBackUp
	n := 70
	if quiet {
		n = 20
	}
	fmt.Printf("finished %s %s: \n", msg, color.GreenString(stringsW.SubStringQuiet(r.Title, 0, n)))
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
		listRecords(-1, true, false, tags, false, "", true, false)
	}
	id := _helpers.ReadInfo(false)
	return id
}

type tagSlice []tag

func (l tagSlice) Len() int {
	return len([]tag(l))
}

func (l tagSlice) Swap(i, j int) {
	ll := []tag(l)
	ll[i], ll[j] = ll[j], ll[i]
}

func (l tagSlice) Less(i, j int) bool {
	ll := []tag(l)
	return ll[i].modifiedDate.Before(ll[j].modifiedDate)
}

func main() {
	defer func() {
		if res := recover(); res != nil {
			fmt.Println("recovered:", res)
		}
	}()
	var n int64 = 50

	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("i", false, "insert a record")
	fs.String("ct", "", "change a record title")
	fs.String("u", "", "update a record")
	fs.String("d", "", "delete a record")
	fs.Int("n", 10, "# of records to list")
	fs.Bool("h", false, "print help information")
	fs.String("push", "objectID to push", "push to remote db (may take a while)")
	fs.String("pull", "objectID to pull", "pull from remote db (may take a while)")
	fs.Bool("r", false, "reverse sort")
	fs.Bool("all", false, "including all record")
	fs.Bool("a", false, "shortcut for -all")
	fs.String("f", "", "finish a record")
	fs.String("nf", "", "set a record UNFINISHED")
	fs.String("p", "", "set a record my problem")
	fs.String("np", "", "set a record NOT my problem")
	fs.String("t", "", "search by tags")
	fs.Bool("include-finished", false, "include finished record")
	fs.String("add-tag", "", "add tags for a record")
	fs.String("del-tag", "", "delete tags for a record")
	fs.String("clean-tag", "", "clean all the records having the tag")
	fs.Bool("tags", false, "list all tags")
	fs.Bool("and", false, "use and logic to match tags")
	fs.Bool("v", false, "verbose (show modify/add time, verbose)")
	fs.String("file", "", "read title from a file, for '-u' & '-ct', file serve as bool, for '-i', needs to pass filename")
	fs.Bool("e", false, "read from editor")
	fs.String("title", "", "search by title")
	fs.String("c", "", "content (alias for title)")
	fs.String("out", "", fmt.Sprintf("output to text file (default is %s)", defaultTxtOutputName))
	fs.Bool("my", false, "only list my problem")
	fs.Bool("remote", false, "operate on the remote server")
	fs.Bool("prev", false, "operate based on the previous ObjectIDs")
	fs.Bool("count", false, "only print the count, not the result")
	fs.Bool("prefix", false, "tag prefix")
	fs.Bool("binary", false, "if the title is binary file")
	fs.Bool("b", false, "shortcut for -binary")
	fs.Bool("force", false, "force overwrite")

	parsed := terminalW.ParseArgsCmd("h", "r", "all", "a",
		"i", "include-finished", "tags", "and", "v", "e", "my", "remote", "prev", "count", "prefix", "binary", "b")

	if parsed == nil {
		records, _ := listRecords(n, false, false, []string{"todo", "urgent"}, false, "", true, true)
		for _, record := range records {
			printSeperator()
			coloringRecord(record, nil)
			fmt.Println(record)
			fmt.Println(color.HiRedString(record.ID.String()))
		}
		return
	}
	positional := parsed.Positional
	prefix := parsed.ContainsFlagStrict("prefix")
	isWindows := utilsW.WINDOWS == utilsW.GetPlatform()

	if parsed.ContainsFlagStrict("remote") {
		initAtlas()
		remote.Set(true)
	}

	if parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	if parsed.ContainsFlagStrict("f") {
		toggle(true, getObjectIdByTags([]string{parsed.GetFlagValueDefault("f", "")}), finish, parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("nf") {
		toggle(false, getObjectIdByTags([]string{parsed.GetFlagValueDefault("nf", "")}), finish, parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("p") {
		toggle(true, getObjectIdByTags([]string{parsed.GetFlagValueDefault("p", "")}), myproblem, parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("np") {
		toggle(false, getObjectIdByTags([]string{parsed.GetFlagValueDefault("np", "")}), myproblem, parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.GetNumArgs() != -1 {
		n = int64(parsed.GetNumArgs())
	}
	if parsed.ContainsFlagStrict("n") {
		n = parsed.MustGetFlagValAsInt64("n")
	}

	all := parsed.ContainsFlagStrict("all") || (parsed.ContainsFlag("a") &&
		!parsed.ContainsFlagStrict("add-tag") && !parsed.ContainsFlagStrict("del-tag") &&
		!parsed.ContainsFlagStrict("tags")) && !parsed.ContainsFlagStrict("binary")
	if all {
		n = math.MaxInt64
	}
	reverse := parsed.ContainsFlag("r") && !parsed.ContainsAnyFlagStrict("prev", "remote", "prefix")
	includeFinished := parsed.ContainsFlagStrict("include-finished") || all
	verbose := parsed.ContainsFlagStrict("v")
	tags := []string{}
	listTagsAndOrderByTime := _helpers.OrderByTime(parsed)
	if parsed.ContainsFlagStrict("out") {
		txtOutputName, _ = parsed.GetFlagVal("out")
		if txtOutputName == "" {
			txtOutputName = defaultTxtOutputName
		}
	}
	toBinary := parsed.ContainsAnyFlagStrict("binary", "b")

	if (parsed.ContainsAnyFlagStrict("t") || parsed.CoExists("t", "a")) && !listTagsAndOrderByTime {
		tags = stringsW.SplitNoEmpty(strings.TrimSpace(parsed.GetMultiFlagValDefault([]string{"t", "ta", "at"}, "")), " ")
	}

	if parsed.ContainsFlagStrict("clean-tag") {
		t := parsed.GetFlagValueDefault("clean-tag", "")
		t = strings.ReplaceAll(t, ",", " ")
		tags = stringsW.SplitNoEmpty(t, " ")
		coloredTags := make([]string, len(tags))
		if len(tags) == 0 {
			fmt.Println("empty tags")
			return
		}
		for i := range tags {
			coloredTags[i] = color.HiRedString(tags[i])
		}
		fmt.Println("cleaning tags:", coloredTags)
		records, _ := listRecords(-1, reverse, true, tags, true, "", false, false)
		// fmt.Println("here", records)
		for _, record := range records {
			record.delete()
		}
		return
	}

	// list by tag name
	if (parsed.ContainsAnyFlagStrict("t") || parsed.CoExists("t", "a")) && !listTagsAndOrderByTime {
		if parsed.ContainsFlagStrict("pull") {
			remote.Set(true)
		}
		records, _ := listRecords(n, reverse, includeFinished || all, tags, parsed.ContainsFlagStrict("and"), "", parsed.ContainsFlag("my") && !all, parsed.ContainsFlagStrict("prefix"))
		if parsed.ContainsFlagStrict("count") {
			fmt.Printf("%d records found\n", len(records))
			return
		}
		if !parsed.ContainsAnyFlagStrict("pull", "push") {
			ignoreFields := []string{"AddDate", "ModifiedDate"}
			if verbose {
				ignoreFields = []string{}
			}
			if !parsed.ContainsFlagStrict("out") && !toBinary {
				for _, record := range records {
					printSeperator()
					coloringRecord(record, nil)
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
						(utilsW.IsExist(filename) && parsed.ContainsFlagStrict("force")) ||
						(utilsW.IsExist(filename) && utilsW.PromptYesOrNo((fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", filename)))) {
						if err := ioutil.WriteFile(filename, []byte(title), 0666); err != nil {
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
					if err = ioutil.WriteFile(txtOutputName, buf.Bytes(), 0666); err != nil {
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
					syncByID(r.ID.Hex(), parsed.ContainsFlagStrict("push"), true)
					fmt.Println("finished syncing")
					wg.Done()
				}(r)
			}
			utilsW.TimeoutWait(&wg, 30*time.Second)
		}
		return
	}

	if parsed.ContainsFlagStrict("u") || positional.Contains("u") {
		positional.Delete("u")
		tags := positional.ToStringSlice()
		var id string
		if len(tags) > 0 {
			if r, _ := listRecords(-1, true, false, tags, false, "", true, prefix); len(r) < 1 {
				fmt.Println(color.YellowString("no records associated with the tags (%v: prefix: %v) found", tags, prefix))
				return
			}
		}
		id = _helpers.ReadInfo(false)
		parsed.Optional["-u"] = id
		if id != "" {
			parsed.Optional["-e"] = ""
		}
		update(parsed, parsed.ContainsFlagStrict("file"), parsed.ContainsFlagStrict("e"), id == "")
		return
	}

	if parsed.ContainsFlagStrict("i") || parsed.CoExists("i", "e") {
		insert(parsed.CoExists("i", "e"), parsed.GetFlagValueDefault("file", ""))
		return
	}

	if parsed.ContainsFlagStrict("ct") || parsed.CoExists("ct", "e") {
		changeTitle(parsed.ContainsFlagStrict("file"),
			parsed.CoExists("ct", "e"),
			parsed.GetMultiFlagValDefault([]string{"ct", "cte", "ect"}, ""),
			parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("d") {
		delete(parsed.GetFlagValueDefault("d", ""), parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("add-tag") {
		addTag(true, parsed.GetFlagValueDefault("add-tag", ""), parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("del-tag") {
		addTag(false, parsed.GetFlagValueDefault("del-tag", ""), parsed.ContainsFlagStrict("prev"))
		return
	}

	if parsed.ContainsFlagStrict("push") {
		fmt.Println("pushing...")
		syncByID(parsed.GetFlagValueDefault("push", ""), true, true)
		return
	}

	if parsed.ContainsFlagStrict("pull") {
		fmt.Println("pulling...")
		syncByID(parsed.GetFlagValueDefault("pull", ""), false, true)
		return
	}
	// list tags, i stands for 'information'
	if listTagsAndOrderByTime || parsed.ContainsFlagStrict("tags") || positional.Contains("tags") || positional.Contains("i") || positional.Contains("t") {
		all = parsed.ContainsAnyFlagStrict("a", "all")
		var tags []tag
		var w int
		var err error
		buf := bytes.NewBufferString("")
		var cursor *mongo.Cursor
		var cli *mongo.Client
		var sortBy = "name"
		op1 := options.FindOptions{}

		if all || listTagsAndOrderByTime {
			allRecords, _ := listRecords(-1, false, !listTagsAndOrderByTime || all, nil, false, "", false, false)
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
				sort.Sort(tagSlice(tags))
			}
			// fmt.Println("tags", tags)
			goto print
		}
		if parsed.GetNumArgs() != -1 {
			n = int64(parsed.GetNumArgs())
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
		cursor, err = cli.Database(dbName).Collection(tagCollectionName).Find(ctx, bson.M{}, &op1)
		if err != nil {
			panic(err)
		}
		cursor.All(ctx, &tags)
	print:
		_, w, err = utilsW.GetTerminalSize()
		for _, tag := range tags {
			if verbose {
				tag.Name = color.HiBlueString(tag.Name)
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
				raw := stringsW.Wrap(buf.String(), w-terminalIndent, terminalIndent, delimiter)
				for _, line := range stringsW.SplitNoEmpty(raw, "\n") {
					arr := stringsW.SplitNoEmpty(line, " ")
					changedArr := make([]string, len(arr))
					for i := range arr {
						idx := strings.Index(arr[i], "[")
						if !isWindows {
							changedArr[i] = fmt.Sprintf("%s%s", color.HiBlueString(arr[i][:idx]), arr[i][idx:])
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

	// list by title
	if parsed.ContainsFlagStrict("title") || parsed.ContainsFlagStrict("c") {
		title := parsed.GetFlagValueDefault("title", "")
		if title == "" {
			title = parsed.GetFlagValueDefault("c", "")
		}
		records, _ := listRecords(n, reverse, includeFinished || all,
			tags, parsed.ContainsFlagStrict("and"), title, parsed.ContainsFlag("my") || all,
			parsed.ContainsFlagStrict("prefix"))

		if parsed.ContainsFlagStrict("count") {
			fmt.Printf("%d records found\n", len(records))
			return
		}
		if !parsed.ContainsFlagStrict("out") && !toBinary {
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
				if err = ioutil.WriteFile(txtOutputName, buf.Bytes(), 0666); err != nil {
					panic(err)
				}
			}
		}
		return
	}
	if positional.Contains("open") {
		positional.Delete("open")
		tags := positional.ToStringSlice()
		isObjectID := _helpers.IsObjectID(tags[0])
		// tags 里面可能是 objectid
		if len(tags) == 1 && isObjectID {
			objectID, _ := primitive.ObjectIDFromHex(tags[0])
			r := &record{ID: objectID}
			r.loadByID()
			_helpers.WriteInfo([]*primitive.ObjectID{&r.ID}, []string{r.Title})
		}
		if !isObjectID && len(tags) > 0 {
			if _, written := listRecords(-1, true, false, tags, false, "", true, prefix); !written {
				fmt.Printf("there are NO urls associated with tags: %v (prefix: %v)\n", tags, prefix)
				return
			}
		}

		_helpers.ReadInfo(true)
		return
	}
}
