// Package archtest proves dependency invariants of the functional module
// mechanically, so they stay true rather than merely documented: the module
// has zero non-stdlib dependencies outside itself, public subpackages
// (including optics) have exactly four permitted sibling edges
// (values/validated to values/nonempty, fold to values/optional, fold to
// values/nonempty, and iterator to tuple; functional/internal/eq is the one
// shared internal exception every public package may use), the internal/core
// DAG stays a star into ownership with no peer edges and no reflect,
// internal/eq production code imports neither reflect nor unsafe, immutable
// and state production code import neither reflect nor unsafe, and
// internal/engine packages obey their own import allow-lists (keyhash: stdlib
// only; hamt: ownership/cow/keyhash only among functional packages; vector:
// ownership/cow only among functional packages; wbtree: ownership/cow only
// among functional packages).
package archtest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

// functionalModulePrefix identifies the root module's package namespace for
// per-package edge checks. Module identity is checked separately below.
const functionalModulePrefix = "go.alvadorn.org/functional/"

const (
	pkgOwnership       = "go.alvadorn.org/functional/internal/core/ownership"
	pkgCow             = "go.alvadorn.org/functional/internal/core/cow"
	pkgTransient       = "go.alvadorn.org/functional/internal/core/transient"
	pkgEq              = "go.alvadorn.org/functional/internal/eq"
	pkgKeyhash         = "go.alvadorn.org/functional/internal/engine/keyhash"
	pkgHamt            = "go.alvadorn.org/functional/internal/engine/hamt"
	pkgVector          = "go.alvadorn.org/functional/internal/engine/vector"
	pkgWbtree          = "go.alvadorn.org/functional/internal/engine/wbtree"
	pkgTuple           = "go.alvadorn.org/functional/structural/tuple"
	pkgLazy            = "go.alvadorn.org/functional/structural/lazy"
	pkgIterator        = "go.alvadorn.org/functional/structural/iterator"
	pkgOptics          = "go.alvadorn.org/functional/structural/optics"
	pkgImmutable       = "go.alvadorn.org/functional/immutable"
	pkgState           = "go.alvadorn.org/functional/state"
	pkgValuesOptional  = "go.alvadorn.org/functional/values/optional"
	pkgValuesResult    = "go.alvadorn.org/functional/values/result"
	pkgValuesEither    = "go.alvadorn.org/functional/values/either"
	pkgValuesNonEmpty  = "go.alvadorn.org/functional/values/nonempty"
	pkgValuesValidated = "go.alvadorn.org/functional/values/validated"
	corePrefix         = "go.alvadorn.org/functional/internal/core/"
	enginePrefix       = "go.alvadorn.org/functional/internal/engine/"
)

// publicPackages are every public subpackage of functional, by import path.
// Public sibling edges must be explicitly allow-listed below.
var publicPackages = []string{
	pkgFunction,
	pkgPredicate,
	pkgOrder,
	pkgFold,
	pkgTuple,
	pkgLazy,
	pkgIterator,
	pkgOptics,
	"go.alvadorn.org/functional/immutable",
	"go.alvadorn.org/functional/state",
	pkgValuesOptional,
	pkgValuesResult,
	pkgValuesEither,
	pkgValuesNonEmpty,
	pkgValuesValidated,
}

var valuesPackages = []string{
	pkgValuesOptional,
	pkgValuesResult,
	pkgValuesEither,
	pkgValuesNonEmpty,
	pkgValuesValidated,
}

var publicAllowedEdges = map[string][]string{
	pkgValuesValidated: {pkgValuesNonEmpty},
	pkgFold:            {pkgValuesOptional, pkgValuesNonEmpty},
	pkgIterator:        {pkgTuple},
}

// publicPackagesNoCore are the public packages that must never import
// internal/core. immutable is deliberately omitted: it may depend on core
// later for the engine contract.
var publicPackagesNoCore = []string{
	pkgFunction,
	pkgPredicate,
	pkgOrder,
	pkgFold,
	pkgTuple,
	pkgLazy,
	pkgIterator,
	pkgOptics,
	"go.alvadorn.org/functional/state",
	pkgValuesOptional,
	pkgValuesResult,
	pkgValuesEither,
	pkgValuesNonEmpty,
	pkgValuesValidated,
}

