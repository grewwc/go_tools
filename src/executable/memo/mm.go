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

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	dbName         = "daily"
	collectionName = "memo"

	localMongoConfigName = "mongo.local"
	mongoDbAtlas         = "mongodb://wwc129:!Grewwc080959@cluster0.myh9q.mongodb.net/daily?retryWrites=true&w=majority"
)

const (
	autoTag = "auto"
)

var (
	remoteURI = fmt.Sprintf("mongodb://wwc129:!Grewwc080959@cluster0.myh9q.mongodb.net/%s?retryWrites=true&w=majority", dbName)
)

var (
	uri           string = remoteURI
	clientOptions        = &options.ClientOptions{}
	ctx           context.Context
	client        *mongo.Client
)

type record struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty"`
	Title        string             `bson:"title,ignoreempty"`
	Tags         []string           `bson:"tags,ignoreempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty"`
	Finished     bool               `bson:"finished,ignoreempty"`
}

type tag struct {
	ID   primitive.ObjectID `bson:"_id,ignoreempty"`
	Name string             `bson:"name,ignoreempty"`
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
}

func (r record) String() string {
	return utilsW.ToString(r, "AddDate", "ModifiedDate")
}

func (r *record) exists() bool {
	collection := client.Database(dbName).Collection(collectionName)
	singleResults := collection.FindOne(context.Background(), bson.M{"title": r.Title, "tags": r.Tags})
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
	case "delete":
		if _, err = collection.DeleteOne(ctx, bson.M{"title": r.Title, "tags": r.Tags}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
	case "deleteByID":
		if _, err = collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
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

func listRecords(limit int64, reverse, includeFinished bool, tags []string) []*record {
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
		m["tags"] = bson.M{"$all": tags}
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

func update(parsed *terminalW.ParsedResults) {
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
	title := strings.TrimSpace(scanner.Text())
	if title != "" {
		changed = true
		newRecord.Title = title
	}
	fmt.Print("input the tags: ")
	scanner.Scan()
	tags := strings.TrimSpace(scanner.Text())
	if tags != "" {
		changed = true
		newRecord.Tags = stringsW.SplitNoEmpty(tags, " ")
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

func create() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("input the title: ")
	scanner.Scan()
	title := strings.TrimSpace(scanner.Text())
	fmt.Print("input the tags: ")
	scanner.Scan()
	tags := stringsW.SplitNoEmpty(strings.TrimSpace(scanner.Text()), " ")
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
	r := newRecord(title, tags...)
	r.save()
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

func addTag() {
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

	fmt.Print("input the New Tag: ")
	scanner.Scan()
	newTags := stringsW.SplitNoEmpty(strings.TrimSpace(scanner.Text()), " ")

	s := (<-c).(*containerW.Set)
	for _, newTag := range newTags {
		s.Add(newTag)
	}
	r.Tags = s.ToStringSlice()
	r.update()
}

func main() {
	defer func() {
		if res := recover(); res != nil {
			fmt.Println(res)
		}
	}()

	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("c", false, "create a record")
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
	fs.Bool("-include-finished", false, "include finished record")
	fs.Bool("-add-tag", false, "add tag")

	parsed := terminalW.ParseArgsCmd("l", "h", "sync", "r", "all", "f", "a",
		"c", "u", "d", "-include-finished", "-add-tag")

	if parsed == nil || parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	if parsed.ContainsFlagStrict("f") {
		setFinish(true)
	}
	var n int64 = 3
	if parsed.GetNumArgs() != -1 {
		n = int64(parsed.GetNumArgs())
	}
	all := parsed.ContainsFlagStrict("all") || (parsed.ContainsFlag("a") && !parsed.ContainsFlagStrict("--add-tags"))
	if all {
		n = math.MaxInt64
	}
	reverse := parsed.ContainsFlag("r")
	includeFinished := parsed.ContainsFlagStrict("--include-finished") || all

	if parsed.ContainsFlag("l") && !parsed.ContainsFlagStrict("--include-finished") {
		records := listRecords(n, reverse, includeFinished, nil)
		for _, record := range records {
			fmt.Println(record)
		}
	}

	if parsed.ContainsFlagStrict("u") {
		update(parsed)
	}

	if parsed.ContainsFlagStrict("c") {
		create()
	}

	if parsed.ContainsFlagStrict("d") {
		delete()
	}

	if parsed.ContainsFlagStrict("t") {
		tags := stringsW.SplitNoEmpty(strings.TrimSpace(parsed.GetFlagValueDefault("t", "")), " ")
		records := listRecords(n, reverse, includeFinished, tags)
		for _, record := range records {
			fmt.Println(record)
		}
	}

	if parsed.ContainsFlagStrict("--add-tag") {
		addTag()
	}

	if parsed.ContainsFlagStrict("sync") {
		fmt.Println("unsupported!!")
		return
	}

	if parsed.ContainsFlagStrict("nf") {
		setFinish(false)
	}
}
