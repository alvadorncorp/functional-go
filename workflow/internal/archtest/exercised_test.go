package archtest

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestSurfaceIsExercised requires exact imported selectors or local result/field
// type reachability from those selectors. Compilation is a separate requirement.
func TestSurfaceIsExercised(t *testing.T) {
	consumed := consumerSelectors(t)
	graph := surfaceTypeGraph(t)
	reachSurfaceTypes(consumed, graph)
	for pkg, names := range exactSurface() {
		for _, name := range names {
			if !consumed[surfaceSymbol{pkg, name}] {
				t.Errorf("%s.%s has no consumer selector or reachable result/field type",
					pkg, name)
			}
		}
	}
}

func shortPackageName(importPath string) string {
	if importPath == workflowModule {
		return "workflow"
	}
	return importPath[strings.LastIndex(importPath, "/")+1:]
}

// Module production files and own examples count; ordinary own tests do not.
func consumerFile(relative string) bool {
	relative = filepath.ToSlash(relative)
	if !strings.HasSuffix(relative, ".go") {
		return false
	}
	if strings.HasPrefix(relative, "internal/archtest/") {
		return false
	}
	return !strings.HasSuffix(relative, "_test.go") || strings.HasSuffix(relative, "example_test.go")
}

func consumerSelectors(t *testing.T) map[surfaceSymbol]bool {
	t.Helper()
	root := moduleRoot(t)
	consumed := map[surfaceSymbol]bool{}
	scanConsumerTree(t, root, consumed)

	// Vet checks that the consumers compile.
	vet := exec.Command("go", "vet", "./...")
	vet.Dir = root
	if output, err := vet.CombinedOutput(); err != nil {
		t.Fatalf("go vet: %v\n%s", err, output)
	}

	return consumed
}

func scanConsumerTree(t *testing.T, root string, consumed map[surfaceSymbol]bool) {
	t.Helper()
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if !consumerFile(relative) {
			return nil
		}
		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return parseErr
		}
		return collectSurfaceSelectors(file, consumed)
	})
	if err != nil {
		t.Fatalf("scan consumers in %s: %v", root, err)
	}
}

func surfaceTypeGraph(t *testing.T) map[surfaceSymbol]surfaceTypeEdges {
	t.Helper()
	graph := map[surfaceSymbol]surfaceTypeEdges{}
	for pkg := range exactSurface() {
		for _, production := range goList(t, "-json", pkg) {
			for _, name := range production.GoFiles {
				path := filepath.Join(production.Dir, name)
				file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
				if err != nil {
					t.Fatalf("parse surface types in %s: %v", path, err)
				}
				addSurfaceTypeEdges(graph, pkg, file)
			}
		}
	}
	return graph
}