// corePackages is every production package under internal/core.
var corePackages = []string{
	pkgOwnership,
	pkgCow,
	pkgTransient,
}

// enginePackages is every production package under internal/engine.
var enginePackages = []string{
	pkgKeyhash,
	pkgHamt,
	pkgVector,
	pkgWbtree,
}

// coreAllowedEdges is the only permitted non-self edge set among core
// packages: a star into ownership. Peer edges and reverse edges are forbidden.
var coreAllowedEdges = map[string][]string{
	pkgOwnership: {},
	pkgCow:       {pkgOwnership},
	pkgTransient: {pkgOwnership},
}

// coreForbiddenTargets are non-core functional packages no core package may
// import (today or later without an explicit archtest change).
var coreForbiddenTargets = []string{
	pkgEq,
	pkgKeyhash,
	pkgHamt,
	pkgVector,
	pkgWbtree,
	"go.alvadorn.org/functional/immutable",
	"go.alvadorn.org/functional/state",
}

// hamtAllowedFunctional is the only permitted set of functional-module
// imports for the hamt engine (among non-self deps). keyhash is stdlib-only
// and is checked separately.
var hamtAllowedFunctional = []string{
	pkgOwnership,
	pkgCow,
	pkgKeyhash,
}

// vectorAllowedFunctional is the only permitted set of functional-module
// imports for the vector engine (among non-self deps). vector must not
// import hamt, keyhash, or transient.
var vectorAllowedFunctional = []string{
	pkgOwnership,
	pkgCow,
}

// wbtreeAllowedFunctional is the only permitted set of functional-module
// imports for the wbtree engine (among non-self deps). wbtree must not import
// keyhash, hamt, vector, or transient.
var wbtreeAllowedFunctional = []string{
	pkgOwnership,
	pkgCow,
}

// engineForbiddenTargets are functional packages no engine may import,
// regardless of other allow-lists (transient is core but not for engines;
// public packages and eq stay above the engine layer).
var engineForbiddenTargets = []string{
	pkgTransient,
	pkgEq,
	"go.alvadorn.org/functional/immutable",
	"go.alvadorn.org/functional/state",
}

// TestNoNonStdlibDependencies proves the module has zero dependencies
// outside the standard library and its own subpackages, permanently — the
// promise doc.go makes for the whole module.
func TestNoNonStdlibDependencies(t *testing.T) {
	t.Parallel()
	const ownModule = "go.alvadorn.org/functional"
	command := exec.Command("go", "list", "-deps", "-f", "{{if not .Standard}}{{.Module.Path}}{{end}}", "./...")
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("list functional dependency modules: %v\n%s", err, output)
	}
	for _, path := range strings.Fields(string(output)) {
		if path != ownModule {
			t.Errorf("functional must not depend on module %s", path)
		}
	}
}

func TestPrimitivesDoNotDependOnModeling(t *testing.T) {
	t.Parallel()
	const modelingPath = functionalModulePrefix + "modeling"
	command := exec.Command("go", "list", "-f", "{{.ImportPath}} {{join .Deps \" \"}}", "./...")
	command.Dir = "../.."
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("list functional package dependencies: %v\n%s", err, output)
	}
	isModeling := func(path string) bool {
		return path == modelingPath || strings.HasPrefix(path, modelingPath+"/")
	}
	for line := range strings.Lines(string(output)) {
		packages := strings.Fields(line)
		if len(packages) == 0 || isModeling(packages[0]) {
			continue
		}
		for _, dependency := range packages[1:] {
			if isModeling(dependency) {
				t.Errorf("%s must not depend on %s", packages[0], dependency)
			}
		}
	}
}

