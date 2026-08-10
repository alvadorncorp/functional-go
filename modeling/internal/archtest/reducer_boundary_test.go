package archtest

import (
	"slices"
	"strings"
	"testing"
)

func reducerProductionPackage(t *testing.T) *listedPackage {
	t.Helper()
	productions := productionPackages(t)
	for i := range productions {
		if productions[i].ImportPath == reducerPackage {
			return &productions[i]
		}
	}
	t.Fatalf("package enumeration is missing %s", reducerPackage)
	return nil
}

func TestReducerHasExactPublicSurface(t *testing.T) {
	pkg := reducerProductionPackage(t)
	got := productionNames(t, pkg)
	slices.Sort(got)
	want := []string{"Reducer", "Replay"}
	if !slices.Equal(got, want) {
		t.Fatalf("reducer exports = %v, want %v", got, want)
	}
}

func TestReducerDoesNotDependOnModelingCapabilities(t *testing.T) {
	pkg := reducerProductionPackage(t)
	for _, imported := range pkg.Imports {
		if imported == modelingModule || strings.HasPrefix(imported, modelingModule+"/") {
			t.Errorf("%s imports forbidden modeling capability %s", reducerPackage, imported)
		}
	}
}

func TestReplayNameIsReducerOnly(t *testing.T) {
	for _, production := range productionPackages(t) {
		if production.ImportPath == reducerPackage {
			continue
		}
		if slices.Contains(productionNames(t, &production), "Replay") {
			t.Errorf("%s exports forbidden runtime API Replay", production.ImportPath)
		}
	}
}

func TestPackageEnumerationIsComplete(t *testing.T) {
	got := make([]string, 0, 7)
	for _, pkg := range goList(t, "-json", modelingModule+"/...") {
		got = append(got, pkg.ImportPath)
	}
	slices.Sort(got)

	want := []string{
		modelingModule,
		changePackage,
		archtestPackage,
		modelingModule + "/internal/decision",
		policyPackage,
		projectionPackage,
		reducerPackage,
	}
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("modeling packages = %v, want exactly %v", got, want)
	}
}

func TestExternallyImportableClassification(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		want       bool
	}{
		{name: "root", importPath: modelingModule, want: true},
		{name: "public change", importPath: changePackage, want: true},
		{name: "public policy", importPath: policyPackage, want: true},
		{name: "public projection", importPath: projectionPackage, want: true},
		{name: "public reducer", importPath: reducerPackage, want: true},
		{name: "internal root", importPath: modelingModule + "/internal", want: false},
		{name: "internal child", importPath: modelingModule + "/internal/decision", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := externallyImportable(test.importPath); got != test.want {
				t.Fatalf("externallyImportable(%q) = %t, want %t", test.importPath, got, test.want)
			}
		})
	}
}
