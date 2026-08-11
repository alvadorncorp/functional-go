package archtest

import (
	"slices"
	"testing"
)

// TestNoForbiddenExportedNames proves no package exports a name that would
// promise an engine, a container, a registry, a resilience policy, or
// distributed atomicity (R7).
func TestNoForbiddenExportedNames(t *testing.T) {
	for _, production := range productionPackages(t) {
		for _, name := range productionNames(t, &production) {
			if _, forbidden := forbiddenExportedNames[name]; forbidden {
				t.Errorf("%s exports forbidden runtime API %s", production.ImportPath, name)
			}
		}
	}
}

// exactSurface is the decided v0.1 public contract. Any addition, removal, or
// rename is a deliberate contract edit and must change this map in the same
// commit, which is what makes the surface reviewable rather than emergent.
func exactSurface() map[string][]string {
	return map[string][]string{
		workflowModule: {
			"Contramap", "FlatMap", "From", "Map", "MapFallible",
			"Pure", "Recover", "Tap", "Then", "Workflow",
		},
		actionPackage: {
			"Action", "Do", "Fail", "FlatMap", "Map",
			"MapFallible", "Pure", "Recover", "Run", "Tap",
		},
		transactionPackage: {
			"Boundary", "Within",
		},
	}
}

// TestExactPublicSurface proves each package exports exactly its decided set
// (R9).
func TestExactPublicSurface(t *testing.T) {
	surface := exactSurface()

	for _, production := range productionPackages(t) {
		want, declared := surface[production.ImportPath]
		if !declared {
			t.Errorf("%s is not declared in exactSurface", production.ImportPath)
			continue
		}
		got := productionNames(t, &production)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("%s exports %v, want %v", production.ImportPath, got, want)
		}
	}
}
