package features

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/sortw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RegisterListTags(parser *terminalw.Parser) {
	positional := parser.Positional

	parser.On(func(p *terminalw.Parser) bool {
		return internal.ListTagsAndOrderByTime ||
			parser.ContainsFlagStrict("tags") ||
			positional.Contains("tags", nil) ||
			positional.Contains("i", nil) ||
			positional.Contains("t", nil)
	}).Do(func() {
		all := parser.ContainsAnyFlagStrict("a", "all")
		var tags []internal.Tag
		var w int
		var err error
		buf := bytes.NewBufferString("")
		var cursor *mongo.Cursor
		var cli *mongo.Client
		var sortBy = "name"
		op1 := options.FindOptions{}
		var m bson.M = bson.M{}
		ctx := context.Background()
		isWindows := utilsw.WINDOWS == utilsw.GetPlatform()

		if all || internal.ListTagsAndOrderByTime {
			allRecords, _ := internal.ListRecords(-1, false, !internal.ListTagsAndOrderByTime || all, nil, false, "", false)

			// modified date map
			mtMap := internal.GetAllTagsModifiedDate(allRecords)
			testTags := cw.NewOrderedMapT[string, int]()
			for _, r := range allRecords {
				for _, t := range r.Tags {
					testTags.Put(t, testTags.GetOrDefault(t, 0)+1)
				}
			}
			for it := range testTags.Iter().Iterate() {
				v := it.Val()
				t := internal.Tag{Name: it.Key(), Count: int64(v), ModifiedDate: mtMap[it.Key()]}
				tags = append(tags, t)
			}
			if internal.ListTagsAndOrderByTime {
				// sort.Sort(tagSlice(tags))
				sortw.Sort(tags, func(t1, t2 internal.Tag) int {
					var res = 0
					if t1.ModifiedDate.Before(t2.ModifiedDate) {
						res = -1
					} else if t1.ModifiedDate.Equal(t2.ModifiedDate) {
						res = 0
					} else {
						res = 1
					}
					if internal.Reverse {
						res *= -1
					}
				})
			}
			tags = tags[:algow.Min(int(internal.RecordLimit), len(tags))]
			// fmt.Println("tags", tags)
			goto print
		}
		op1.SetLimit(internal.RecordLimit)
		if internal.Reverse {
			op1.SetSort(bson.M{sortBy: -1})
		} else {
			op1.SetSort(bson.M{sortBy: 1})
		}
		cli = internal.AtlasClient
		if internal.Remote.Get().(bool) {
			cli = internal.AtlasClient
		}
		if !internal.ListSpecial {
			m["name"] = bson.M{"$regex": primitive.Regex{Pattern: internal.BuildMongoRegularExpExclude(internal.SpecialTagPatterns)}}
		}
		cursor, err = cli.Database(internal.DbName).Collection(internal.TagCollectionName).Find(ctx, m, &op1)
		if err != nil {
			panic(err)
		}
		cursor.All(ctx, &tags)
	print:
		_, w, err = utilsw.GetTerminalSize()
		// filter records
		if parser.GetFlagValueDefault("ex", "") != "" {
			tags = internal.FilterTags(tags, utilsw.GetCommandList(parser.MustGetFlagVal("ex")))
		}
		for _, tag := range tags {
			if internal.Verbose {
				tag.Name = color.HiGreenString(tag.Name)
				internal.PrintSeperator()
				fmt.Println(utilsw.ToString(tag))
			} else {
				fmt.Fprintf(buf, `%s[%d]  `, tag.Name, tag.Count)
			}
		}
		if !internal.Verbose {
			if err == nil {
				terminalIndent := 2
				delimiter := "   "
				raw := strw.Wrap(buf.String(), w-terminalIndent, terminalIndent, delimiter)
				for _, line := range strw.SplitNoEmpty(raw, "\n") {
					arr := strw.SplitNoEmpty(line, " ")
					changedArr := make([]string, len(arr))
					for i := range arr {
						idx := strings.Index(arr[i], "[")
						if !isWindows {
							changedArr[i] = fmt.Sprintf("%s%s", color.HiGreenString(arr[i][:idx]), arr[i][idx:])
						} else { //not working on windows
							changedArr[i] = arr[i]
						}
					}
					fmt.Fprintf(color.Output, "%s%s\n", strings.Repeat(" ", terminalIndent), strings.Join(changedArr, delimiter))
				}
			} else {
				panic(err)
			}
		}
	})

}
