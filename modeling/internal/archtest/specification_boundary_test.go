package archtest

import (
	"slices"
	"strings"
	"testing"
)

func specificationProductionPackage(t *testing.T) *listedPackage {
	t.Helper()
	productions := productionPackages(t)
	for i := range productions {
		if productions[i].ImportPath == specificationPackage {
			return &productions[i]
		}
	}
	t.Fatalf("production packages missing %s", specificationPackage)
	return nil
}

func TestSpecificationHasExactPublicSurface(t *testing.T) {
	got := productionNames(t, specificationProductionPackage(t))
	slices.Sort(got)
	want := []string{"All", "Any", "Contramap", "Not", "Specification"}
	if !slices.Equal(got, want) {
		t.Fatalf("specification exports = %v, want %v", got, want)
	}
}

func TestSpecificationHasNoProductionImports(t *testing.T) {
	if got := specificationProductionPackage(t).Imports; len(got) != 0 {
		t.Fatalf("specification imports = %v, want none", got)
	}
}

func TestSpecificationDoesNotDependOnModelingCapabilities(t *testing.T) {
	for _, imported := range specificationProductionPackage(t).Imports {
		if imported == modelingModule || strings.HasPrefix(imported, modelingModule+"/") {
			t.Errorf("%s imports forbidden modeling capability %s", specificationPackage, imported)
		}
	}
}

func TestRootDoesNotReexportSpecification(t *testing.T) {
	forbidden := []string{"All", "Any", "Contramap", "Not", "Specification"}
	for _, production := range productionPackages(t) {
		if production.ImportPath != modelingModule {
			continue
		}
		for _, name := range productionNames(t, &production) {
			if slices.Contains(forbidden, name) {
				t.Errorf("root modeling reexports specification symbol %s", name)
			}
		}
	}
}
