package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
	"golang.org/x/tools/go/packages"
)

var ignoreName = cw.NewSet()
var forceRebuildName = cw.NewSet()

type localPackage struct {
	importPath string
	dir        string
	files      []string
}

type dependencyGraph struct {
	modulePath  string
	moduleFile  string
	packages    map[string]*localPackage
	fileGraph   *cw.DirectedGraph[string]
	fileModTime map[string]time.Time
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

	packages, loaded, err := discoverLocalPackages(repoRoot, modulePath)
	if err != nil {
		return nil, err
	}

	fileGraph, fileModTime, err := buildFileDependencyGraph(packages, loaded)
	if err != nil {
		return nil, err
	}

	return &dependencyGraph{
		modulePath:  modulePath,
		moduleFile:  moduleFile,
		packages:    packages,
		fileGraph:   fileGraph,
		fileModTime: fileModTime,
	}, nil
}

func discoverLocalPackages(repoRoot, modulePath string) (map[string]*localPackage, map[string]*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
		Dir: repoRoot,
	}

	roots, err := packages.Load(cfg, "./src/executable/...")
	if err != nil {
		return nil, nil, err
	}
	if n := packages.PrintErrors(roots); n > 0 {
		return nil, nil, fmt.Errorf("failed to load local package dependency graph: %d package errors", n)
	}

	loaded := collectLoadedPackages(roots)
	packages := make(map[string]*localPackage)
	for importPath, pkg := range loaded {
		if !isLocalImportPath(importPath, modulePath) {
			continue
		}
		if len(pkg.CompiledGoFiles) == 0 {
			continue
		}

		files := make([]string, 0, len(pkg.CompiledGoFiles))
		for _, filename := range pkg.CompiledGoFiles {
			files = append(files, filepath.Clean(filename))
		}
		sort.Strings(files)

		packages[importPath] = &localPackage{
			importPath: importPath,
			dir:        filepath.Dir(files[0]),
			files:      files,
		}
	}

	return packages, loaded, nil
}

func collectLoadedPackages(roots []*packages.Package) map[string]*packages.Package {
	loaded := make(map[string]*packages.Package)
	var visit func(pkg *packages.Package)
	visit = func(pkg *packages.Package) {
		if pkg == nil {
			return
		}
		if existing, ok := loaded[pkg.PkgPath]; ok {
			if len(existing.CompiledGoFiles) >= len(pkg.CompiledGoFiles) {
				return
			}
		}
		loaded[pkg.PkgPath] = pkg
		for _, dep := range pkg.Imports {
			visit(dep)
		}
	}
	for _, pkg := range roots {
		visit(pkg)
	}
	return loaded
}

func isLocalImportPath(importPath, modulePath string) bool {
	return importPath == modulePath || strings.HasPrefix(importPath, modulePath+"/")
}

func buildFileDependencyGraph(localPackages map[string]*localPackage, loaded map[string]*packages.Package) (*cw.DirectedGraph[string], map[string]time.Time, error) {
	fileGraph := cw.NewDirectedGraph[string](nil)
	fileModTime := make(map[string]time.Time)

	for _, pkg := range localPackages {
		for _, filename := range pkg.files {
			fileGraph.AddNode(filename)
			info, err := os.Stat(filename)
			if err != nil {
				return nil, nil, err
			}
			fileModTime[filename] = info.ModTime()
		}
	}
	objDefFile := buildObjectDefinitionFiles(loaded, fileModTime)

	for importPath, pkg := range loaded {
		localPkg, ok := localPackages[importPath]
		if !ok || len(localPkg.files) == 0 {
			continue
		}
		if pkg.TypesInfo == nil || pkg.Fset == nil {
			continue
		}

		for i, syntaxFile := range pkg.Syntax {
			if i >= len(pkg.CompiledGoFiles) {
				break
			}
			srcFile := filepath.Clean(pkg.CompiledGoFiles[i])
			if _, ok := fileModTime[srcFile]; !ok {
				continue
			}

			ast.Inspect(syntaxFile, func(n ast.Node) bool {
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true
				}

				obj := pkg.TypesInfo.Uses[ident]
				if obj == nil {
					return true
				}
				depFile, ok := objDefFile[obj]
				if !ok {
					return true
				}
				if depFile == srcFile {
					return true
				}
				fileGraph.AddEdge(srcFile, depFile)
				return true
			})

			for _, imp := range syntaxFile.Imports {
				if imp.Name == nil {
					continue
				}
				alias := imp.Name.Name
				if alias != "_" && alias != "." {
					continue
				}
				depImportPath := strings.Trim(imp.Path.Value, "\"")
				depPkg, ok := localPackages[depImportPath]
				if !ok {
					continue
				}
				for _, depFile := range depPkg.files {
					if _, ok := fileModTime[depFile]; ok {
						fileGraph.AddEdge(srcFile, depFile)
					}
				}
			}
		}
	}

	fileGraph.Mark()
	return fileGraph, fileModTime, nil
}

func buildObjectDefinitionFiles(loaded map[string]*packages.Package, fileModTime map[string]time.Time) map[types.Object]string {
	objDefFile := make(map[types.Object]string)
	for _, pkg := range loaded {
		if pkg == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
			continue
		}
		for ident, obj := range pkg.TypesInfo.Defs {
			if ident == nil || obj == nil {
				continue
			}
			pos := ident.Pos()
			if pos == token.NoPos {
				continue
			}
			filename := filepath.Clean(pkg.Fset.PositionFor(pos, false).Filename)
			if filename == "" {
				continue
			}
			if _, ok := fileModTime[filename]; !ok {
				continue
			}
			objDefFile[obj] = filename
		}
	}
	return objDefFile
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
	pkg, ok := g.packages[importPath]
	if !ok {
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

	for _, filename := range pkg.files {
		if modTime, ok := g.fileModTime[filename]; ok && modTime.After(latest) {
			latest = modTime
		}
	}

	nodes := g.fileGraph.Nodes()
	seen := make(map[string]struct{})
	for _, startFile := range pkg.files {
		for _, node := range nodes {
			if _, ok := seen[node]; ok {
				continue
			}
			if !g.fileGraph.Reachable(startFile, node) {
				continue
			}
			seen[node] = struct{}{}
			if modTime, ok := g.fileModTime[node]; ok && modTime.After(latest) {
				latest = modTime
			}
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
