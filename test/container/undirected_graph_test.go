package test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/sortw"
	"github.com/grewwc/go_tools/src/strw"
)

func prepareData() []*cw.Tuple {
	data := `0 - 1
			0 - 2
			1 - 2
			1 - 3
			2 - 4
			3 - 4
			3 - 5
			5 - 6
			6 - 7
			7 - 8
			8 - 5

			9 - 10
			10 - 11
			11 - 9

			12 - 13
			13 - 14`
	var res []*cw.Tuple
	for line := range strw.SplitByToken(strings.NewReader(data), "\n", false) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "-")
		a, b := parts[0], parts[1]
		a, b = strings.TrimSpace(a), strings.TrimSpace(b)
		res = append(res, cw.NewTuple(a, b))
	}
	return res
}

func sortGroups(groups [][]string) {
	for _, group := range groups {
		sortw.Sort(group, func(s1, s2 string) int {
			i1, _ := strconv.Atoi(s1)
			i2, _ := strconv.Atoi(s2)
			return i1 - i2
		})
	}
	sortw.Sort(groups, func(g1, g2 []string) int {
		i1, _ := strconv.Atoi(g1[0])
		i2, _ := strconv.Atoi(g2[0])
		return i1 - i2
	})
}

func equalsGroup(g1, g2 [][]string) bool {
	if len(g1) != len(g2) {
		return false
	}
	for i := 0; i < len(g1); i++ {
		gg1, gg2 := g1[i], g2[i]
		if len(gg1) != len(gg2) {
			return false
		}
		for j := 0; j < len(gg1); j++ {
			if gg1[j] != gg2[j] {
				return false
			}
		}
	}
	return true
}

func connectionTest(g *cw.UndirectedGraph[string]) bool {
	if !g.Connected("1", "8") || !g.Connected("8", "1") {
		return false
	}
	if g.Connected("11", "2") {
		return false
	}
	return true
}

func TestUndirectedGraph(t *testing.T) {
	data := prepareData()
	g := cw.NewUndirectedGraph[string](nil)
	for _, tup := range data {
		g.AddEdge(tup.Get(0).(string), tup.Get(1).(string))
	}
	g.Mark()
	groups := g.Groups()
	sortGroups(groups)
	gtGroups := [][]string{
		{"0", "1", "2", "3", "4", "5", "6", "7", "8"},
		{"9", "10", "11"},
		{"12", "13", "14"},
	}
	if !equalsGroup(groups, gtGroups) {
		t.Errorf("groups are not the same.")
	}
	if !connectionTest(g) {
		t.Errorf("connection test error")
	}
}
