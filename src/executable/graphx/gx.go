package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
)

type graphEdge struct {
	from   string
	to     string
	weight float64
}

type edgeOut struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight,omitempty"`
}

type analysisOptions struct {
	Mode       string
	From       string
	To         string
	Weighted   bool
	Undirected bool
	Viz        string
	Source     string
}

type analysisResult struct {
	Source     string     `json:"source,omitempty"`
	Mode       string     `json:"mode"`
	Nodes      int        `json:"nodes"`
	Edges      int        `json:"edges"`
	Weighted   bool       `json:"weighted"`
	Undirected bool       `json:"undirected"`
	HasCycle   *bool      `json:"has_cycle,omitempty"`
	Cycle      []string   `json:"cycle,omitempty"`
	Components [][]string `json:"components,omitempty"`
	Order      []string   `json:"order,omitempty"`
	Path       []string   `json:"path,omitempty"`
	Hops       *int       `json:"hops,omitempty"`
	Total      *float64   `json:"total_weight,omitempty"`
	MSTEdges   []edgeOut  `json:"mst_edges,omitempty"`
	Warning    string     `json:"warning,omitempty"`
	Error      string     `json:"error,omitempty"`
	Viz        string     `json:"viz,omitempty"`
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}

func splitFields(line, sep string) []string {
	if sep == "" {
		return strings.Fields(line)
	}
	parts := strings.Split(line, sep)
	res := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		res = append(res, part)
	}
	return res
}

func parseGraphEdges(reader io.Reader, weighted bool, sep string) ([]graphEdge, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	edges := make([]graphEdge, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := splitFields(line, sep)
		if len(parts) < 2 {
			return nil, fmt.Errorf("line %d: expected at least 2 columns, got %d", lineNo, len(parts))
		}
		edge := graphEdge{from: parts[0], to: parts[1], weight: 1}
		if weighted {
			if len(parts) < 3 {
				return nil, fmt.Errorf("line %d: expected 3 columns for weighted graph", lineNo)
			}
			weight, err := strconv.ParseFloat(parts[2], 64)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid weight %q: %w", lineNo, parts[2], err)
			}
			edge.weight = weight
		}
		edges = append(edges, edge)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}

func countNodes(edges []graphEdge) int {
	nodes := cw.NewSetT[string]()
	for _, edge := range edges {
		nodes.Add(edge.from)
		nodes.Add(edge.to)
	}
	return nodes.Size()
}

func sortedNodes(edges []graphEdge) []string {
	nodeSet := cw.NewSetT[string]()
	for _, edge := range edges {
		nodeSet.Add(edge.from)
		nodeSet.Add(edge.to)
	}
	nodes := nodeSet.ToSlice()
	sort.Strings(nodes)
	return nodes
}

func buildDirected(edges []graphEdge) *cw.DirectedGraph[string] {
	g := cw.NewDirectedGraph[string](nil)
	for _, edge := range edges {
		g.AddEdge(edge.from, edge.to)
	}
	return g
}

func buildUndirected(edges []graphEdge) *cw.UndirectedGraph[string] {
	g := cw.NewUndirectedGraph[string](nil)
	for _, edge := range edges {
		g.AddEdge(edge.from, edge.to)
	}
	return g
}

func buildWeightedUndirected(edges []graphEdge) *cw.WeightedUndirectedGraph[string] {
	g := cw.NewWeightedUndirectedGraph[string](nil)
	for _, edge := range edges {
		g.AddEdge(edge.from, edge.to, edge.weight)
	}
	return g
}

func normalizeComponents(groups [][]string) [][]string {
	res := make([][]string, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}
		sort.Strings(group)
		res = append(res, group)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i][0] == res[j][0] {
			return len(res[i]) < len(res[j])
		}
		return res[i][0] < res[j][0]
	})
	return res
}

func reverseStrings(arr []string) {
	for i := 0; i < len(arr)/2; i++ {
		arr[i], arr[len(arr)-i-1] = arr[len(arr)-i-1], arr[i]
	}
}

func formatWeight(weight float64) string {
	if math.Abs(weight-math.Round(weight)) < 1e-9 {
		return strconv.FormatInt(int64(math.Round(weight)), 10)
	}
	return strconv.FormatFloat(weight, 'f', -1, 64)
}

func csvJoin(items []string) string {
	return strings.Join(items, ";")
}

