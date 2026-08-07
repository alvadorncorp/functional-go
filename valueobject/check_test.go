package valueobject_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alvadorncorp/gopackages/functional/persistent"
	"github.com/alvadorncorp/gopackages/functional/valueobject"
)

// Good is a well-formed value object: every field unexported, one plain
// string, and one persistent.List[string] — proving that functional/persistent
// is allowed by default without valueobject importing it.
type Good struct {
	name string
	tags persistent.List[string]
}

// Bad has one exported field and one unallowed slice field: two distinct,
// independently reported violations.
type Bad struct {
	Name string
	tags []string
}

// Inner and Outer prove recursion into a nested unexported struct field,
// with the violation's Path carrying every level of the descent.
type Inner struct {
	items []string
}

type Outer struct {
	inner Inner
}

// FreelyMutable would violate by default; AllowType[FreelyMutable]() frees
// it explicitly.
type FreelyMutable struct {
	items []string
}

type HasAllowedField struct {
	allowed FreelyMutable
}

// Handler is an interface field: its dynamic type can never be proven
// immutable by structural inspection alone.
type Handler interface{ Handle() }

type HasInterface struct {
	handler Handler
}

// HasTimes proves time.Time and time.Duration are immutable leaves without
// any Option.
type HasTimes struct {
	when time.Time
	ttl  time.Duration
}

// MutableBox and HasBox exist only to prove AllowPackage frees an entire
// package by import-path string, independent of the persistent default.
type MutableBox struct {
	data []string
}

type HasBox struct {
	box MutableBox
}

// arrayFixture reproduces achado 1: Inspect previously ignored
// reflect.Array entirely, so a [3]*int field passed Check with zero
// violations even though every *int obtained by indexing into ids remains
// mutable from outside the value object, and the field is not exported
// either.
type arrayFixture struct{ ids [3]*int }

// ArrayOfInterface has an array of an interface element: the same
// non-provable-dynamic-type problem as a bare interface field (Handler,
// declared above), just reached by indexing instead of direct field
// access.
type ArrayOfInterface struct {
	handlers [2]Handler
}

// ArrayOfStruct has an array of a nested unexported struct (Inner, declared
// above) that itself violates, proving Inspect recurses into the array's
// ELEMENT TYPE once — not once per array position — and marks the crossing
// with "[]" in Path.
type ArrayOfStruct struct {
	items [3]Inner
}

