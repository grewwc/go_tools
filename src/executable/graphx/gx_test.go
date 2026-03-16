package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGraphEdgesWeighted(t *testing.T) {
	input := strings.NewReader(`# comment
A B 1.5

// skip
B C 2
`)
	edges, err := parseGraphEdges(input, true, "")
	if err != nil {
		t.Fatalf("parseGraphEdges returned error: %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d", len(edges))
	}
	if edges[0].from != "A" || edges[0].to != "B" || edges[0].weight != 1.5 {
		t.Fatalf("unexpected edge[0]: %+v", edges[0])
	}
	if edges[1].from != "B" || edges[1].to != "C" || edges[1].weight != 2 {
		t.Fatalf("unexpected edge[1]: %+v", edges[1])
	}
}

func TestNormalizeComponents(t *testing.T) {
	groups := [][]string{{"b", "a"}, {}, {"d"}, {"c", "b"}}
	got := normalizeComponents(groups)
	if len(got) != 3 {
		t.Fatalf("expected 3 non-empty groups, got %d", len(got))
	}
	if strings.Join(got[0], ",") != "a,b" {
		t.Fatalf("unexpected first group: %v", got[0])
	}
	if strings.Join(got[1], ",") != "b,c" {
		t.Fatalf("unexpected second group: %v", got[1])
	}
	if strings.Join(got[2], ",") != "d" {
		t.Fatalf("unexpected third group: %v", got[2])
	}
}

func TestBuildVizOutputs(t *testing.T) {
	edges := []graphEdge{{from: "A", to: "B", weight: 2.5}, {from: "B", to: "C", weight: 1}}
	dot := buildViz(edges, "dot", false, true)
	if !strings.Contains(dot, "digraph G {") {
		t.Fatalf("dot output missing graph header: %q", dot)
	}
	if !strings.Contains(dot, "\"A\" -> \"B\" [label=\"2.5\"]") {
		t.Fatalf("dot output missing weighted edge: %q", dot)
	}

	mermaid := buildViz(edges, "mermaid", true, false)
	if !strings.Contains(mermaid, "graph TD") {
		t.Fatalf("mermaid output missing header: %q", mermaid)
	}
	if !strings.Contains(mermaid, "---") {
		t.Fatalf("mermaid output missing undirected edge syntax: %q", mermaid)
	}
}

func TestRenderResultsJSONAndCSV(t *testing.T) {
	hops := 2
	total := 3.5
	results := []analysisResult{{
		Source:     "case1.txt",
		Mode:       "sp",
		Nodes:      3,
		Edges:      2,
		Weighted:   true,
		Undirected: false,
		Path:       []string{"A", "B", "C"},
		Hops:       &hops,
		Total:      &total,
		Warning:    "warn",
		Viz:        "digraph G {}",
	}}

	jsonOut, err := renderResults(results, "json")
	if err != nil {
		t.Fatalf("renderResults json returned error: %v", err)
	}
	if !strings.Contains(jsonOut, "\"mode\": \"sp\"") {
		t.Fatalf("json output missing mode: %q", jsonOut)
	}
	if !strings.Contains(jsonOut, "\"path\": [") {
		t.Fatalf("json output missing path: %q", jsonOut)
	}

	csvOut, err := renderResults(results, "csv")
	if err != nil {
		t.Fatalf("renderResults csv returned error: %v", err)
	}
	if !strings.Contains(csvOut, "source,mode,nodes,edges,weighted") {
		t.Fatalf("csv output missing header: %q", csvOut)
	}
	if !strings.Contains(csvOut, "case1.txt,sp,3,2,true,false") {
		t.Fatalf("csv output missing row data: %q", csvOut)
	}
}

func TestCollectInputFiles(t *testing.T) {
	tmpDir := t.TempDir()
	rootA := filepath.Join(tmpDir, "a.txt")
	rootB := filepath.Join(tmpDir, "b.md")
	subDir := filepath.Join(tmpDir, "nested")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	subC := filepath.Join(subDir, "c.graph")

	if err := os.WriteFile(rootA, []byte("A B\n"), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", rootA, err)
	}
	if err := os.WriteFile(rootB, []byte("B C\n"), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", rootB, err)
	}
	if err := os.WriteFile(subC, []byte("C D\n"), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", subC, err)
	}

	files, err := collectInputFiles("", tmpDir, ".txt,.graph", false)
	if err != nil {
		t.Fatalf("collectInputFiles non-recursive returned error: %v", err)
	}
	if len(files) != 1 || files[0] != rootA {
		t.Fatalf("unexpected non-recursive files: %v", files)
	}

	files, err = collectInputFiles("", tmpDir, ".txt,.graph", true)
	if err != nil {
		t.Fatalf("collectInputFiles recursive returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 recursive matches, got %d: %v", len(files), files)
	}
	if files[0] != rootA || files[1] != subC {
		t.Fatalf("unexpected recursive files order/content: %v", files)
	}
}