func bfsShortestPath(edges []graphEdge, from, to string, undirected bool) ([]string, error) {
	if from == to {
		return []string{from}, nil
	}
	adj := make(map[string][]string)
	for _, edge := range edges {
		adj[edge.from] = append(adj[edge.from], edge.to)
		if undirected {
			adj[edge.to] = append(adj[edge.to], edge.from)
		}
	}
	queue := []string{from}
	visited := map[string]bool{from: true}
	prev := make(map[string]string)
	for head := 0; head < len(queue); head++ {
		curr := queue[head]
		if curr == to {
			break
		}
		for _, next := range adj[curr] {
			if visited[next] {
				continue
			}
			visited[next] = true
			prev[next] = curr
			queue = append(queue, next)
		}
	}
	if !visited[to] {
		return nil, fmt.Errorf("no path from %q to %q", from, to)
	}
	path := []string{to}
	for curr := to; curr != from; {
		p, ok := prev[curr]
		if !ok {
			return nil, fmt.Errorf("no path from %q to %q", from, to)
		}
		path = append(path, p)
		curr = p
	}
	reverseStrings(path)
	return path, nil
}

func bellmanFordShortestPath(edges []graphEdge, from, to string, undirected bool) ([]string, float64, bool, error) {
	nodes := make(map[string]struct{})
	for _, edge := range edges {
		nodes[edge.from] = struct{}{}
		nodes[edge.to] = struct{}{}
	}
	nodes[from] = struct{}{}
	nodes[to] = struct{}{}
	n := len(nodes)
	if n == 0 {
		return nil, 0, false, fmt.Errorf("empty graph")
	}
	const eps = 1e-12
	inf := math.MaxFloat64 / 4
	dist := make(map[string]float64, n)
	prev := make(map[string]string, n)
	for node := range nodes {
		dist[node] = inf
	}
	dist[from] = 0

	relax := func(u, v string, weight float64) bool {
		if dist[u] >= inf/2 {
			return false
		}
		next := dist[u] + weight
		if next+eps < dist[v] {
			dist[v] = next
			prev[v] = u
			return true
		}
		return false
	}
	canRelax := func(u, v string, weight float64) bool {
		if dist[u] >= inf/2 {
			return false
		}
		return dist[u]+weight+eps < dist[v]
	}

	for i := 0; i < n-1; i++ {
		changed := false
		for _, edge := range edges {
			if relax(edge.from, edge.to, edge.weight) {
				changed = true
			}
			if undirected && relax(edge.to, edge.from, edge.weight) {
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	negativeCycle := false
	for _, edge := range edges {
		if canRelax(edge.from, edge.to, edge.weight) {
			negativeCycle = true
			break
		}
		if undirected && canRelax(edge.to, edge.from, edge.weight) {
			negativeCycle = true
			break
		}
	}

	if dist[to] >= inf/2 {
		return nil, 0, negativeCycle, fmt.Errorf("no path from %q to %q", from, to)
	}

	path := []string{to}
	for curr := to; curr != from; {
		p, ok := prev[curr]
		if !ok {
			return nil, 0, negativeCycle, fmt.Errorf("no path from %q to %q", from, to)
		}
		path = append(path, p)
		curr = p
		if len(path) > n+1 {
			return nil, 0, negativeCycle, fmt.Errorf("path reconstruction overflow")
		}
	}
	reverseStrings(path)
	return path, dist[to], negativeCycle, nil
}

func buildDOT(edges []graphEdge, undirected, weighted bool) string {
	lines := make([]string, 0, len(edges)+8)
	edgeOp := "->"
	header := "digraph G {"
	if undirected {
		edgeOp = "--"
		header = "graph G {"
	}
	lines = append(lines, header)
	nodes := sortedNodes(edges)
	for _, node := range nodes {
		lines = append(lines, fmt.Sprintf("  \"%s\";", strings.ReplaceAll(node, "\"", "\\\"")))
	}
	for _, edge := range edges {
		from := strings.ReplaceAll(edge.from, "\"", "\\\"")
		to := strings.ReplaceAll(edge.to, "\"", "\\\"")
		if weighted {
			lines = append(lines, fmt.Sprintf("  \"%s\" %s \"%s\" [label=\"%s\"] ;", from, edgeOp, to, formatWeight(edge.weight)))
		} else {
			lines = append(lines, fmt.Sprintf("  \"%s\" %s \"%s\";", from, edgeOp, to))
		}
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func mermaidID(idx int) string {
	return fmt.Sprintf("N%d", idx)
}

func buildMermaid(edges []graphEdge, undirected, weighted bool) string {
	nodes := sortedNodes(edges)
	idMap := make(map[string]string, len(nodes))
	lines := []string{"graph TD"}
	for idx, node := range nodes {
		id := mermaidID(idx)
		idMap[node] = id
		label := strings.ReplaceAll(node, "\"", "\\\"")
		lines = append(lines, fmt.Sprintf("    %s[\"%s\"]", id, label))
	}
	edgeOp := "-->"
	if undirected {
		edgeOp = "---"
	}
	for _, edge := range edges {
		fromID := idMap[edge.from]
		toID := idMap[edge.to]
		if weighted {
			lines = append(lines, fmt.Sprintf("    %s %s|%s| %s", fromID, edgeOp, formatWeight(edge.weight), toID))
		} else {
			lines = append(lines, fmt.Sprintf("    %s %s %s", fromID, edgeOp, toID))
		}
	}
	return strings.Join(lines, "\n")
}

func buildViz(edges []graphEdge, viz string, undirected, weighted bool) string {
	switch strings.ToLower(viz) {
	case "dot":
		return buildDOT(edges, undirected, weighted)
	case "mermaid", "mmd":
		return buildMermaid(edges, undirected, weighted)
	default:
		return ""
	}
}

func analyzeGraph(edges []graphEdge, opts analysisOptions) analysisResult {
	result := analysisResult{
		Source:     opts.Source,
		Mode:       opts.Mode,
		Nodes:      countNodes(edges),
		Edges:      len(edges),
		Weighted:   opts.Weighted,
		Undirected: opts.Undirected,
	}
	mode := strings.ToLower(opts.Mode)
	switch mode {
	case "cycle", "cy":
		if opts.Undirected {
			g := buildUndirected(edges)
			has := g.HasCycle()
			result.HasCycle = boolPtr(has)
		} else {
			g := buildDirected(edges)
			has := g.HasCycle()
			result.HasCycle = boolPtr(has)
			if has {
				result.Cycle = g.Cycle()
			}
		}
	case "scc":
		if opts.Undirected {
			result.Error = "scc is for directed graph only"
			break
		}
		g := buildDirected(edges)
		result.Components = normalizeComponents(g.StrongComponents())
	case "cc", "components":
		g := buildUndirected(edges)
		result.Components = normalizeComponents(g.Groups())
	case "topo", "tsort":
		if opts.Undirected {
			result.Error = "topo is for directed graph only"
			break
		}
		g := buildDirected(edges)
		orderIter := g.Sorted()
		if orderIter == nil {
			result.Error = "graph has cycle, topological order not available"
			cy := g.Cycle()
			if len(cy) > 0 {
				result.Cycle = cy
			}
			break
		}
		order := make([]string, 0, g.NumNodes())
		for node := range orderIter.Iterate() {
			order = append(order, node)
		}
		result.Order = order
	case "sp", "path":
		if opts.From == "" || opts.To == "" {
			result.Error = "sp mode requires -from and -to"
			break
		}
		if !opts.Weighted {
			path, err := bfsShortestPath(edges, opts.From, opts.To, opts.Undirected)
			if err != nil {
				result.Error = err.Error()
				break
			}
			result.Path = path
			result.Hops = intPtr(len(path) - 1)
			break
		}
		path, total, hasNegativeCycle, err := bellmanFordShortestPath(edges, opts.From, opts.To, opts.Undirected)
		if err != nil {
			result.Error = err.Error()
			break
		}
		if hasNegativeCycle {
			result.Warning = "negative cycle detected, shortest-path result may be unreliable"
		}
		result.Path = path
		result.Total = floatPtr(total)
	case "mst":
		g := buildWeightedUndirected(edges)
		forest := g.Mst()
		mstEdges := forest.Edges()
		sort.Slice(mstEdges, func(i, j int) bool {
			a1, b1 := mstEdges[i].V1(), mstEdges[i].V2()
			a2, b2 := mstEdges[j].V1(), mstEdges[j].V2()
			if a1 > b1 {
				a1, b1 = b1, a1
			}
			if a2 > b2 {
				a2, b2 = b2, a2
			}
			if a1 == a2 {
				if b1 == b2 {
					return mstEdges[i].Weight() < mstEdges[j].Weight()
				}
				return b1 < b2
			}
			return a1 < a2
		})
		edgesOut := make([]edgeOut, 0, len(mstEdges))
		for _, edge := range mstEdges {
			edgesOut = append(edgesOut, edgeOut{From: edge.V1(), To: edge.V2(), Weight: edge.Weight()})
		}
		result.MSTEdges = edgesOut
		total := forest.TotalWeight()
		result.Total = floatPtr(total)
		if len(mstEdges) < result.Nodes-1 {
			result.Warning = "graph is disconnected, result is a minimum spanning forest"
		}
	default:
		result.Error = fmt.Sprintf("unknown mode %q", opts.Mode)
	}
	if opts.Viz != "" {
		result.Viz = buildViz(edges, opts.Viz, opts.Undirected, opts.Weighted)
	}
	return result
}

func printModeHelp() {
	fmt.Println("modes:")
	fmt.Println("  cycle      detect cycle")
	fmt.Println("  scc        strongly connected components (directed)")
	fmt.Println("  topo       topological sort (directed DAG)")
	fmt.Println("  sp         shortest path: requires -from and -to")
	fmt.Println("  cc         connected components")
	fmt.Println("  mst        minimum spanning tree (undirected weighted)")
	fmt.Println()
	fmt.Println("input edge format:")
	fmt.Println("  unweighted: <from> <to>")
	fmt.Println("  weighted:   <from> <to> <weight>")
	fmt.Println("  comments:   lines starting with # or //")
	fmt.Println()
	fmt.Println("extra output:")
	fmt.Println("  -fmt text|json|csv   result output format")
	fmt.Println("  -viz dot|mermaid     include graph visualization text")
	fmt.Println("  -out <file>          write output to file")
	fmt.Println("  -dir <path>          batch analyze all graph files in directory")
}

func formatTextResult(r analysisResult) string {
	var b strings.Builder
	if r.Source != "" {
		b.WriteString(fmt.Sprintf("source: %s\n", r.Source))
	}
	b.WriteString(fmt.Sprintf("mode: %s\n", r.Mode))
	b.WriteString(fmt.Sprintf("graph: nodes=%d edges=%d weighted=%v undirected=%v\n", r.Nodes, r.Edges, r.Weighted, r.Undirected))
	if r.Error != "" {
		b.WriteString(fmt.Sprintf("error: %s\n", r.Error))
		if r.Viz != "" {
			b.WriteString("viz:\n")
			b.WriteString(r.Viz)
			b.WriteString("\n")
		}
		return b.String()
	}
	if r.HasCycle != nil {
		b.WriteString(fmt.Sprintf("has_cycle: %v\n", *r.HasCycle))
	}
	if len(r.Cycle) > 0 {
		b.WriteString(fmt.Sprintf("cycle: %s\n", strings.Join(r.Cycle, " -> ")))
	}
	if len(r.Components) > 0 {
		b.WriteString(fmt.Sprintf("components: %d\n", len(r.Components)))
		for i, group := range r.Components {
			b.WriteString(fmt.Sprintf("%d (%d): %s\n", i+1, len(group), strings.Join(group, ", ")))
		}
	}
	if len(r.Order) > 0 {
		b.WriteString(fmt.Sprintf("order: %s\n", strings.Join(r.Order, " -> ")))
	}
	if len(r.Path) > 0 {
		b.WriteString(fmt.Sprintf("path: %s\n", strings.Join(r.Path, " -> ")))
	}
	if r.Hops != nil {
		b.WriteString(fmt.Sprintf("hops: %d\n", *r.Hops))
	}
	if r.Total != nil {
		b.WriteString(fmt.Sprintf("total_weight: %.6f\n", *r.Total))
	}
	if len(r.MSTEdges) > 0 {
		b.WriteString(fmt.Sprintf("mst_edges: %d\n", len(r.MSTEdges)))
		for i, edge := range r.MSTEdges {
			b.WriteString(fmt.Sprintf("%d. (%s-%s) %.6f\n", i+1, edge.From, edge.To, edge.Weight))
		}
	}
	if r.Warning != "" {
		b.WriteString(fmt.Sprintf("warning: %s\n", r.Warning))
	}
	if r.Viz != "" {
		b.WriteString("viz:\n")
		b.WriteString(r.Viz)
		b.WriteString("\n")
	}
	return b.String()
}

func renderTextResults(results []analysisResult) string {
	if len(results) == 0 {
		return ""
	}
	chunks := make([]string, 0, len(results))
	for _, result := range results {
		chunks = append(chunks, strings.TrimRight(formatTextResult(result), "\n"))
	}
	return strings.Join(chunks, "\n\n---\n\n") + "\n"
}

func flattenComponents(groups [][]string) string {
	if len(groups) == 0 {
		return ""
	}
	parts := make([]string, 0, len(groups))
	for _, group := range groups {
		parts = append(parts, strings.Join(group, ","))
	}
	return strings.Join(parts, "|")
}

func flattenMSTEdges(edges []edgeOut) string {
	if len(edges) == 0 {
		return ""
	}
	parts := make([]string, 0, len(edges))
	for _, edge := range edges {
		parts = append(parts, fmt.Sprintf("%s-%s:%s", edge.From, edge.To, formatWeight(edge.Weight)))
	}
	return strings.Join(parts, "|")
}

func renderCSVResults(results []analysisResult) (string, error) {
	buf := &strings.Builder{}
	w := csv.NewWriter(buf)
	headers := []string{
		"source", "mode", "nodes", "edges", "weighted", "undirected", "has_cycle", "cycle", "components_count", "components", "order", "path", "hops", "total_weight", "mst_edges_count", "mst_edges", "warning", "error", "viz",
	}
	if err := w.Write(headers); err != nil {
		return "", err
	}
	for _, result := range results {
		hasCycle := ""
		if result.HasCycle != nil {
			hasCycle = strconv.FormatBool(*result.HasCycle)
		}
		hops := ""
		if result.Hops != nil {
			hops = strconv.Itoa(*result.Hops)
		}
		totalWeight := ""
		if result.Total != nil {
			totalWeight = strconv.FormatFloat(*result.Total, 'f', -1, 64)
		}
		row := []string{
			result.Source,
			result.Mode,
			strconv.Itoa(result.Nodes),
			strconv.Itoa(result.Edges),
			strconv.FormatBool(result.Weighted),
			strconv.FormatBool(result.Undirected),
			hasCycle,
			csvJoin(result.Cycle),
			strconv.Itoa(len(result.Components)),
			flattenComponents(result.Components),
			csvJoin(result.Order),
			csvJoin(result.Path),
			hops,
			totalWeight,
			strconv.Itoa(len(result.MSTEdges)),
			flattenMSTEdges(result.MSTEdges),
			result.Warning,
			result.Error,
			result.Viz,
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderResults(results []analysisResult, format string) (string, error) {
	switch strings.ToLower(format) {
	case "text", "txt", "":
		return renderTextResults(results), nil
	case "json":
		if len(results) == 1 {
			b, err := json.MarshalIndent(results[0], "", "  ")
			if err != nil {
				return "", err
			}
			return string(b) + "\n", nil
		}
		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b) + "\n", nil
	case "csv":
		return renderCSVResults(results)
	default:
		return "", fmt.Errorf("unsupported format %q, use text|json|csv", format)
	}
}

func normalizeExtSet(raw string) (map[string]bool, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return map[string]bool{}, true
	}
	set := map[string]bool{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		if part == "*" {
			return map[string]bool{}, true
		}
		if !strings.HasPrefix(part, ".") {
			part = "." + part
		}
		set[part] = true
	}
	if len(set) == 0 {
		return map[string]bool{}, true
	}
	return set, false
}

func collectInputFiles(singleFile, dirPath, extFilter string, recursive bool) ([]string, error) {
	files := make([]string, 0)
	if strings.TrimSpace(singleFile) != "" {
		if info, err := os.Stat(singleFile); err != nil {
			return nil, err
		} else if info.IsDir() {
			return nil, fmt.Errorf("-f expects a file, got directory %q", singleFile)
		}
		files = append(files, singleFile)
	}
	if strings.TrimSpace(dirPath) == "" {
		sort.Strings(files)
		return files, nil
	}
	extSet, allowAll := normalizeExtSet(extFilter)
	addIfMatch := func(path string) {
		if allowAll {
			files = append(files, path)
			return
		}
		ext := strings.ToLower(filepath.Ext(path))
		if extSet[ext] {
			files = append(files, path)
		}
	}
	if !recursive {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			addIfMatch(filepath.Join(dirPath, entry.Name()))
		}
		sort.Strings(files)
		return files, nil
	}
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		addIfMatch(path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func analyzeSingleSource(reader io.Reader, opts analysisOptions, sep string) analysisResult {
	edges, err := parseGraphEdges(reader, opts.Weighted, sep)
	if err != nil {
		res := analysisResult{Source: opts.Source, Mode: opts.Mode, Weighted: opts.Weighted, Undirected: opts.Undirected}
		res.Error = err.Error()
		return res
	}
	if len(edges) == 0 {
		res := analysisResult{Source: opts.Source, Mode: opts.Mode, Weighted: opts.Weighted, Undirected: opts.Undirected}
		res.Error = "no edges found in input"
		return res
	}
	return analyzeGraph(edges, opts)
}

func openInputFile(path string) (io.ReadCloser, error) {
	if path == "" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(path)
}

func main() {
	parser := terminalw.NewParser(terminalw.DisableParserNumber)
	parser.String("f", "", "edge list file (default: stdin)")
	parser.String("dir", "", "directory containing edge list files (batch mode)")
	parser.String("ext", ".txt,.graph,.edgelist,.csv", "extension filter for -dir, comma separated")
	parser.Bool("r", false, "recursive scan for -dir")
	parser.String("from", "", "source node for sp")
	parser.String("to", "", "target node for sp")
	parser.Bool("u", false, "undirected graph")
	parser.Bool("w", false, "weighted edges: <from> <to> <weight>")
	parser.String("sep", "", "field separator, default whitespace")
	parser.String("fmt", "text", "output format: text|json|csv")
	parser.String("viz", "", "visualization output: dot|mermaid")
	parser.String("out", "", "write result to file")
	parser.Bool("h", false, "print help")
	parser.ParseArgsCmd()

	if parser.ContainsFlagStrict("h") || parser.Empty() {
		parser.PrintDefaults()
		printModeHelp()
		return
	}

	args := parser.GetPositionalArgs(true)
	if len(args) == 0 {
		parser.PrintDefaults()
		printModeHelp()
		return
	}
	mode := strings.ToLower(args[0])
	inputFile := parser.GetFlagValueDefault("f", "")
	if inputFile == "" && len(args) > 1 {
		inputFile = args[1]
	}
	dirPath := parser.GetFlagValueDefault("dir", "")
	sep := parser.GetFlagValueDefault("sep", "")
	format := strings.ToLower(parser.GetFlagValueDefault("fmt", "text"))
	viz := strings.ToLower(parser.GetFlagValueDefault("viz", ""))
	outputFile := parser.GetFlagValueDefault("out", "")
	recursive := parser.ContainsFlagStrict("r")

	if viz != "" && viz != "dot" && viz != "mermaid" && viz != "mmd" {
		fmt.Fprintf(os.Stderr, "unsupported viz %q, use dot|mermaid\n", viz)
		os.Exit(1)
	}
	if viz == "mmd" {
		viz = "mermaid"
	}
	if format != "text" && format != "json" && format != "csv" {
		fmt.Fprintf(os.Stderr, "unsupported format %q, use text|json|csv\n", format)
		os.Exit(1)
	}

	weighted := parser.ContainsFlagStrict("w") || mode == "mst"
	undirected := parser.ContainsFlagStrict("u") || mode == "cc" || mode == "mst"
	from := parser.GetFlagValueDefault("from", "")
	to := parser.GetFlagValueDefault("to", "")

	opts := analysisOptions{
		Mode:       mode,
		From:       from,
		To:         to,
		Weighted:   weighted,
		Undirected: undirected,
		Viz:        viz,
	}

	files, err := collectInputFiles(inputFile, dirPath, parser.GetFlagValueDefault("ext", ".txt,.graph,.edgelist,.csv"), recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to collect input files: %v\n", err)
		os.Exit(1)
	}

	results := make([]analysisResult, 0)
	if len(files) == 0 {
		reader, err := openInputFile(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open input: %v\n", err)
			os.Exit(1)
		}
		defer reader.Close()
		opts.Source = "<stdin>"
		if strings.TrimSpace(inputFile) != "" {
			opts.Source = inputFile
		}
		results = append(results, analyzeSingleSource(reader, opts, sep))
	} else {
		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				res := analysisResult{Source: file, Mode: mode, Weighted: weighted, Undirected: undirected, Error: err.Error()}
				results = append(results, res)
				continue
			}
			localOpts := opts
			localOpts.Source = file
			results = append(results, analyzeSingleSource(f, localOpts, sep))
			f.Close()
		}
	}

	output, err := renderResults(results, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to render output: %v\n", err)
		os.Exit(1)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write output file: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(output)
	}

	hasErr := false
	for _, result := range results {
		if result.Error != "" {
			hasErr = true
			break
		}
	}
	if hasErr {
		os.Exit(1)
	}
}
