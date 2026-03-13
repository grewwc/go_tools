package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNeedsBuildWhenTransitiveDependencyIsNewer(t *testing.T) {
	repoRoot := t.TempDir()
	baseTime := time.Unix(1700000000, 0)
	writeTestFile(t, filepath.Join(repoRoot, "go.mod"), "module example.com/test\n\ngo 1.24.0\n", baseTime)
	writeTestFile(t, filepath.Join(repoRoot, "src", "executable", "app", "app.go"), `package main

import "example.com/test/src/lib/mid"

func main() {
	mid.Run()
}
`, baseTime.Add(time.Minute))
	writeTestFile(t, filepath.Join(repoRoot, "src", "lib", "mid", "mid.go"), `package mid

import "example.com/test/src/lib/leaf"

func Run() {
	leaf.Run()
}
`, baseTime.Add(2*time.Minute))
	writeTestFile(t, filepath.Join(repoRoot, "src", "lib", "leaf", "leaf.go"), `package leaf

func Run() {}
`, baseTime.Add(3*time.Minute))

	graph, err := loadDependencyGraph(repoRoot)
	if err != nil {
		t.Fatalf("loadDependencyGraph() error = %v", err)
	}

	outputDir := filepath.Join(repoRoot, "bin")
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := mustFindTarget(t, repoRoot, graph, outputDir, "app.go")
	writeTestFile(t, target.outputFile, "", baseTime.Add(150*time.Second))

	shouldBuild, err := needsBuild(target, graph, false, false)
	if err != nil {
		t.Fatalf("needsBuild() error = %v", err)
	}
	if !shouldBuild {
		t.Fatal("expected transitive dependency change to trigger rebuild")
	}
}

func TestNeedsBuildSkipsWhenBinaryIsNewest(t *testing.T) {
	repoRoot := t.TempDir()
	baseTime := time.Unix(1700000000, 0)
	writeTestFile(t, filepath.Join(repoRoot, "go.mod"), "module example.com/test\n\ngo 1.24.0\n", baseTime)
	writeTestFile(t, filepath.Join(repoRoot, "src", "executable", "app", "app.go"), `package main

import "example.com/test/src/lib/shared"

func main() {
	shared.Run()
}
`, baseTime.Add(time.Minute))
	writeTestFile(t, filepath.Join(repoRoot, "src", "lib", "shared", "shared.go"), `package shared

func Run() {}
`, baseTime.Add(2*time.Minute))

	graph, err := loadDependencyGraph(repoRoot)
	if err != nil {
		t.Fatalf("loadDependencyGraph() error = %v", err)
	}

	outputDir := filepath.Join(repoRoot, "bin")
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := mustFindTarget(t, repoRoot, graph, outputDir, "app.go")
	writeTestFile(t, target.outputFile, "", baseTime.Add(4*time.Minute))

	shouldBuild, err := needsBuild(target, graph, false, false)
	if err != nil {
		t.Fatalf("needsBuild() error = %v", err)
	}
	if shouldBuild {
		t.Fatal("expected newest binary to skip rebuild")
	}
}

func mustFindTarget(t *testing.T, repoRoot string, graph *dependencyGraph, outputDir, entryFilename string) buildTarget {
	t.Helper()
	targets, err := collectBuildTargets(filepath.Join(repoRoot, "src", "executable"), outputDir, graph.modulePath)
	if err != nil {
		t.Fatalf("collectBuildTargets() error = %v", err)
	}
	for _, target := range targets {
		if target.entryFilename == entryFilename {
			return target
		}
	}
	t.Fatalf("cannot find target %q", entryFilename)
	return buildTarget{}
}

func writeTestFile(t *testing.T, filename, content string, modTime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filename, err)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", filename, err)
	}
	if err := os.Chtimes(filename, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(%q) error = %v", filename, err)
	}
}
