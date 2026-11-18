package archtest

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"testing"
)

// TestGateCanFail proves every predicate this package relies on actually
// rejects a bad input. Without it a refactor could silently turn a gate into a
// tautology and every architecture test would stay green while the boundary
// eroded.
func TestGateCanFail(t *testing.T) {
	t.Run("dependencyAllowed rejects a foreign module", func(t *testing.T) {
		foreign := listedPackage{ImportPath: "go.alvadorn.org/functional/modeling"}
		if dependencyAllowed(&foreign) {
			t.Error("dependencyAllowed accepted a foreign module")
		}
	})

	t.Run("dependencyAllowed accepts stdlib and the module itself", func(t *testing.T) {
		stdlib := listedPackage{ImportPath: "context", Standard: true}
		own := listedPackage{ImportPath: actionPackage}
		if !dependencyAllowed(&stdlib) || !dependencyAllowed(&own) {
			t.Error("dependencyAllowed rejected an approved dependency")
		}
	})

	t.Run("importForbidden rejects infrastructure and non-goal primitives", func(t *testing.T) {
		for _, path := range []string{"database/sql", "net/http", "reflect", "unsafe", "sync", "time"} {
			if !importForbidden(workflowModule, path) {
				t.Errorf("importForbidden(%q) = false, want true", path)
			}
		}
		if importForbidden(workflowModule, "context") {
			t.Error("importForbidden(\"context\") = true, want false")
		}
		if importForbidden(resiliencePackage, "time") {
			t.Error("resilience time import rejected")
		}
		if !importForbidden(workflowModule, "time") {
			t.Error("workflow time import accepted")
		}
	})

	t.Run("forbiddenExportedNames covers the documented non-goals", func(t *testing.T) {
		for _, name := range []string{
			"Container", "DistributedTransaction", "Fork", "Join", "ParZip", "Race",
			"Register", "Repository", "Retry", "Runner", "SagaEngine", "Timeout", "Worker",
		} {
			if _, forbidden := forbiddenExportedNames[name]; !forbidden {
				t.Errorf("%q missing from forbiddenExportedNames", name)
			}
		}
		if _, forbidden := forbiddenExportedNames["Provide"]; forbidden {
			t.Error("Provide remains forbidden after becoming input provision")
		}
		for _, allowed := range []string{"Attempt", "AttemptResult", "Pair", "Tap"} {
			if _, forbidden := forbiddenExportedNames[allowed]; forbidden {
				t.Errorf("%q is a decided algebra export but remains forbidden", allowed)
			}
		}
		if exportForbidden(resiliencePackage, "Retry") || exportForbidden(resiliencePackage, "Timeout") {
			t.Error("resilience Retry/Timeout export rejected")
		}
		if !exportForbidden(actionPackage, "Retry") || !exportForbidden(workflowModule, "Timeout") {
			t.Error("Retry/Timeout export accepted outside resilience")
		}
	})

	t.Run("allowedInternalEdges forbids the reverse edge", func(t *testing.T) {
		edges := allowedInternalEdges()
		if _, permitted := edges[actionPackage][workflowModule]; permitted {
			t.Error("action -> workflow is permitted, want forbidden")
		}
		if _, permitted := edges[workflowModule][actionPackage]; !permitted {
			t.Error("workflow -> action is forbidden, want permitted")
		}
		if len(edges[planPackage]) != 0 {
			t.Errorf("plan edges = %v, want none", edges[planPackage])
		}
		if _, permitted := edges[resiliencePackage][actionPackage]; !permitted {
			t.Error("resilience -> action is forbidden, want permitted")
		}
		for _, forbiddenTarget := range []string{workflowModule, planPackage, transactionPackage} {
			if _, permitted := edges[resiliencePackage][forbiddenTarget]; permitted {
				t.Errorf("resilience -> %s is permitted, want forbidden", forbiddenTarget)
			}
		}
	})

	t.Run("exactSurface is sorted and non-empty", func(t *testing.T) {
		for pkg, names := range exactSurface() {
			if len(names) == 0 {
				t.Errorf("%s declares an empty surface", pkg)
			}
			if !slices.IsSorted(names) {
				t.Errorf("%s surface is not sorted: %v", pkg, names)
			}
		}
	})

	t.Run("forbiddenContextSelectors watches storage creation and shielding", func(t *testing.T) {
		selectors := forbiddenContextSelectors()
		for _, name := range []string{
			"AfterFunc", "Background", "TODO", "Value", "WithCancel", "WithCancelCause",
			"WithDeadline", "WithDeadlineCause", "WithoutCancel", "WithTimeout",
			"WithTimeoutCause", "WithValue",
		} {
			if _, watched := selectors[name]; !watched {
				t.Errorf("%q is not watched, want watched", name)
			}
		}
	})

	t.Run("expectedFiles declares every production package", func(t *testing.T) {
		envelope := expectedFiles()
		for _, pkg := range []string{workflowModule, actionPackage, planPackage, resiliencePackage, transactionPackage} {
			files, declared := envelope[pkg]
			if !declared || len(files) == 0 {
				t.Errorf("%s missing from expectedFiles", pkg)
			}
			if !slices.IsSorted(files) {
				t.Errorf("%s file list is not sorted: %v", pkg, files)
			}
		}
	})

	t.Run("execution policy rejects go and recover", func(t *testing.T) {
		for name, source := range map[string]string{
			"go":      "package p; func f() { go func(){}() }",
			"recover": "package p; func f() { _ = recover() }",
		} {
			file, err := parser.ParseFile(token.NewFileSet(), name+".go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			found := ""
			ast.Inspect(file, func(node ast.Node) bool {
				if violation := forbiddenExecutionNode(node); violation != "" {
					found = violation
				}
				return true
			})
			if found == "" {
				t.Errorf("execution policy did not reject %s snippet", name)
			}
		}
	})

	t.Run("context policy rejects aliases dot imports and named receivers", func(t *testing.T) {
		for name, source := range map[string]string{
			"alias":  "package p; import c \"context\"; func f() { _ = c.Background() }",
			"derive": "package p; import c \"context\"; func f(parent c.Context) { _, _ = c.WithCancel(parent) }",
			"dot":    "package p; import . \"context\"; func f() { _ = Background() }",
			"value":  "package p; import \"context\"; func f(runCtx context.Context) { _ = runCtx.Value(\"key\") }",
		} {
			file, err := parser.ParseFile(token.NewFileSet(), name+".go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			if violation, forbidden := forbiddenContextUse(workflowModule, file); !forbidden {
				t.Errorf("context policy accepted %s snippet; violation=%q", name, violation)
			}
		}
	})

	t.Run("context policy permits WithTimeout only in resilience", func(t *testing.T) {
		source := `package p; import (
	"context"
	"time"
); func f(parent context.Context) { _, cancel := context.WithTimeout(parent, time.Second); cancel() }`
		file, err := parser.ParseFile(token.NewFileSet(), "withtimeout.go", source, 0)
		if err != nil {
			t.Fatal(err)
		}
		if violation, forbidden := forbiddenContextUse(resiliencePackage, file); forbidden {
			t.Errorf("resilience WithTimeout rejected; violation=%q", violation)
		}
		if violation, forbidden := forbiddenContextUse(workflowModule, file); !forbidden {
			t.Errorf("workflow WithTimeout accepted; violation=%q", violation)
		}
	})
}
