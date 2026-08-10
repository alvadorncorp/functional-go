package archtest

import (
	"path/filepath"
	"slices"
	"sort"
	"testing"
)

func TestPolicyPackageHasExactDependencies(t *testing.T) {
	for _, production := range productionPackages(t) {
		if production.ImportPath != policyPackage {
			continue
		}
		got := slices.Clone(production.Imports)
		sort.Strings(got)
		want := []string{functionalPredicate, functionalOptional}
		if !slices.Equal(got, want) {
			t.Fatalf("policy imports = %v, want %v", got, want)
		}
		return
	}
	t.Fatalf("production packages missing %s", policyPackage)
}

func TestPolicyFilesHaveExactPublicSurface(t *testing.T) {
	tests := map[string][]string{
		"conditional.go": {"Fallback", "FirstApplicable", "When"},
		"policy.go":      {"Constant", "Policy"},
		"transform.go":   {"Contramap", "Map"},
	}
	for filename, want := range tests {
		t.Run(filename, func(t *testing.T) {
			got := exportedNamesInSource(t, filepath.Join("..", "..", "policy", filename))
			if !slices.Equal(got, want) {
				t.Fatalf("exports in %s = %v, want %v", filename, got, want)
			}
		})
	}
}

func TestRootDoesNotReexportPolicy(t *testing.T) {
	forbiddenNames := []string{
		"Policy", "Constant", "Map", "Contramap", "When", "FirstApplicable", "Fallback",
	}
	for _, production := range productionPackages(t) {
		if production.ImportPath != modelingModule {
			continue
		}
		for _, got := range productionNames(t, &production) {
			for _, forbidden := range forbiddenNames {
				if got == forbidden {
					t.Errorf("root modeling reexports policy symbol %s", got)
				}
			}
		}
	}
}
