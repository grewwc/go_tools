package features

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func highlightSearchText(text string, pattern *regexp.Regexp) string {
	if strings.TrimSpace(text) == "" {
		return text
	}
	if pattern == nil {
		return color.HiWhiteString(text)
	}
	indices := pattern.FindAllStringIndex(text, -1)
	if len(indices) == 0 {
		return color.HiWhiteString(text)
	}
	var builder strings.Builder
	last := 0
	for _, idx := range indices {
		if idx[0] > last {
			builder.WriteString(color.HiWhiteString(text[last:idx[0]]))
		}
		builder.WriteString(color.RedString(text[idx[0]:idx[1]]))
		last = idx[1]
	}
	if last < len(text) {
		builder.WriteString(color.HiWhiteString(text[last:]))
	}
	return builder.String()
}

func listBySearch(parser *terminalw.Parser) {
	query := strings.TrimSpace(parser.GetFlagValueDefault("search", ""))
	tags := strw.SplitNoEmpty(strings.TrimSpace(parser.GetMultiFlagValDefault([]string{"t", "ta", "at"}, "")), " ")
	includeFinished := parser.ContainsAnyFlagStrict("all", "a", "include-finished")
	results := internal.SearchRecords(query, internal.RecordLimit, includeFinished,
		tags, parser.ContainsFlagStrict("and"), internal.Prefix)

	if parser.ContainsFlagStrict("count") {
		fmt.Printf("%d records found\n", len(results))
		return
	}

	if !parser.ContainsFlagStrict("out") && !internal.ToBinary {
		pattern := internal.SearchHighlightPattern(query)
		for idx := len(results) - 1; idx >= 0; idx-- {
			if idx < len(results)-1 {
				fmt.Println()
			}
			printSearchResult(idx, results[idx], query, pattern)
		}
		return
	}

	if internal.ToBinary {
		panic("not supported")
	}

	if (utilsw.IsExist(internal.TxtOutputName) && utilsw.PromptYesOrNo(fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", internal.TxtOutputName))) ||
		!utilsw.IsExist(internal.TxtOutputName) {
		buf := bytes.NewBufferString("")
		for idx := len(results) - 1; idx >= 0; idx-- {
			result := results[idx]
			buf.WriteString(fmt.Sprintf("%s score=%.3f %s\n", strings.Repeat("-", 10), result.Score, strings.Repeat("-", 10)))
			buf.WriteString(result.Record.Title)
			buf.WriteString("\n")
		}
		if err := os.WriteFile(internal.TxtOutputName, buf.Bytes(), 0666); err != nil {
			panic(err)
		}
	}
}

func printSearchResult(index int, result *internal.SearchResult, query string, pattern *regexp.Regexp) {
	headerParts := []string{color.HiBlueString("[%d]", index+1), color.HiWhiteString("score %.3f", result.Score)}
	if result.Record.Finished {
		headerParts = append(headerParts, color.YellowString("finished"))
	}
	fmt.Println(strings.Join(headerParts, "  "))
	if len(result.Record.Tags) > 0 {
		fmt.Printf("    %s %s\n", color.HiBlackString("tags:"), color.HiGreenString(strings.Join(result.Record.Tags, ", ")))
	}
	if internal.Verbose {
		display := *result.Record
		display.Tags = append([]string(nil), result.Record.Tags...)
		ignoreFields := []string{}
		if utilsw.IsText([]byte(display.Title)) {
			internal.ColoringRecord(&display, pattern)
		} else {
			display.Title = color.HiYellowString("<binary>")
		}
		for _, line := range strings.Split(utilsw.ToString(&display, ignoreFields...), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			fmt.Printf("    %s\n", line)
		}
		fmt.Printf("    %s %s\n", color.HiBlackString("id:"), color.HiRedString(result.Record.ID.Hex()))
		return
	}
	if utilsw.IsText([]byte(result.Record.Title)) {
		preview := result.Preview
		if len(preview) == 0 {
			preview = internal.SearchPreview(result.Record, query)
		}
		for _, line := range preview {
			printSearchPreviewLine(line, pattern)
		}
	} else {
		fmt.Printf("    %s %s\n", color.CyanString("-"), color.HiYellowString("<binary>"))
	}
	fmt.Printf("    %s %s\n", color.HiBlackString("id:"), color.HiRedString(result.Record.ID.Hex()))
}

func printSearchPreviewLine(line string, pattern *regexp.Regexp) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	prefix, url, hasURL := splitSearchPreviewURL(line)
	if hasURL && prefix != "" {
		fmt.Printf("    %s %s\n", color.CyanString("-"), highlightSearchText(prefix, pattern))
		fmt.Printf("      %s\n", highlightSearchText(url, pattern))
		return
	}
	fmt.Printf("    %s %s\n", color.CyanString("-"), highlightSearchText(line, pattern))
}

func splitSearchPreviewURL(line string) (string, string, bool) {
	lower := strings.ToLower(line)
	start := strings.Index(lower, "https://")
	if start < 0 {
		start = strings.Index(lower, "http://")
	}
	if start < 0 {
		return "", "", false
	}
	prefix := strings.TrimSpace(line[:start])
	url := strings.TrimSpace(line[start:])
	return prefix, url, url != ""
}

func RegisterSearch(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("search") && strings.TrimSpace(parser.GetFlagValueDefault("search", "")) != ""
	}).Do(func() {
		listBySearch(parser)
	})
}
