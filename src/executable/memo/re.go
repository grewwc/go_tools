package main

import (
	"fmt"
	"math"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/executable/memo/internal/features"

	"github.com/grewwc/go_tools/src/terminalw"
)

func main() {
	parser := terminalw.NewParser()
	parser.Bool("i", false, "insert a record")
	parser.String("ct", "", "change a record title")
	parser.String("u", "", "update a record")
	parser.String("d", "", "delete a record")
	parser.Int("n", 100, "# of records to list")
	parser.Bool("h", false, "print help information")
	parser.String("push", "", "push one record from the current local backend (mongo/sqlite) to remote sqlite managed by re, requires --host or .configW:re.remote.host")
	parser.String("pull", "", "pull one record from remote sqlite into the current local backend (mongo/sqlite), requires --host or .configW:re.remote.host")
	parser.String("host", internal.DefaultRemoteHost, "remote host for sqlite sync, e.g. 10.37.110.250, user@10.37.110.250, or user@10.37.110.250:22. '.configW:re.remote.host'")
	parser.Bool("r", false, "reverse sort")
	parser.Bool("all", false, "including all record")
	parser.Bool("a", false, "shortcut for -all")
	parser.String("f", "", "finish a record")
	parser.String("nf", "", "set a record UNFINISHED")
	parser.String("t", "", "search by tags")
	parser.Bool("include-finished", false, "include finished record")
	parser.String("add-tag", "", "add tags for a record")
	parser.String("del-tag", "", "delete tags for a record")
	parser.String("clean-tag", "", "clean all the records having the tag")
	parser.Bool("tags", false, "list all tags")
	parser.Bool("and", false, "use and logic to match tags")
	parser.Bool("v", false, "verbose (show modify/add time, verbose)")
	parser.String("file", "", "read title from a file, for '-u' & '-ct', file serve as bool, for '-i', needs to pass filename")
	parser.Bool("e", false, "read from editor")
	parser.String("title", "", "search by title")
	parser.String("c", "", "content (alias for title)")
	parser.String("search", "", "fuzzy search records")
	parser.String("out", "", fmt.Sprintf("output to text file (default is %s)", internal.DefaultTxtOutputName))
	parser.Bool("remote", false, "operate on the remote server specified by --host")
	parser.Bool("prev", false, "operate based on the previous ObjectIDs")
	parser.Bool("count", false, "only print the count, not the result")
	parser.Bool("prefix", false, "tag prefix")
	parser.Bool("pre", false, "tag prefix (short for -prefix)")
	parser.Bool("binary", false, "if the title is binary file")
	parser.Bool("b", false, "shortcut for -binary")
	parser.Bool("force", false, "force overwrite")
	parser.Bool("sp", false, fmt.Sprintf("if list tags started with special: %v (config in .configW->special.tags)", internal.SpecialTagPatterns.ToSlice()))
	parser.String("ex", "", "exclude some tag prefix when list tags")
	parser.Bool("code", false, "if use vscode as input editor (default false)")
	parser.Bool("s", false, "short format, only print titles")
	parser.Bool("l", false, "list tags")
	parser.String("backend", internal.DefaultBackendMode, "local backend: auto|mongo|sqlite. '.configW:re.backend'")

	parser.Alias("search", "q")
	parser.ParseArgsCmd()
	internal.SetRemoteHost(parser.GetFlagValueDefault("host", internal.DefaultRemoteHost))
	internal.SetLocalBackendMode(parser.GetFlagValueDefault("backend", internal.DefaultBackendMode))
	internal.Prefix = parser.ContainsAnyFlagStrict("prefix", "pre", "all", "a")
	internal.OnlyTags = parser.ContainsFlagStrict("s") || parser.CoExists("a", "s")

	internal.UseVsCode = parser.ContainsAllFlagStrict("code")
	if parser.GetNumArgs() != -1 {
		internal.RecordLimit = int64(parser.GetNumArgs())
	}

	if parser.ContainsFlagStrict("remote") {
		internal.InitRemote()
	}

	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	if parser.ContainsFlagStrict("n") {
		internal.RecordLimit = parser.MustGetFlagValAsInt64("n")
	}

	all := parser.ContainsFlagStrict("all") || (parser.ContainsFlag("a") &&
		!parser.ContainsFlagStrict("add-tag") && !parser.ContainsFlagStrict("del-tag") &&
		!parser.ContainsFlagStrict("tags")) && !parser.ContainsFlagStrict("binary")
	if all {
		internal.RecordLimit = math.MaxInt64
	}

	internal.ListSpecial = parser.ContainsFlagStrict("sp") || all
	internal.Reverse = parser.ContainsFlag("r") && !parser.ContainsAnyFlagStrict("prev", "remote", "prefix", "pre")
	internal.IncludeFinished = parser.ContainsFlagStrict("include-finished") || all

	internal.Verbose = parser.ContainsFlagStrict("v")
	internal.ListTagsAndOrderByTime = internal.OrderByTime(parser)
	internal.ToBinary = parser.ContainsAnyFlagStrict("binary", "b")
	if parser.ContainsFlagStrict("out") {
		internal.TxtOutputName, _ = parser.GetFlagVal("out")
		if internal.TxtOutputName == "" {
			internal.TxtOutputName = internal.DefaultTxtOutputName
		}
	}
	features.RegisterDefault(parser)
	features.RegisterNf(parser)
	features.RegisterF(parser)
	features.RegisterOpen(parser)
	features.RegisterCleanTag(parser)
	features.RegisterLog(parser)
	features.RegisterWeek(parser)
	features.RegisterListTags(parser)
	features.RegisterUpdate(parser)
	features.RegisterInsert(parser)
	features.RegisterChangeTitle(parser)
	features.RegisterDelete(parser)
	features.RegisterAddTag(parser)
	features.RegisterDelTag(parser)
	features.RegisterPush(parser)
	features.RegisterPull(parser)
	features.RegisterListByTagName(parser)
	features.RegisterListByTitle(parser)
	features.RegisterSearch(parser)

	parser.Execute()
}
