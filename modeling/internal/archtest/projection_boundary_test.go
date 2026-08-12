package archtest

import (
	"slices"
	"strings"
	"testing"
)

func projectionProductionPackage(t *testing.T) *listedPackage {
	t.Helper()
	productions := productionPackages(t)
	for i := range productions {
		if productions[i].ImportPath == projectionPackage {
			return &productions[i]
		}
	}
	t.Fatalf("package enumeration is missing %s", projectionPackage)
	return nil
}

func TestProjectionHasExactPublicSurface(t *testing.T) {
	pkg := projectionProductionPackage(t)
	got := productionNames(t, pkg)
	slices.Sort(got)
	want := []string{"Project", "Projection"}
	if !slices.Equal(got, want) {
		t.Fatalf("projection exports = %v, want %v", got, want)
	}
}

func TestProjectionHasExactFileEnvelope(t *testing.T) {
	pkg := projectionProductionPackage(t)
	production := slices.Clone(pkg.GoFiles)
	internalTests := slices.Clone(pkg.TestGoFiles)
	externalTests := slices.Clone(pkg.XTestGoFiles)
	slices.Sort(production)
	slices.Sort(internalTests)
	slices.Sort(externalTests)

	wantProduction := []string{"project.go", "projection.go"}
	wantExternalTests := []string{"project_test.go", "projection_test.go"}
	if !slices.Equal(production, wantProduction) {
		t.Fatalf("projection production files = %v, want %v", production, wantProduction)
	}
	if len(internalTests) != 0 {
		t.Fatalf("projection internal test files = %v, want none", internalTests)
	}
	if !slices.Equal(externalTests, wantExternalTests) {
		t.Fatalf("projection external test files = %v, want %v", externalTests, wantExternalTests)
	}
}

func TestProjectionHasExactDependencyBoundary(t *testing.T) {
	pkg := projectionProductionPackage(t)
	imports := slices.Clone(pkg.Imports)
	slices.Sort(imports)
	want := []string{"iter"}
	if !slices.Equal(imports, want) {
		t.Fatalf("projection imports = %v, want exactly %v", imports, want)
	}
	for _, imported := range imports {
		if imported == modelingModule || strings.HasPrefix(imported, modelingModule+"/") {
			t.Errorf("%s imports forbidden modeling capability %s", projectionPackage, imported)
		}
		if imported == functionalImmutable || strings.HasPrefix(imported, functionalModule) {
			t.Errorf("%s imports forbidden functional package %s", projectionPackage, imported)
		}
	}
}

func TestProjectionNamesAreProjectionOnly(t *testing.T) {
	wantNames := []string{"Project", "Projection"}
	for _, production := range productionPackages(t) {
		if production.ImportPath == projectionPackage {
			continue
		}
		names := productionNames(t, &production)
		for _, name := range wantNames {
			if slices.Contains(names, name) {
				t.Errorf("%s exports projection-only API %s", production.ImportPath, name)
			}
		}
	}
}