// ArrayOfScalar has an array of a plain scalar: immutable by value
// semantics regardless of length, so it must never violate.
type ArrayOfScalar struct {
	codes [4]int
}

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		check   func() error
		wantErr bool
	}{
		{"good value object has no violation", func() error { return valueobject.Check[Good]() }, false},
		{"exported and reference fields violate", func() error { return valueobject.Check[Bad]() }, true},
		{"violation nested inside an unexported struct field", func() error { return valueobject.Check[Outer]() }, true},
		{
			"AllowType frees a type that would otherwise violate", func() error {
				return valueobject.Check[HasAllowedField](valueobject.AllowType[FreelyMutable]())
			}, false,
		},
		{"interface field violates", func() error { return valueobject.Check[HasInterface]() }, true},
		{"time.Time and time.Duration are leaves by default", func() error { return valueobject.Check[HasTimes]() }, false},
		{"non-struct T is treated as a leaf", func() error { return valueobject.Check[string]() }, false},
		{"array of pointers violates", func() error { return valueobject.Check[arrayFixture]() }, true},
		{"array of interface violates", func() error { return valueobject.Check[ArrayOfInterface]() }, true},
		{"array of nested struct violates via recursion", func() error { return valueobject.Check[ArrayOfStruct]() }, true},
		{"array of scalar does not violate", func() error { return valueobject.Check[ArrayOfScalar]() }, false},
		{
			"AllowType frees an array field's exact type", func() error {
				return valueobject.Check[arrayFixture](valueobject.AllowType[[3]*int]())
			}, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.check()
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInspectBadViolations(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[Bad]()
	want := []valueobject.Violation{
		{Path: "Bad.Name", Type: "string", Rule: valueobject.RuleExportedField},
		{Path: "Bad.tags", Type: "[]string", Rule: valueobject.RuleReferenceField},
	}

	if len(report.Violations) != len(want) {
		t.Fatalf("len(Violations) = %d, want %d (got %+v)", len(report.Violations), len(want), report.Violations)
	}
	for i, w := range want {
		got := report.Violations[i]
		if got.Path != w.Path || got.Type != w.Type || got.Rule != w.Rule {
			t.Errorf("Violations[%d] = %+v, want Path=%q Type=%q Rule=%v", i, got, w.Path, w.Type, w.Rule)
		}
		if got.Reason == "" {
			t.Errorf("Violations[%d].Reason is empty", i)
		}
	}
}

func TestReportErrFormat(t *testing.T) {
	t.Parallel()

	err := valueobject.Check[Bad]()
	if err == nil {
		t.Fatal("Check[Bad]() = nil, want a non-nil error")
	}

	lines := strings.Split(err.Error(), "\n")
	if len(lines) != 2 {
		t.Fatalf("error has %d lines, want 2: %q", len(lines), err)
	}
	for _, line := range lines {
		if !strings.Contains(line, ": ") || !strings.HasSuffix(line, ")") {
			t.Errorf("line %q does not match '<path>: <reason> (<type>)'", line)
		}
	}
}

func TestInspectNestedStructPath(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[Outer]()
	if len(report.Violations) != 1 {
		t.Fatalf("len(Violations) = %d, want 1 (got %+v)", len(report.Violations), report.Violations)
	}

	got := report.Violations[0]
	if got.Path != "Outer.inner.items" {
		t.Errorf("Path = %q, want %q", got.Path, "Outer.inner.items")
	}
	if got.Rule != valueobject.RuleReferenceField {
		t.Errorf("Rule = %v, want RuleReferenceField", got.Rule)
	}
}

func TestInspectInterfaceFieldPath(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[HasInterface]()
	if len(report.Violations) != 1 {
		t.Fatalf("len(Violations) = %d, want 1 (got %+v)", len(report.Violations), report.Violations)
	}

	got := report.Violations[0]
	if got.Path != "HasInterface.handler" {
		t.Errorf("Path = %q, want %q", got.Path, "HasInterface.handler")
	}
	if got.Rule != valueobject.RuleInterfaceField {
		t.Errorf("Rule = %v, want RuleInterfaceField", got.Rule)
	}
}

// TestInspectArrayOfPointerViolation reproduces achado 1 directly:
// Inspect previously ignored reflect.Array, so arrayFixture passed Check
// with zero violations despite ids[i] being a mutable *int reachable by
// indexing.
func TestInspectArrayOfPointerViolation(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[arrayFixture]()
	if len(report.Violations) != 1 {
		t.Fatalf("len(Violations) = %d, want 1 (got %+v)", len(report.Violations), report.Violations)
	}

	got := report.Violations[0]
	if got.Path != "arrayFixture.ids" {
		t.Errorf("Path = %q, want %q", got.Path, "arrayFixture.ids")
	}
	if got.Type != "[3]*int" {
		t.Errorf("Type = %q, want %q", got.Type, "[3]*int")
	}
	if got.Rule != valueobject.RuleReferenceField {
		t.Errorf("Rule = %v, want RuleReferenceField", got.Rule)
	}
}

// TestInspectArrayOfInterfaceViolation proves an array of an interface
// element reports RuleInterfaceField, the same rule a bare interface field
// reports, just reached by indexing.
func TestInspectArrayOfInterfaceViolation(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[ArrayOfInterface]()
	if len(report.Violations) != 1 {
		t.Fatalf("len(Violations) = %d, want 1 (got %+v)", len(report.Violations), report.Violations)
	}

	got := report.Violations[0]
	if got.Path != "ArrayOfInterface.handlers" {
		t.Errorf("Path = %q, want %q", got.Path, "ArrayOfInterface.handlers")
	}
	if got.Rule != valueobject.RuleInterfaceField {
		t.Errorf("Rule = %v, want RuleInterfaceField", got.Rule)
	}
}

// TestInspectArrayOfStructPath proves Inspect recurses into an array
// field's element TYPE once — not once per array position — and marks the
// crossing with "[]" in Path.
func TestInspectArrayOfStructPath(t *testing.T) {
	t.Parallel()

	report := valueobject.Inspect[ArrayOfStruct]()
	if len(report.Violations) != 1 {
		t.Fatalf("len(Violations) = %d, want 1 (got %+v)", len(report.Violations), report.Violations)
	}

	got := report.Violations[0]
	if got.Path != "ArrayOfStruct.items[].items" {
		t.Errorf("Path = %q, want %q", got.Path, "ArrayOfStruct.items[].items")
	}
	if got.Rule != valueobject.RuleReferenceField {
		t.Errorf("Rule = %v, want RuleReferenceField", got.Rule)
	}
}

// PersistentPointerElement documents Inspect's known structural gap:
// persistent types are treated as immutable leaves by PkgPath alone,
// without regard to their type parameter, because reflect exposes no API
// to inspect a generic instantiation's type arguments. See Inspect's doc
// comment for the full explanation.
type PersistentPointerElement struct {
	items persistent.List[*int]
}

// TestInspectKnownGap_PersistentElementMutability pins a documented, permanent
// limitation of Inspect (see its doc comment): the persistent leaf check
// ignores the element type parameter entirely. A persistent.List[*int]
// field therefore passes Check with zero violations even though the *int
// elements obtained through Values()/Items() remain mutable through the
// pointer from outside the value object — the same underlying problem that
// IS correctly caught for optional.Optional[*Widget] (RuleReferenceField),
// because Optional holds its type parameter inline instead of behind a
// leaf-by-package check.
//
// This is NOT a bug: reflect has no API to inspect a generic
// instantiation's type arguments (no reflect.Type.TypeArgs()), so Inspect
// cannot discover that List was instantiated with *int and recurse into
// it. This test exists so the asymmetry never silently regresses further.
func TestInspectKnownGap_PersistentElementMutability(t *testing.T) {
	t.Parallel()

	if err := valueobject.Check[PersistentPointerElement](); err != nil {
		t.Fatalf("Check[PersistentPointerElement]() = %v, want nil (documented gap: "+
			"the persistent leaf check ignores the element type parameter)", err)
	}
}

func TestAllowPackageFreesWholePackage(t *testing.T) {
	t.Parallel()

	if err := valueobject.Check[HasBox](); err == nil {
		t.Fatal("Check[HasBox]() without AllowPackage = nil, want a violation")
	}

	pkgPath := reflect.TypeOf(MutableBox{}).PkgPath()
	err := valueobject.Check[HasBox](valueobject.AllowPackage(pkgPath))
	if err != nil {
		t.Fatalf("Check[HasBox](AllowPackage(%q)) = %v, want nil", pkgPath, err)
	}
}

// TestAllowTypeFreesInterfaceType reproduces achado 2: AllowType[T]() used
// to build its allow-list entry from reflect.TypeOf(zero), and for T an
// interface type, the zero value is a nil interface, so
// reflect.TypeOf(zero) is nil and the registration was silently dropped.
// AllowType[Handler]() therefore had no effect at all, and
// Check[HasInterface](AllowType[Handler]()) kept reporting the interface
// violation as if AllowType had never been called.
func TestAllowTypeFreesInterfaceType(t *testing.T) {
	t.Parallel()

	if err := valueobject.Check[HasInterface](); err == nil {
		t.Fatal("Check[HasInterface]() without AllowType = nil, want a violation")
	}

	err := valueobject.Check[HasInterface](valueobject.AllowType[Handler]())
	if err != nil {
		t.Fatalf("Check[HasInterface](AllowType[Handler]()) = %v, want nil", err)
	}
}

// genericWrapper is a minimal generic type instantiated ONLY from this
// external _test package, used solely to probe whether routing T through
// AllowType's reflect.TypeFor[T]() (a library call, not a direct
// reflect.TypeFor[Generic[T]]() in a _test.go file) trips the go1.26.4
// linker dead-code-elimination bug documented next to Inspect's own
// reflect.TypeOf choice. If this test compiles, links, and passes, the bug
// does not reach through this call path, and AllowType safely uses
// reflect.TypeFor[T]().
type genericWrapper[T any] struct{ v T }

// TestAllowTypeGenericInstantiationLinkerRisk is the achado 2 risk probe:
// AllowType[genericWrapper[string]]() instantiates a generic type only
// known to this external test package and routes it through AllowType's
// reflect.TypeFor[T](). A linker panic here (rather than a normal pass or
// fail) would mean the known go1.26.4 reflect.TypeFor bug reaches through a
// library function boundary, not just a direct test-file call — and would
// mean AllowType must NOT use reflect.TypeFor.
func TestAllowTypeGenericInstantiationLinkerRisk(t *testing.T) {
	t.Parallel()

	type hasWrapper struct {
		w genericWrapper[string]
	}

	err := valueobject.Check[hasWrapper](valueobject.AllowType[genericWrapper[string]]())
	if err != nil {
		t.Fatalf("Check[hasWrapper](AllowType[genericWrapper[string]]()) = %v, want nil", err)
	}
}
