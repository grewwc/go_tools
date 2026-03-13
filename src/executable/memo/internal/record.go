package internal

import (
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

	Invalid bool
}

func (r *Record) String() string {
	return utilsw.ToString(r, "AddDate", "ModifiedDate", "Invalid")
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
	if Remote.Get().(bool) {
		InitRemote()
		collection := AtlasClient.Database(DbName).Collection(CollectionName)
		switch action {
		case "load":
			if err := collection.FindOne(ctx, bson.M{"_id": r.ID}).Decode(r); err != nil {
				if err != mongo.ErrNoDocuments {
					panic(err)
				} else {
					r.Invalid = true
				}
			}

		case "save":
			if r.exists() {
				return
			}
			if !noUpdateModifiedDate {
				r.ModifiedDate = time.Now()
			}
			if _, err := collection.InsertOne(ctx, r); err != nil {
				panic(err)
			}
			incrementTagCount(r.Tags, 1)
		case "delete":
			if _, err := collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
				panic(err)
			}
			incrementTagCount(r.Tags, -1)
		case "deleteByID":
			if _, err := collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
				panic(err)
			}
			incrementTagCount(r.Tags, -1)
		case "update":
			if _, err := collection.UpdateOne(ctx, bson.M{"_id": r.ID}, bson.M{"$set": r}); err != nil {
				panic(err)
			}

		default:
			panic("unknow action " + action)
		}
		return
	}
	if useLocalSQLite() {
		switch action {
		case "load":
			loaded, err := sqliteLoadRecord(r.ID)
			if err != nil {
				panic(err)
			}
			if loaded == nil {
				r.Invalid = true
				return
			}
			*r = *loaded
		case "save":
			if err := sqliteSaveRecord(r, noUpdateModifiedDate); err != nil {
				panic(err)
			}
		case "delete", "deleteByID":
			if err := sqliteDeleteRecord(r); err != nil {
				panic(err)
			}
		case "update":
			if err := sqliteUpdateRecord(r); err != nil {
				panic(err)
			}
		default:
			panic("unknow action " + action)
		}
		return
	}
	collection := Client.Database(DbName).Collection(CollectionName)
	switch action {
	case "load":
		if err := collection.FindOne(ctx, bson.M{"_id": r.ID}).Decode(r); err != nil {
			if err != mongo.ErrNoDocuments {
				panic(err)
			} else {
				r.Invalid = true
			}
		}

	case "save":
		if r.exists() {
			return
		}
		if !noUpdateModifiedDate {
			r.ModifiedDate = time.Now()
		}
		if _, err := collection.InsertOne(ctx, r); err != nil {
			panic(err)
		}
		incrementTagCount(r.Tags, 1)
	case "delete":
		if _, err := collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
			panic(err)
		}
		incrementTagCount(r.Tags, -1)
	case "deleteByID":
		if _, err := collection.DeleteOne(ctx, bson.M{"_id": r.ID}); err != nil {
			panic(err)
		}
		incrementTagCount(r.Tags, -1)
	case "update":
		if _, err := collection.UpdateOne(ctx, bson.M{"_id": r.ID}, bson.M{"$set": r}); err != nil {
			panic(err)
		}

	default:
		panic("unknow action " + action)
	}
}
