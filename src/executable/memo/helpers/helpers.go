package _helpers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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

func PromptYesOrNo(msg string) bool {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ans := strings.TrimSpace(scanner.Text())
	if strings.ToLower(ans) == "y" {
		return true
	}
	return false
}

func WriteUrls(titles []string) {

}
