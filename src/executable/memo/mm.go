package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/executable/memo/helpers"
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
)

const (
	autoTag = "auto"
)

var (
	uri           string
	clientOptions = &options.ClientOptions{}
	ctx           context.Context
	client        *mongo.Client
)

type record struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty"`
	Tags         []string           `bson:"tags,ignoreempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty"`
	Finished     bool               `bson:"finished,ignoreempty"`
	Title        string             `bson:"title,ignoreempty"`
}

type tag struct {
	ID    primitive.ObjectID `bson:"_id,ignoreempty"`
	Name  string             `bson:"name,ignoreempty"`
	Count int64              `bson:"count,ignoreempty"`
}

func newRecord(title string, tags ...string) *record {
	r := &record{Title: title, Tags: tags, Finished: false}
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
		log.Fatalln(err)
	}

	// check if tags and memo collections exists
	db := client.Database(dbName)
	if !helpers.CollectionExists(db, ctx, tagCollectionName) {
		db.Collection(tagCollectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.DocElem{Name: "name", Value: "text"}}.Map(),
			Options: options.Index().SetUnique(true),
		})
	}

	if !helpers.CollectionExists(db, ctx, collectionName) {
		db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{bson.DocElem{Name: "title", Value: "text"}}.Map(),
			Options: options.Index().SetUnique(true),
		})
	}
}

func (r record) String() string {
	return utilsW.ToString(r, "AddDate", "ModifiedDate")
}

func incrementTagCount(db *mongo.Database, tags []string, val int) {
	_, err := db.Collection(tagCollectionName).UpdateMany(ctx, bson.M{"name": bson.M{"$in": tags}},
		bson.M{"$inc": bson.M{"count": val}}, options.Update().SetUpsert(true))
	if err != nil {
		panic(err)
	}

	if _, err = db.Collection(tagCollectionName).DeleteMany(ctx, bson.M{"count": bson.M{"$lt": 1}}); err != nil {
		panic(err)
	}
}

func (r *record) exists() bool {
	collection := client.Database(dbName).Collection(collectionName)
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
	db := client.Database(dbName)
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
	r.ModifiedDate = time.Now()
	r.do("save")
}

func (r *record) delete() {
	r.do("delete")
}

func (r *record) deleteByID() {
	r.do("deleteByID")
}

func (r *record) update() {
	r.ModifiedDate = time.Now()
	r.do("update")
}

func (r *record) loadByID() {
	r.do("load")
}

func listRecords(limit int64, reverse, includeFinished bool, tags []string, useAnd bool) []*record {
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
	collection := client.Database(dbName).Collection(collectionName)
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
			m["tags"] = bson.M{"$in": tags}
		}
	}
	cursor, err := collection.Find(ctx, m, addDateOption, modifiedDataOption)
	if err != nil {
		panic(err)
	}
	var res []*record
	if err = cursor.All(ctx, &res); err != nil {
		panic(err)
	}
	return res
}

