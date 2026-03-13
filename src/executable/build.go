package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var ignoreName = cw.NewSet()
var forceRebuildName = cw.NewSet()

type localPackage struct {
	importPath string
	dir        string
	files      []string
	imports    []string
	modTime    time.Time
}

type dependencyGraph struct {
	modulePath string
	moduleFile string
	packages   map[string]*localPackage
	graph      *cw.DirectedGraph[string]
}

type buildTarget struct {
	dir               string
	entryFilename     string
	outputFile        string
	packageImportPath string
}

func init() {
	if utilsw.GetPlatform() != utilsw.WINDOWS {
		// add Folder name, NOT filename
		ignoreName.AddAll("cat", "head", "rm", "stat", "tail", "touch", "open", "ls")
	}
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("h", false, "print help information")
	parser.Bool("f", false, "force rebuild (shortcut form)")
	parser.Bool("a", false, "force rebuild all")
	parser.Bool("force", false, "force rebuilds")
	parser.ParseArgsCmd("f", "force", "a", "h")
	var force bool = parser.ContainsFlag("f") || parser.ContainsFlag("force")
	var all bool = parser.ContainsFlagStrict("a")
	for fname := range parser.Positional.Iter().Iterate() {
		forceRebuildName.Add(fname.Value() + ".go")
	}

	executableRoot := utilsw.GetDirOfTheFile()
	repoRoot := filepath.Clean(filepath.Join(executableRoot, "../", "../"))
	graph, err := loadDependencyGraph(repoRoot)
	if err != nil {
		log.Fatalln(err)
	}

	outputDir := filepath.Join(executableRoot, "../", "../", "bin/")
	if !utilsw.IsExist(outputDir) {
		os.MkdirAll(outputDir, os.ModePerm)
	} else if !utilsw.IsDir(outputDir) {
		log.Fatalf("cannot install, because %q is not a directory", outputDir)
	}
	targets, err := collectBuildTargets(executableRoot, outputDir, graph.modulePath)
	if err != nil {
		log.Fatalln(err)
	}
	for _, target := range targets {
		shouldBuild, err := needsBuild(target, graph, all, force)
		if err != nil {
			log.Println(err)
			continue
		}
		if !shouldBuild {
			continue
		}

		fmt.Printf("building %q\n", target.entryFilename)
		if err := buildBinary(target); err != nil {
			log.Println(err)
		}
	}
}

func loadDependencyGraph(repoRoot string) (*dependencyGraph, error) {
	moduleFile := filepath.Join(repoRoot, "go.mod")
	modulePath, err := readModulePath(moduleFile)
	if err != nil {
		return nil, err
	}

	packages, err := discoverLocalPackages(repoRoot, modulePath)
	if err != nil {
		return nil, err
	}

	graph := cw.NewDirectedGraph[string](nil)
	for importPath := range packages {
		graph.AddNode(importPath)
	}
	for importPath, pkg := range packages {
		for _, dep := range pkg.imports {
			graph.AddEdge(importPath, dep)
		}
	}
	graph.Mark()
	if graph.HasCycle() {
		return nil, fmt.Errorf("local package import graph contains a cycle: %v", graph.Cycle())
	}

	return &dependencyGraph{
		modulePath: modulePath,
		moduleFile: moduleFile,
		packages:   packages,
		graph:      graph,
	}, nil
}

