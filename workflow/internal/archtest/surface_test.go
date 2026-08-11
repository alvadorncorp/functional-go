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
			if exportForbidden(production.ImportPath, name) {
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
			"Contramap", "From", "Map", "MapFallible", "Provide", "Then", "Workflow",
		},
		actionPackage: {
			"AcquireUseRelease", "Action", "Attempt", "AttemptResult", "Both", "Bracket", "Do", "Ensure",
			"Fail", "FlatMap", "Map", "MapFallible", "Pair", "Pure", "Recover", "Run", "Tap", "Zip",
		},
		planPackage: {
			"Kind", "KindEmpty", "KindSequence", "KindStep", "KindTransaction",
			"Plan", "Sequence", "Step", "Transaction",
		},
		resiliencePackage: {
			"Backoff", "Retry", "RetryPolicy", "Timeout",
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
