// Package archtest proves two dependency invariants of the functional
// module mechanically, so they stay true rather than merely documented: the
// module has zero non-stdlib dependencies outside itself, and no public
// subpackage imports another (functional/internal/eq is the one shared
// exception every public package may use).
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

// functionalModulePrefix is the import-path prefix every non-stdlib
// dependency of a functional subpackage must fall under: the module cannot
// depend on anything outside itself.
const functionalModulePrefix = "github.com/alvadorncorp/gopackages/functional/"

// publicPackages are every public subpackage of functional, by import path.
// None of these may depend on another.
var publicPackages = []string{
	"github.com/alvadorncorp/gopackages/functional/optional",
	"github.com/alvadorncorp/gopackages/functional/frozen",
	"github.com/alvadorncorp/gopackages/functional/persistent",
	"github.com/alvadorncorp/gopackages/functional/valueobject",
}

// allPackages is publicPackages plus the one internal package every public
// package is allowed to share: functional/internal/eq.
var allPackages = []string{
	"github.com/alvadorncorp/gopackages/functional/optional",
	"github.com/alvadorncorp/gopackages/functional/frozen",
	"github.com/alvadorncorp/gopackages/functional/persistent",
	"github.com/alvadorncorp/gopackages/functional/valueobject",
	"github.com/alvadorncorp/gopackages/functional/internal/eq",
}

// TestNoNonStdlibDependencies proves the module has zero dependencies
// outside the standard library and its own subpackages, permanently — the
// promise doc.go makes for the whole module.
func TestNoNonStdlibDependencies(t *testing.T) {
	t.Parallel()
	for _, pkg := range allPackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, dep := range deps(t, pkg) {
				if !strings.HasPrefix(dep, functionalModulePrefix) {
					t.Errorf("%s must not depend on %s", pkg, dep)
				}
			}
		})
	}
}

// TestNoPublicSiblingDependencies proves "no public subpackage imports
// another" as a tested invariant, not just a documented one. It is what
// guarantees, for example, that frozen.Map.Get returns comma-ok instead of
// optional.Optional by design, not by oversight: if it ever imported
// optional to do so, this test would fail.
//
// Scope, by design: this only ever sees PRODUCTION imports, because deps
// below runs `go list -deps -json` WITHOUT `-test`, which never inspects
// `_test.go` files. That is deliberate, not a blind spot in this test: a
// public subpackage's test code MAY legitimately import a sibling public
// package to exercise or demonstrate cross-package behavior — e.g.
// valueobject/check_test.go imports functional/frozen to prove
// AllowPackage's default leaf list actually frees frozen fields, not merely
// that the code compiles. That is a real, intentional dependency between
// test binaries, and it does not violate the architectural boundary this
// test protects (production code never importing a public sibling). Adding
// `-test` here would make that legitimate test import fail this test; do
// not do that. See docs/overview.md, "No public subpackage imports another
// (in production code)", for the full rationale.
func TestNoPublicSiblingDependencies(t *testing.T) {
	t.Parallel()
	for _, pkg := range publicPackages {
		t.Run(pkg, func(t *testing.T) {
			t.Parallel()
			for _, dep := range deps(t, pkg) {
				if dep == pkg {
					continue // deps(t, pkg) includes pkg itself
				}
				if slices.Contains(publicPackages, dep) {
					t.Errorf("%s must not depend on public sibling %s", pkg, dep)
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
