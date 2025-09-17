package internal

import (
	"context"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

type Record struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty" json:"id,omitempty"`
	Tags         []string           `bson:"tags,ignoreempty" json:"tags,omitempty"`
	AddDate      time.Time          `bson:"add_date,ignoreempty" json:"add_date,omitempty"`
	ModifiedDate time.Time          `bson:"modified_date,ignoreempty" json:"modified_date,omitempty"`
	MyProblem    bool               `bson:"my_problem,ignoreempty" json:"my_problem,omitempty"`
	Finished     bool               `bson:"finished,ignoreempty" json:"finished,omitempty"`
	Hold         bool               `bson:"hold,ignoreempty" json:"hold,omitempty"`
	Title        string             `bson:"title,ignoreempty" json:"title,omitempty"`
}

func (r *Record) String() string {
	return utilsw.ToString(r, "AddDate", "ModifiedDate")
}

func NewRecord(title string, tags ...string) *Record {
	if len(tags) == 0 {
		tags = []string{autoTag}
	}
	r := &Record{Title: title, Tags: tags, Finished: false, MyProblem: true}
	t := time.Now()
	r.AddDate = t
	r.ModifiedDate = t
	r.ID = primitive.NewObjectID()
	return r
}

func (r *Record) Save(noUpdateModifiedDate bool) {
	if noUpdateModifiedDate {
		r.do("save", "noUpdateModifiedDate")
	} else {
		r.do("save")
	}
}

func (r *Record) Delete() {
	r.do("delete")
}

func (r *Record) DeleteByID() {
	r.do("deleteByID")
}

func (r *Record) Update(changeModifiedDate bool) {
	if changeModifiedDate {
		r.ModifiedDate = time.Now()
	}
	r.do("update")
}

func (r *Record) LoadByID() {
	r.do("load")
}

func (r *Record) do(action string, options ...string) {
	var err error
	var db *mongo.Database
	if !Remote.Get().(bool) {
		db = Client.Database(DbName)
	} else {
		db = AtlasClient.Database(DbName)
	}
	collection := db.Collection(CollectionName)
	session, err := Client.StartSession()
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
			s := cw.NewSet()
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