func discoverLocalPackages(repoRoot, modulePath string) (map[string]*localPackage, error) {
	packages := make(map[string]*localPackage)
	err := filepath.WalkDir(repoRoot, func(curr string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		switch d.Name() {
		case ".git", "bin", "vendor":
			return filepath.SkipDir
		}

		pkg, ok, err := parseLocalPackage(repoRoot, curr, modulePath)
		if err != nil {
			return err
		}
		if ok {
			packages[pkg.importPath] = pkg
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		filtered := make([]string, 0, len(pkg.imports))
		seen := make(map[string]struct{})
		for _, dep := range pkg.imports {
			if _, ok := packages[dep]; !ok {
				continue
			}
			if _, ok := seen[dep]; ok {
				continue
			}
			seen[dep] = struct{}{}
			filtered = append(filtered, dep)
		}
		pkg.imports = filtered
	}

	return packages, nil
}

func parseLocalPackage(repoRoot, dir, modulePath string) (*localPackage, bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, false, err
	}

	files := make([]string, 0, len(entries))
	imports := make(map[string]struct{})
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filename := filepath.Join(dir, name)
		files = append(files, filename)
		parsed, err := parser.ParseFile(fset, filename, nil, parser.ImportsOnly)
		if err != nil {
			return nil, false, err
		}
		for _, item := range parsed.Imports {
			importPath := strings.Trim(item.Path.Value, "\"")
			if strings.HasPrefix(importPath, modulePath+"/") {
				imports[importPath] = struct{}{}
			}
		}
	}
	if len(files) == 0 {
		return nil, false, nil
	}

	relDir, err := filepath.Rel(repoRoot, dir)
	if err != nil {
		return nil, false, err
	}
	importPath := modulePath
	if relDir != "." {
		importPath = path.Join(modulePath, filepath.ToSlash(relDir))
	}

	modTime, err := newestFileModTime(files)
	if err != nil {
		return nil, false, err
	}

	pkg := &localPackage{
		importPath: importPath,
		dir:        dir,
		files:      files,
		imports:    make([]string, 0, len(imports)),
		modTime:    modTime,
	}
	for dep := range imports {
		pkg.imports = append(pkg.imports, dep)
	}
	return pkg, true, nil
}

func newestFileModTime(files []string) (time.Time, error) {
	var latest time.Time
	for _, filename := range files {
		info, err := os.Stat(filename)
		if err != nil {
			return time.Time{}, err
		}
		if info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest, nil
}

func readModulePath(goModFile string) (string, error) {
	data, err := os.ReadFile(goModFile)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("cannot find module path in %q", goModFile)
}

func collectBuildTargets(executableRoot, outputDir, modulePath string) ([]buildTarget, error) {
	subdirs := utilsw.LsDir(executableRoot, nil, nil)
	targets := make([]buildTarget, 0, len(subdirs))
	for _, subdir := range subdirs {
		trimmed := strings.TrimSpace(subdir)
		dir := filepath.Join(executableRoot, subdir)
		if !utilsw.IsDir(dir) || trimmed == "bin" || ignoreName.Contains(trimmed) {
			continue
		}

		entryFilename, ok := findEntryFile(dir)
		if !ok {
			continue
		}

		outputFile := filepath.Join(outputDir, utilsw.TrimFileExt(entryFilename))
		if utilsw.GetPlatform() == utilsw.WINDOWS {
			outputFile += ".exe"
		}

		targets = append(targets, buildTarget{
			dir:               dir,
			entryFilename:     entryFilename,
			outputFile:        outputFile,
			packageImportPath: path.Join(modulePath, "src", "executable", filepath.ToSlash(trimmed)),
		})
	}
	return targets, nil
}

func findEntryFile(dir string) (string, bool) {
	var filename string
	for _, name := range utilsw.LsDir(dir, nil, nil) {
		if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
			continue
		}
		filename = name
	}
	return filename, filename != ""
}

func needsBuild(target buildTarget, graph *dependencyGraph, all, force bool) (bool, error) {
	if all || force || forceRebuildName.Contains(target.entryFilename) {
		return true, nil
	}

	info, err := os.Stat(target.outputFile)
	if os.IsNotExist(err) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	latest, err := graph.latestDependencyModTime(target.packageImportPath)
	if err != nil {
		return false, err
	}
	return latest.After(info.ModTime()), nil
}

func (g *dependencyGraph) latestDependencyModTime(importPath string) (time.Time, error) {
	if _, ok := g.packages[importPath]; !ok {
		return time.Time{}, fmt.Errorf("cannot find local package %q", importPath)
	}

	var latest time.Time
	if g.moduleFile != "" {
		info, err := os.Stat(g.moduleFile)
		if err != nil {
			return time.Time{}, err
		}
		latest = info.ModTime()
	}

	for _, node := range g.graph.Nodes() {
		if !g.graph.Reachable(importPath, node) {
			continue
		}
		pkg, ok := g.packages[node]
		if !ok {
			continue
		}
		if pkg.modTime.After(latest) {
			latest = pkg.modTime
		}
	}
	return latest, nil
}

func buildBinary(target buildTarget) error {
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", target.outputFile)
	cmd.Dir = target.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
