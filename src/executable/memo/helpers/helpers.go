package helpers

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

func CollectionExists(db *mongo.Database, ctx context.Context, collectionName string) bool {
	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		if name == collectionName {
			return true
		}
	}
	return false
}
