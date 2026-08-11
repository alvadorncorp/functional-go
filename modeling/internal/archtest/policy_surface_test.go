package archtest

import (
	"slices"
	"sort"
	"testing"
)

func policyProductionPackage(t *testing.T) *listedPackage {
	t.Helper()
	productions := productionPackages(t)
	for i := range productions {
		if productions[i].ImportPath == policyPackage {
			return &productions[i]
		}
	}
	t.Fatalf("production packages missing %s", policyPackage)
	return nil
}

func TestPolicyPackageHasExactDependencies(t *testing.T) {
	got := slices.Clone(policyProductionPackage(t).Imports)
	sort.Strings(got)
	want := []string{functionalPredicate, functionalOptional}
	if !slices.Equal(got, want) {
		t.Fatalf("policy imports = %v, want %v", got, want)
	}
}

func TestPolicyHasExactPublicSurface(t *testing.T) {
	got := productionNames(t, policyProductionPackage(t))
	slices.Sort(got)
	want := []string{
		"Choose", "Compose", "Constant", "Contramap", "Fallback", "FirstApplicable",
		"Map", "OrElse", "Policy", "When", "Zip",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("policy exports = %v, want %v", got, want)
	}
}

func TestRootDoesNotReexportPolicy(t *testing.T) {
	forbiddenNames := []string{
		"Choose", "Compose", "Constant", "Contramap", "Fallback", "FirstApplicable",
		"Map", "OrElse", "Policy", "When", "Zip",
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