func update(parsed *terminalW.ParsedResults, fromFile bool, fromEditor bool) {
	var err error
	var changed bool
	scanner := bufio.NewScanner(os.Stdin)
	id := parsed.GetFlagValueDefault("u", "")
	newRecord := record{}
	if id == "" {
		fmt.Print("input the Object ID: ")
		scanner.Scan()
		id = scanner.Text()
	}
	if newRecord.ID, err = primitive.ObjectIDFromHex(id); err != nil {
		panic(err)
	}

	newRecord.loadByID()
	oldTitle := newRecord.Title
	oldTags := newRecord.Tags

	fmt.Print("input the title: ")
	scanner.Scan()
	var title string
	if fromEditor {
		title = utilsW.InputWithEditor()
		fmt.Println()
	} else {
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
	if tags != "" {
		changed = true
		newRecord.Tags = stringsW.SplitNoEmpty(tags, " ")
		c := make(chan interface{}, 1)
		go func(c chan interface{}) {
			incrementTagCount(client.Database(dbName), oldTags, -1)
			c <- nil
		}(c)
		go func(c chan interface{}) {
			incrementTagCount(client.Database(dbName), newRecord.Tags, 1)
			c <- nil
		}(c)
		<-c
		<-c
	}
	if !changed {
		return
	}
	fmt.Println("==> Changes: ")
	fmt.Printf("title: %q -> %q\n", oldTitle, newRecord.Title)
	fmt.Printf("tags: %q -> %q\n", oldTags, newRecord.Tags)
	fmt.Print("Do you want to update the record? (y/n): ")
	scanner.Scan()
	ans := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if ans == "y" {
		newRecord.update()
	} else {
		fmt.Println("Abort change")
	}
}

func insert(fromFile, fromEditor bool) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the title: ")
	var title string
	if fromEditor {
		title = utilsW.InputWithEditor()
		fmt.Println()
	} else {
		scanner.Scan()
		title = strings.TrimSpace(scanner.Text())
		if fromFile {
			title = utilsW.ReadString(title)
		}
	}
	fmt.Print("input the tags: ")
	scanner.Scan()
	tags := stringsW.SplitNoEmpty(strings.TrimSpace(scanner.Text()), " ")
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
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
	fmt.Println(r)
}

func setFinish(finish bool) {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the Object ID: ")
	scanner.Scan()
	r := record{}
	if r.ID, err = primitive.ObjectIDFromHex(strings.TrimSpace(scanner.Text())); err != nil {
		panic(err)
	}
	r.loadByID()
	r.Finished = finish
	r.update()
}

func delete() {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the Object ID: ")
	scanner.Scan()
	r := record{}
	if r.ID, err = primitive.ObjectIDFromHex(strings.TrimSpace(scanner.Text())); err != nil {
		panic(err)
	}
	r.loadByID()
	r.delete()
}

