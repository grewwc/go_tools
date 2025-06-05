package test

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/sortw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/typesw"
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

func initGraphFromFile(fname string) *cw.WeightedUndirectedGraph[int] {
	g := cw.NewWeightedUndirectedGraph[int](nil)
	f, _ := os.Open(fname)
	for line := range strw.SplitByToken(f, "\n", false) {
		parts := strw.SplitNoEmpty(line, " ")
		if len(parts) != 3 {
			continue
		}
		u, _ := strconv.Atoi(parts[0])
		v, _ := strconv.Atoi(parts[1])
		weight, _ := strconv.ParseFloat(parts[2], 64)
		g.AddEdge(u, v, weight)
	}
	return g
}

func initDirectedGraphFromFile(fname string) *cw.WeightedDirectedGraph[int] {
	g := cw.NewWeightedDirectedGraph[int](nil)
	f, _ := os.Open(fname)
	for line := range strw.SplitByToken(f, "\n", false) {
		parts := strw.SplitNoEmpty(line, " ")
		if len(parts) != 3 {
			continue
		}
		u, _ := strconv.Atoi(parts[0])
		v, _ := strconv.Atoi(parts[1])
		weight, _ := strconv.ParseFloat(parts[2], 64)
		g.AddEdge(u, v, weight)
	}
	return g
}

func TestMst(t *testing.T) {
	fname := "./weighted.txt"
	g := initGraphFromFile(fname)
	sortEdge := func(arr []*cw.Edge[int]) {
		sortw.StableSort(arr, func(e1, e2 *cw.Edge[int]) int {
			v11 := algow.Min(e1.V1(), e1.V2())
			v12 := algow.Max(e1.V1(), e1.V2())
			v21 := algow.Min(e2.V1(), e2.V2())
			v22 := algow.Max(e2.V1(), e2.V2())
			if v11 < v21 {
				return -1
			} else if v11 > v21 {
				return 1
			}
			if v12 < v22 {
				return -1
			} else if v12 > v22 {
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
			b1 := []byte(s1[i])
			b2 := []byte(s2[i])
			sortw.Sort(b1, nil)
			sortw.Sort(b2, nil)
			if !bytes.Equal(b1, b2) {
				fmt.Println(string(b1), string(b2))
				return false
			}
		}
		return true
	}
	mst := g.Mst()
	edges := mst.Edges()
	sortEdge(edges)
	calc := toSlice(edges)
	truth := []string{
		"(0-2) 0.260",
		"(0-7) 0.160",
		"(1-7) 0.190",
		"(2-3) 0.170",
		"(6-2) 0.400",
		"(4-5) 0.350",
		"(5-7) 0.280",
	}
	if !compareEdge(calc, truth) {
		t.Errorf("mst failed")
		fmt.Println(calc)
		fmt.Println(mst.TotalWeight())
	}
	if algow.Abs(mst.TotalWeight()-1.81) > 1e-5 {
		t.Errorf("mst weight failed. calc:%.3f, truth:%.3f", mst.TotalWeight(), 1.81)
	}
}

func TestDijkstra(t *testing.T) {
	fname := "./test_dijkstra.txt"
	g := initGraphFromFile(fname)
	g.Mark()
	for edge := range g.ShortestPath(0, 3).Iterate() {
		fmt.Println(edge)
	}
}

func totalWeight(it typesw.IterableT[*cw.Edge[int]]) float64 {
	var res float64
	for val := range it.Iterate() {
		res += val.Weight()
	}
	return res
}

func TestBellmanford(t *testing.T) {
	fname := "./test_bellmanford.txt"
	g := initGraphFromFile(fname)

	if g.NumEdges() != 10 {
		t.Errorf("edge number failed. expected: 20, found: %d\n", g.NumEdges())
	}

	if g.NumNodes() != 5 {
		t.Errorf("nodes number failed. expected: 5, found: %d\n", g.NumNodes())
	}

	if g.NumGroups() != 1 {
		t.Errorf("group number failed. expected: 1, found: %d\n", g.NumGroups())
	}

	for e := range g.ShortestPath(0, 3).Iterate() {
		fmt.Println(e)
	}

	truth := []float64{
		0, // 0-0
		2, // 0-1
		3, // 0-2
		2, // 0-3
		1, // 0-4
	}
	for i := 0; i < len(truth); i++ {
		if algow.Abs(totalWeight(g.ShortestPath(0, i))-truth[i]) > 1e-3 {
			t.Errorf("0-%d weight error. expected: %f, found: %f\n", i, truth[i], totalWeight(g.ShortestPath(0, i)))
		}
	}

}

func TestDirectedWeightedGraph(t *testing.T) {
	fname := "./test_bellmanford.txt"
	g := initDirectedGraphFromFile(fname)
	truth := []float64{
		0, // 0-0
		2, // 0-1
		3, // 0-2
		2, // 0-3
		1, // 0-4
	}

	if g.NumEdges() != 18 {
		t.Errorf("edge number failed. expected: 20, found: %d\n", g.NumEdges())
	}

	if g.NumNodes() != 5 {
		t.Errorf("nodes number failed. expected: 5, found: %d\n", g.NumNodes())
	}
	g.Mark()
	fmt.Println(g.HasCycle())
	fmt.Println(g.Cycle())
	fmt.Println(g.Sorted())
	for i := 0; i < len(truth); i++ {
		if algow.Abs(totalWeight(g.ShortestPath(0, i))-truth[i]) > 1e-3 {
			t.Errorf("0-%d weight error. expected: %f, found: %f\n", i, truth[i], totalWeight(g.ShortestPath(0, i)))
		}
	}
}
