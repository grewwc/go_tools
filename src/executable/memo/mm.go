package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

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
	Tag          string             `bson:"tag,ignoreempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty"`
	Finished     bool               `bson:"finished,ignoreempty"`
}

func newRecord(title, tag string) *record {
	r := &record{Title: title, Tag: tag, Finished: false}
	t := time.Now()
	r.AddDate = t
	r.ModifiedDate = t
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
	clientOptions.SetMaxPoolSize(10)
	// ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalln(err)
	}
}

func (r *record) exists() bool {
	collection := client.Database(dbName).Collection(collectionName)
	singleResults := collection.FindOne(context.Background(), bson.M{"title": r.Title, "tag": r.Tag})
	err := singleResults.Err()
	if err == nil {
		return true
	}

	if err == mongo.ErrNoDocuments {
		return false
	}
	panic(err)
}

func (r *record) save() {
	r.do("save")
}

func (r *record) do(action string, data ...string) {
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
	case "save":
		if r.exists() {
			return
		}
		if _, err = collection.InsertOne(context.Background(), r); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
	case "delete":
		if _, err = collection.DeleteOne(ctx, bson.M{"title": r.Title, "tag": r.Tag}); err != nil {
			session.AbortTransaction(ctx)
			panic(err)
		}
	case "update":
		var m bson.M
		switch len(data) {
		case 1:
			m = bson.M{"$set": bson.M{"title": data[0]}}
		case 2:
			m = bson.M{"$set": bson.M{"title": data[0], "tag": data[1]}}
		default:
			panic("title + tag (too many data)")
		}
		if _, err = collection.UpdateOne(ctx, bson.M{"title": r.Title, "tag": r.Tag}, m); err != nil {
			panic(err)
		}

	default:
		panic("unknow action " + action)
	}

	if err = session.CommitTransaction(ctx); err != nil {
		panic(err)
	}
}

func (r *record) delete() {
	r.do("delete")
}

func (r *record) update(title, tag string) {
	r.do("update", title, tag)
}

func listRecords(limit int64) []record {
	if limit < 0 {
		limit = math.MaxInt64
	}
	collection := client.Database(dbName).Collection(collectionName)
	findOpts := options.Find()
	findOpts.SetLimit(limit)
	findOpts.SetSort(bson.D{{"modified_date", -1}}.Map())
	cursor, err := collection.Find(ctx, bson.D{}.Map(), findOpts)
	if err != nil {
		panic(err)
	}
	var res []record
	if err = cursor.All(ctx, &res); err != nil {
		panic(err)
	}
	fmt.Println(res)
	return nil
}

func main() {
	// r := newRecord("hello worlds", "test")
	listRecords(-1)
}