// TestPublicDependencyDAG proves the exact public-package dependency DAG as a
// tested invariant, not just a documented one. It is what
// guarantees, for example, that immutable.Map.Get returns comma-ok instead
// of a values/optional.Optional by design, not by oversight: if it ever
// imported values/optional to do so, this test would fail.
//
// Scope, by design: this only ever sees PRODUCTION imports, because deps
// below runs `go list -deps -json` WITHOUT `-test`, which never inspects
// `_test.go` files. That is deliberate, not a blind spot in this test: a
// public subpackage's test code MAY legitimately import a sibling public
// package to exercise or demonstrate cross-package behavior — for example
// state tests may compose with immutable.Map/Set/List fields. That is a
// real, intentional dependency between test binaries, and it does not
// violate the architectural boundary this test protects (production code
// never importing a public sibling). Adding `-test` here would make that
// legitimate test import fail this test; do not do that. Production sibling
// edges are explicitly allow-listed; every other public sibling edge is
// forbidden. See docs/overview.md for the full rationale.
func TestPublicDependencyDAG(t *testing.T) {
	t.Parallel()
	for _, pkg := range publicPackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			allowed := publicAllowedEdges[pkg]
			for _, dep := range deps(t, pkg) {
				if dep == pkg || !slices.Contains(publicPackages, dep) {
					continue
				}
				if !slices.Contains(allowed, dep) {
					t.Errorf("%s must not depend on public sibling %s (allowed: %v)", pkg, dep, allowed)
				}
			}
		})
	}
}

func TestValuesDependencyDAG(t *testing.T) {
	t.Parallel()
	for _, pkg := range valuesPackages {
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

// TestValuesNoReflectOrUnsafe proves each values package's production code
// stays free of direct reflection and unsafe imports.
func TestValuesNoReflectOrUnsafe(t *testing.T) {
	t.Parallel()
	for _, pkg := range valuesPackages {
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

// TestCoreDependencyDAG proves the core package graph is a star into
// ownership: cow and transient may each depend on ownership, and nothing else
// in functional. Peer edges, reverse edges into ownership, and edges into
// eq/public packages are forbidden.
func TestCoreDependencyDAG(t *testing.T) {
	t.Parallel()
	for _, pkg := range corePackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			allowed := coreAllowedEdges[pkg]
			for _, dep := range deps(t, pkg) {
				if dep == pkg {
					continue
				}
				if !strings.HasPrefix(dep, functionalModulePrefix) {
					continue // stdlib / outside handled by TestNoNonStdlibDependencies
				}
				if slices.Contains(coreForbiddenTargets, dep) {
					t.Errorf("%s must not depend on %s", pkg, dep)
					continue
				}
				if strings.HasPrefix(dep, corePrefix) || dep == strings.TrimSuffix(corePrefix, "/") {
					if !slices.Contains(allowed, dep) {
						t.Errorf("%s must not depend on core package %s (allowed: %v)", pkg, dep, allowed)
					}
					continue
				}
				// Any other functional package is also forbidden for core.
				t.Errorf("%s must not depend on %s", pkg, dep)
			}
		})
	}
}

// TestPublicPackagesDoNotImportCore proves public packages other than
// immutable never reach into internal/core. immutable is allowed to import
// core (engine contract) and is not checked here.
func TestPublicPackagesDoNotImportCore(t *testing.T) {
	t.Parallel()
	for _, pkg := range publicPackagesNoCore {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, dep := range deps(t, pkg) {
				if strings.HasPrefix(dep, corePrefix) {
					t.Errorf("%s must not depend on core package %s", pkg, dep)
				}
			}
		})
	}
}

