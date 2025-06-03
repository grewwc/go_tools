package test

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/grewwc/go_tools/src/algow"
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

func TestMst(t *testing.T) {
	f, _ := os.Open("./weighted.txt")
	g := cw.NewWeightedUndirectedGraph[int](nil)
	sortEdge := func(arr []*cw.Edge[int]) {
		sortw.StableSort(arr, func(e1, e2 *cw.Edge[int]) int {
			if e1.V1() < e2.V1() {
				return -1
			} else if e1.V1() > e2.V1() {
				return 1
			}
			if e1.V2() < e2.V2() {
				return -1
			} else if e1.V2() > e2.V2() {
				return 1
			}
			if e1.Weight() < e2.Weight() {
				return -1
			} else if e1.Weight() > e2.Weight() {
				return 1
			}
			return 0
		})
	}

	toSlice := func(arr []*cw.Edge[int]) []string {
		var res []string
		for _, e := range arr {
			res = append(res, e.String())
		}
		return res
	}

	compareEdge := func(s1 []string, s2 []string) bool {
		if len(s1) != len(s2) {
			return false
		}
		for i := range s1 {
			if s1[i] != s2[i] {
				return false
			}
		}
		return true
	}

	for line := range strw.SplitByToken(f, "\n", false) {
		parts := strings.Split(line, " ")
		if len(parts) != 3 {
			continue
		}
		u, _ := strconv.Atoi(parts[0])
		v, _ := strconv.Atoi(parts[1])
		weight, _ := strconv.ParseFloat(parts[2], 64)
		g.AddEdge(u, v, weight)
	}

	mst := g.Mst()
	edges := mst.Edges()
	sortEdge(edges)
	calc := toSlice(edges)
	truth := []string{
		"Edge{0,2,0.260}",
		"Edge{0,7,0.160}",
		"Edge{1,7,0.190}",
		"Edge{2,3,0.170}",
		"Edge{4,5,0.350}",
		"Edge{5,7,0.280}",
		"Edge{6,2,0.400}",
	}
	if !compareEdge(calc, truth) {
		t.Errorf("mst failed")
	}
	if algow.Abs(mst.TotalWeight()-1.81) > 1e-5 {
		t.Errorf("mst weight failed. calc:%.3f, truth:%.3f", mst.TotalWeight(), 1.81)
	}
}
