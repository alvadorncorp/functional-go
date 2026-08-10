package archtest

import (
	"os"
	"slices"
	"strings"
	"testing"
)

const (
	pkgFunction  = "github.com/alvadorncorp/gopackages/functional/function"
	pkgPredicate = "github.com/alvadorncorp/gopackages/functional/predicate"
	pkgOrder     = "github.com/alvadorncorp/gopackages/functional/order"
	pkgFold      = "github.com/alvadorncorp/gopackages/functional/fold"
)

// primitivePackages is the growing inventory of root-level functional
// primitives. Each entry has a phase-specific dependency and import-safety
// contract below.
var primitivePackages = []string{pkgFunction, pkgPredicate, pkgOrder, pkgFold, pkgTuple, pkgLazy}

func TestPrimitiveDependencyDAG(t *testing.T) {
	t.Parallel()
	for _, pkg := range primitivePackages {
		pkg := pkg
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			allowed := publicAllowedEdges[pkg]
			for _, dep := range deps(t, pkg) {
				if dep == pkg || !strings.HasPrefix(dep, functionalModulePrefix) {
					continue
				}
				if !slices.Contains(allowed, dep) {
					t.Errorf("%s must not depend on %s (allowed: %v)", pkg, dep, allowed)
				}
			}
		})
	}
}

func TestPrimitivesNoReflectOrUnsafe(t *testing.T) {
	t.Parallel()
	for _, pkg := range primitivePackages {
		pkg := pkg
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, imp := range productionImports(t, pkg) {
				if imp == "reflect" || imp == "unsafe" {
					t.Errorf("%s must not import %s", pkg, imp)
				}
			}
		})
	}
}

// TestGoModHasNoRequirements catches an accidental `go get` at the source:
// TestNoNonStdlibDependencies only sees a stray dependency once some file
// actually imports it, while go.mod records it the moment it is added.
func TestGoModHasNoRequirements(t *testing.T) {
	t.Parallel()

	path := goModPath(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") || strings.HasPrefix(line, "replace") {
			t.Errorf("%s must not have a require/replace line, found: %q", path, line)
		}
	}
}

// TestCoreNoReflect proves no production file under internal/core imports
// "reflect". Tests may still use the testing package; this gate inspects
// production Imports only (go list without -test).
func TestCoreNoReflect(t *testing.T) {
	t.Parallel()
	for _, pkg := range corePackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, imp := range productionImports(t, pkg) {
				if imp == "reflect" {
					t.Errorf("%s must not import reflect", pkg)
				}
			}
		})
	}
}

// TestEqNoReflect proves no production file under internal/eq imports
// "reflect". Equality is resolved via Equal methods and interface == only.
func TestEqNoReflect(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgEq) {
		if imp == "reflect" {
			t.Errorf("%s must not import reflect", pkgEq)
		}
	}
}

// TestEqNoUnsafe proves no production file under internal/eq imports
// "unsafe".
func TestEqNoUnsafe(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgEq) {
		if imp == "unsafe" {
			t.Errorf("%s must not import unsafe", pkgEq)
		}
	}
}

// TestPersistentNoReflect proves no production file under persistent imports
// "reflect". Public collection APIs stay free of reflection; tests may still
// use reflect (e.g. comparability probes). This gate inspects production
// Imports only (go list without -test).
func TestPersistentNoReflect(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgPersistent) {
		if imp == "reflect" {
			t.Errorf("%s must not import reflect", pkgPersistent)
		}
	}
}

// TestPersistentNoUnsafe proves no production file under persistent imports
// "unsafe".
func TestPersistentNoUnsafe(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgPersistent) {
		if imp == "unsafe" {
			t.Errorf("%s must not import unsafe", pkgPersistent)
		}
	}
}

// TestStateNoReflect proves no production file under state imports
// "reflect". The semantic layer stays free of reflection; tests may still
// use reflect. This gate inspects production Imports only (go list without
// -test).
func TestStateNoReflect(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgState) {
		if imp == "reflect" {
			t.Errorf("%s must not import reflect", pkgState)
		}
	}
}

// TestStateNoUnsafe proves no production file under state imports "unsafe".
func TestStateNoUnsafe(t *testing.T) {
	t.Parallel()
	for _, imp := range productionImports(t, pkgState) {
		if imp == "unsafe" {
			t.Errorf("%s must not import unsafe", pkgState)
		}
	}
}
