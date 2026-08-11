package archtest

import (
	"go/ast"
	"go/token"
	"os"
	"slices"
	"strings"
	"testing"
)

// TestOnlyApprovedDependencyClosure proves the module's entire transitive
// closure is the standard library plus itself (R1).
func TestOnlyApprovedDependencyClosure(t *testing.T) {
	for _, production := range productionPackages(t) {
		t.Run(production.ImportPath, func(t *testing.T) {
			for _, dependency := range goList(t, "-deps", "-json", production.ImportPath) {
				if !dependencyAllowed(&dependency) {
					t.Errorf("%s depends on unapproved %s", production.ImportPath, dependency.ImportPath)
				}
			}
		})
	}
}

// TestGoModHasNoRequirements catches an accidental `go get` at the source: an
// import-graph gate only sees a stray dependency once some file imports it,
// while go.mod records it the moment it is added (R2).
func TestGoModHasNoRequirements(t *testing.T) {
	path := goModPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") || strings.HasPrefix(line, "replace") {
			t.Errorf("%s must not have a require/replace line, found: %q", path, line)
		}
	}
}

// TestActionIsASink proves the action package imports exactly "context", which
// makes it the sink of the module's dependency DAG (R4).
func TestActionIsASink(t *testing.T) {
	for _, production := range productionPackages(t) {
		if production.ImportPath != actionPackage {
			continue
		}
		if !slices.Equal(production.Imports, []string{"context", "errors"}) {
			t.Errorf("%s imports %v, want exactly [context errors]", actionPackage, production.Imports)
		}
		return
	}
	t.Fatalf("%s not found among production packages", actionPackage)
}

func TestPlanIsAStdlibLeaf(t *testing.T) {
	for _, production := range productionPackages(t) {
		if production.ImportPath != planPackage {
			continue
		}
		if !slices.Equal(production.Imports, []string{"slices"}) {
			t.Errorf("%s imports %v, want exactly [slices]", planPackage, production.Imports)
		}
		return
	}
	t.Fatalf("%s not found among production packages", planPackage)
}

func TestResilienceDependsOnlyOnActionAndStdlib(t *testing.T) {
	for _, production := range productionPackages(t) {
		if production.ImportPath != resiliencePackage {
			continue
		}
		want := []string{"context", actionPackage, "time"}
		got := slices.Clone(production.Imports)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("%s imports %v, want exactly %v", resiliencePackage, production.Imports, want)
		}
		return
	}
	t.Fatalf("%s not found among production packages", resiliencePackage)
}

// TestNoForbiddenDirectImports proves no production package reaches for
// infrastructure or for the primitives that would implement a documented
// non-goal (R5).
func TestNoForbiddenDirectImports(t *testing.T) {
	for _, production := range productionPackages(t) {
		t.Run(production.ImportPath, func(t *testing.T) {
			for _, imported := range production.Imports {
				if importForbidden(production.ImportPath, imported) {
					t.Errorf("%s directly imports forbidden %s", production.ImportPath, imported)
				}
			}
		})
	}
}

// TestCoreDoesNotDependOnModeling proves the core never gains a build
// dependency on a sibling module, however indirectly (R6).
func TestCoreDoesNotDependOnModeling(t *testing.T) {
	siblings := []string{
		"github.com/alvadorncorp/gopackages/modeling",
		"github.com/alvadorncorp/gopackages/functional",
		"github.com/alvadorncorp/gopackages/postgres",
	}

	// Walk each core package's own closure rather than the workspace-wide
	// wildcard, which would sweep in the nested examples module.
	for _, production := range productionPackages(t) {
		for _, dependency := range goList(t, "-deps", "-json", production.ImportPath) {
			for _, sibling := range siblings {
				if dependency.ImportPath == sibling ||
					strings.HasPrefix(dependency.ImportPath, sibling+"/") {
					t.Errorf("%s must not depend on %s", production.ImportPath, dependency.ImportPath)
				}
			}
		}
	}
}

// TestExamplesDoNotLeakIntoCore proves the nested examples module is invisible
// to the core's closure (R14).
func TestExamplesDoNotLeakIntoCore(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, dependency := range goList(t, "-deps", "-json", production.ImportPath) {
			if isExamples(dependency.ImportPath) {
				t.Errorf("%s must not depend on the examples module: %s",
					production.ImportPath, dependency.ImportPath)
			}
		}
	}
}

// TestNoMutableProductionGlobals proves the module holds no package-level
// mutable state, which is what makes every Workflow and Action safe to reuse
// concurrently (R8).
func TestNoMutableProductionGlobals(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, file := range productionFiles(t, &production) {
			for _, declaration := range file.Decls {
				generic, ok := declaration.(*ast.GenDecl)
				if ok && generic.Tok == token.VAR {
					t.Errorf("%s has a package-level var declaration", production.ImportPath)
				}
			}
		}
	}
}
