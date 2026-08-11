package archtest

import (
	"slices"
	"testing"
)

// TestGateCanFail proves every predicate this package relies on actually
// rejects a bad input. Without it a refactor could silently turn a gate into a
// tautology and every architecture test would stay green while the boundary
// eroded.
func TestGateCanFail(t *testing.T) {
	t.Run("dependencyAllowed rejects a foreign module", func(t *testing.T) {
		foreign := listedPackage{ImportPath: "github.com/alvadorncorp/gopackages/modeling"}
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
			if !importForbidden(path) {
				t.Errorf("importForbidden(%q) = false, want true", path)
			}
		}
		if importForbidden("context") {
			t.Error("importForbidden(\"context\") = true, want false")
		}
	})

	t.Run("forbiddenExportedNames covers the documented non-goals", func(t *testing.T) {
		for _, name := range []string{"Register", "Container", "Repository", "Retry", "Timeout", "DistributedTransaction"} {
			if _, forbidden := forbiddenExportedNames[name]; !forbidden {
				t.Errorf("%q missing from forbiddenExportedNames", name)
			}
		}
		if _, forbidden := forbiddenExportedNames["Provide"]; forbidden {
			t.Error("Provide remains forbidden after becoming input provision")
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

	t.Run("contextValueSelectors watches both directions", func(t *testing.T) {
		selectors := contextValueSelectors()
		for _, name := range []string{"WithValue", "Value"} {
			if _, watched := selectors[name]; !watched {
				t.Errorf("%q is not watched, want watched", name)
			}
		}
	})

	t.Run("expectedFiles declares every production package", func(t *testing.T) {
		envelope := expectedFiles()
		for _, pkg := range []string{workflowModule, actionPackage, transactionPackage} {
			files, declared := envelope[pkg]
			if !declared || len(files) == 0 {
				t.Errorf("%s missing from expectedFiles", pkg)
			}
			if !slices.IsSorted(files) {
				t.Errorf("%s file list is not sorted: %v", pkg, files)
			}
		}
	})
}
