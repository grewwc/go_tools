package internal

import (
	"time"

	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tag struct {
	ID           primitive.ObjectID `bson:"_id,ignoreempty"`
	Name         string             `bson:"name,ignoreempty"`
	Count        int64              `bson:"count,ignoreempty"`
	ModifiedDate time.Time
}

func (t *Tag) String() string {
	return utilsw.ToString(t, "ID", "Name", "Count")
}