func changeTitle(fromFile, fromEditor bool) {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the Object ID: ")
	scanner.Scan()
	c := make(chan interface{})
	r := record{}
	if r.ID, err = primitive.ObjectIDFromHex(strings.TrimSpace(scanner.Text())); err != nil {
		panic(err)
	}
	go func(chan interface{}) {
		c <- nil
		r.loadByID()
	}(c)
	<-c
	fmt.Print("input the New Title: ")
	if fromEditor {
		r.Title = utilsW.InputWithEditor()
		fmt.Println()
	} else {
		scanner.Scan()
		r.Title = strings.TrimSpace(scanner.Text())
		if fromFile {
			r.Title = utilsW.ReadString(r.Title)
		}
	}
	go func(c chan interface{}) {
		r.update()
		c <- nil
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println(r)
}

func addTag(add bool) {
	var err error
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the Object ID: ")
	scanner.Scan()
	r := record{}
	if r.ID, err = primitive.ObjectIDFromHex(strings.TrimSpace(scanner.Text())); err != nil {
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
	newTags := stringsW.SplitNoEmpty(strings.TrimSpace(scanner.Text()), " ")

	s := (<-c).(*containerW.Set)
	newTagSet := containerW.NewSet()
	for _, newTag := range newTags {
		if !s.Contains(newTag) {
			newTagSet.Add(newTag)
		}
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
		incrementTagCount(client.Database(dbName), newTagSet.ToStringSlice(), incVal)
	}(c)
	<-c
	go func(c chan interface{}) {
		defer func() {
			c <- nil
		}()
		r.Tags = s.ToStringSlice()
		r.update()
	}(c)
	<-c
	fmt.Println("New Record: ")
	fmt.Println(r)
}

func main() {
	defer func() {
		if res := recover(); res != nil {
			fmt.Println(res)
		}
	}()
	var n int64 = 3

	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("i", false, "insert a record")
	fs.Bool("ct", false, "change a record")
	fs.Bool("u", false, "update a record")
	fs.Bool("d", false, "delete a record")
	fs.Bool("l", false, "list records")
	fs.Int("n", 3, "# of records to list")
	fs.Bool("h", false, "print help information")
	fs.Bool("sync", false, "sync to remote db (may take a while)")
	fs.Bool("r", false, "if true, newer first")
	fs.Bool("all", false, "including all record")
	fs.Bool("a", false, "shortcut for -all")
	fs.Bool("f", false, "finish a record")
	fs.Bool("nf", false, "set a record UNFINISHED")
	fs.String("t", "", "search by tags")
	fs.Bool("include-finished", false, "include finished record")
	fs.Bool("add-tag", false, "add tags")
	fs.Bool("del-tag", false, "delete tags")
	fs.Bool("tags", false, "list all tags")
	fs.Bool("and", false, "use and logic to match tags")
	fs.Bool("v", false, "verbose (show modify/add time)")
	fs.Bool("file", false, "read title from a file")
	fs.Bool("e", false, "read from editor")

	parsed := terminalW.ParseArgsCmd("l", "h", "sync", "r", "all", "f", "a",
		"ct", "i", "u", "d", "include-finished", "add-tag", "del-tag", "tags", "and", "v", "file", "e")

	if parsed == nil {
		records := listRecords(n, false, false, []string{"todo", "urgent"}, false)
		for _, record := range records {
			fmt.Println(color.GreenString(strings.Repeat("~", 10)))
			fmt.Println(record)
		}
		return
	}

	if parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	if parsed.ContainsFlagStrict("f") {
		setFinish(true)
		return
	}

	if parsed.GetNumArgs() != -1 {
		n = int64(parsed.GetNumArgs())
	}
	all := parsed.ContainsFlagStrict("all") || (parsed.ContainsFlag("a") &&
		!parsed.ContainsFlagStrict("add-tag") && !parsed.ContainsFlagStrict("del-tag"))
	if all {
		n = math.MaxInt64
	}
	reverse := parsed.ContainsFlag("r")
	includeFinished := parsed.ContainsFlagStrict("include-finished") || all

	if parsed.ContainsFlag("l") &&
		!parsed.ContainsFlagStrict("include-finished") &&
		!parsed.ContainsFlagStrict("del-tag") &&
		!parsed.ContainsFlagStrict("file") {
		var tags []string = nil
		if parsed.ContainsFlagStrict("t") {
			tags = stringsW.SplitNoEmpty(strings.TrimSpace(parsed.GetFlagValueDefault("t", "")), " ")
		}
		records := listRecords(n, reverse, includeFinished, tags, parsed.ContainsFlagStrict("and"))
		ignoreFields := []string{"AddDate", "ModifiedDate"}
		if parsed.ContainsFlag("v") {
			ignoreFields = []string{}
		}
		for _, record := range records {
			fmt.Println(utilsW.ToString(record, ignoreFields...))
		}
	}

	if parsed.ContainsFlagStrict("u") {
		update(parsed, parsed.ContainsFlagStrict("file"), parsed.ContainsFlagStrict("e"))
		return
	}

	if parsed.ContainsFlagStrict("i") {
		insert(parsed.ContainsFlagStrict("file"), parsed.ContainsFlagStrict("e"))
		return
	}

	if parsed.ContainsFlagStrict("ct") {
		changeTitle(parsed.ContainsFlagStrict("file"), parsed.ContainsFlagStrict("e"))
		return
	}

	if parsed.ContainsFlagStrict("d") {
		delete()
		return
	}

	if parsed.ContainsFlagStrict("add-tag") {
		addTag(true)
		return
	}

	if parsed.ContainsFlagStrict("del-tag") {
		addTag(false)
		return
	}

	if parsed.ContainsFlagStrict("sync") {
		fmt.Println("unsupported!!")
		return
	}

	if parsed.ContainsFlagStrict("nf") {
		setFinish(false)
		return
	}

	if parsed.ContainsFlagStrict("tags") {
		cursor, err := client.Database(dbName).Collection(tagCollectionName).Find(ctx, bson.M{})
		if err != nil {
			panic(err)
		}
		var tags []tag
		cursor.All(ctx, &tags)
		for _, tag := range tags {
			fmt.Println(utilsW.ToString(tag))
		}
	}
	return
}