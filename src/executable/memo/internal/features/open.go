package features

import (
	"fmt"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RegisterOpen(parser *terminalw.Parser) {
	positional := parser.Positional
	parser.On(func(p *terminalw.Parser) bool {
		return positional.Contains("open", nil) || positional.Contains("o", nil)
	}).Do(func() {
		positional.Delete("open", nil)
		positional.Delete("o", nil)
		internal.ListSpecial = true
		tags := positional.ToStringSlice()
		isObjectID := false
		if !positional.Empty() {
			isObjectID = internal.IsObjectID(tags[0])
		}
		// tags 里面可能是 objectid
		if len(tags) == 1 && isObjectID {
			objectID, _ := primitive.ObjectIDFromHex(tags[0])
			r := &internal.Record{ID: objectID}
			r.LoadByID()
			internal.WriteInfo([]*primitive.ObjectID{&r.ID}, []string{r.Title})
		}
		if !isObjectID && len(tags) > 0 {
			if _, written := internal.ListRecords(-1, true, true, tags, false, "", internal.Prefix); !written {
				fmt.Printf("there are NO urls associated with tags: %v (prefix: %v)\n", tags, internal.Prefix)
				return
			}
		}

		internal.ReadInfo(true)
	})
}