// TestEngineNoReflect proves no production file under internal/engine
// imports "reflect". Tests may still use the testing package; this gate
// inspects production Imports only (go list without -test).
func TestEngineNoReflect(t *testing.T) {
	t.Parallel()
	for _, pkg := range enginePackages {
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

// TestEngineDependencies proves engine import allow-lists:
//   - keyhash: standard library only (no functional packages)
//   - hamt: among functional packages, only ownership, cow, keyhash
//   - vector: among functional packages, only ownership, cow
//   - wbtree: among functional packages, only ownership, cow
//   - no engine may import transient, eq, immutable, or state
//   - hamt ↛ vector and vector ↛ hamt (peer engines stay disconnected);
//     wbtree's allow-list excludes every peer, and wbtree is outside the
//     hamt and vector allow-lists
//
// Core packages already cannot import engines via TestCoreDependencyDAG
// (any non-allowed functional edge fails); engines are also listed in
// coreForbiddenTargets for explicit failure messages.
func TestEngineDependencies(t *testing.T) {
	t.Parallel()
	for _, pkg := range enginePackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, dep := range deps(t, pkg) {
				if dep == pkg {
					continue
				}
				if !strings.HasPrefix(dep, functionalModulePrefix) {
					continue // stdlib / outside handled by TestNoNonStdlibDependencies
				}
				if slices.Contains(engineForbiddenTargets, dep) {
					t.Errorf("%s must not depend on %s", pkg, dep)
					continue
				}
				// Peer engines must not import each other.
				if pkg == pkgHamt && dep == pkgVector {
					t.Errorf("%s must not depend on %s", pkg, dep)
					continue
				}
				if pkg == pkgVector && dep == pkgHamt {
					t.Errorf("%s must not depend on %s", pkg, dep)
					continue
				}
				// keyhash stays stdlib-only and must not reach either peer engine.
				if pkg == pkgKeyhash && (dep == pkgVector || dep == pkgHamt) {
					t.Errorf("%s must not depend on %s", pkg, dep)
					continue
				}
				switch pkg {
				case pkgKeyhash:
					t.Errorf("%s must depend on stdlib only; found functional package %s", pkg, dep)
				case pkgHamt:
					if !slices.Contains(hamtAllowedFunctional, dep) {
						t.Errorf("%s must not depend on %s (allowed functional: %v)", pkg, dep, hamtAllowedFunctional)
					}
				case pkgVector:
					if !slices.Contains(vectorAllowedFunctional, dep) {
						t.Errorf("%s must not depend on %s (allowed functional: %v)", pkg, dep, vectorAllowedFunctional)
					}
				case pkgWbtree:
					if !slices.Contains(wbtreeAllowedFunctional, dep) {
						t.Errorf("%s must not depend on %s (allowed functional: %v)", pkg, dep, wbtreeAllowedFunctional)
					}
				default:
					t.Errorf("unhandled engine package %s depending on %s", pkg, dep)
				}
			}
		})
	}
}

// TestCoreDoesNotImportEngine is an explicit star-boundary check: no core
// package may reach into internal/engine. TestCoreDependencyDAG already
// rejects any non-allowed functional edge; this names the engine layer so
// a regression is obvious without reading the generic rule.
func TestCoreDoesNotImportEngine(t *testing.T) {
	t.Parallel()
	for _, pkg := range corePackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, dep := range deps(t, pkg) {
				if strings.HasPrefix(dep, enginePrefix) {
					t.Errorf("%s must not depend on engine package %s", pkg, dep)
				}
			}
		})
	}
}

// goModPath resolves the functional module's go.mod via `go env GOMOD`. The
// command inherits this test binary's working directory, which `go test`
// sets to this package's own directory — inside the functional module — so
// the module resolved is always functional's, not the workspace's.
func goModPath(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	path := strings.TrimSpace(string(out))
	if path == "" || path == os.DevNull {
		t.Fatalf("go env GOMOD found no module for this directory: %q", path)
	}
	return path
}

// deps returns every non-stdlib dependency of pkg (including pkg itself),
// via `go list -deps -json`, decoded in stream.
func deps(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-deps", "-json", pkg).Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	var found []string
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var e struct {
			ImportPath string
			Standard   bool
		}
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if !e.Standard {
			found = append(found, e.ImportPath)
		}
	}
	return found
}

// productionImports returns the direct production Imports of pkg (no test
// files), via `go list -json`. Includes stdlib paths such as "reflect".
func productionImports(t *testing.T, pkg string) []string {
	t.Helper()
	out, err := exec.Command("go", "list", "-json", pkg).Output()
	if err != nil {
		t.Fatalf("go list -json %s: %v", pkg, err)
	}
	var e struct {
		Imports []string
	}
	if err := json.Unmarshal(out, &e); err != nil {
		t.Fatalf("decode go list -json %s: %v", pkg, err)
	}
	return e.Imports
}
