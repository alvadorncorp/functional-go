package valueobject

import (
	"errors"
	"reflect"
	"strings"
	"time"
)

// Rule names the structural property a violation breaks.
type Rule uint8

const (
	RuleUnknown Rule = iota
	// RuleExportedField reports a field a caller can assign to directly.
	RuleExportedField
	// RuleReferenceField reports a pointer, slice, map, chan, or func field —
	// the value it refers to can be mutated through it even when the field
	// itself is unexported. An array field whose ELEMENT type is one of
	// these reports the same rule: indexing into the array exposes the same
	// reference regardless of the array field's own export status.
	RuleReferenceField
	// RuleInterfaceField reports an interface field, or an array field whose
	// element type is an interface, whose dynamic type cannot be proven
	// immutable by structural inspection.
	RuleInterfaceField
)

// Violation locates one broken rule. Path is a dotted field path from the
// inspected type, e.g. "Money.tags" or "Order.lines.items".
type Violation struct {
	Path   string
	Type   string // the field's Go type name, e.g. "[]string" or "*User"
	Rule   Rule
	Reason string
}

// Report is the result of Inspect: every Violation found, in the order
// fields were visited.
type Report struct{ Violations []Violation }

// OK reports whether Inspect found no violation.
func (r Report) OK() bool { return len(r.Violations) == 0 }

// Err returns nil when OK, otherwise a single error listing every
// violation, one per line, as "<path>: <reason> (<type>)".
func (r Report) Err() error {
	if r.OK() {
		return nil
	}
	lines := make([]string, len(r.Violations))
	for i, v := range r.Violations {
		lines[i] = v.Path + ": " + v.Reason + " (" + v.Type + ")"
	}
	return errors.New(strings.Join(lines, "\n"))
}

// config accumulates the leaf allow-list built from Option values.
type config struct {
	types    map[reflect.Type]struct{}
	packages []string
}

// Option configures Inspect's allow-list of types treated as immutable
// leaves. It is called Option, following the functional-options convention
// used elsewhere in this repository (eventbus.Option, the workopher engine
// and adapter packages) — a different concept from the sibling package
// functional/optional's Optional[T], which this package does not import.
type Option func(*config)

// AllowType marks T as an immutable leaf: Inspect stops descending into it
// and never reports it as a violation.
//
// This uses reflect.TypeFor[T](), unlike Inspect and isLeaf's default
// leaves, which go through reflect.TypeOf on a concrete value (see Inspect's
// doc comment). That asymmetry is deliberate, not an inconsistency: T here
// can be an interface type (e.g. AllowType[SomeInterface]()), and its zero
// value is a nil interface, so reflect.TypeOf(zero) would return nil and
// silently drop the registration — AllowType[Handler]() would have compiled
// and appeared to work but never actually freed Handler, exactly the
// symptom TestAllowTypeFreesInterfaceType in check_test.go pins as a
// regression test. reflect.TypeFor[T]() (Go 1.22+) returns T's STATIC type
// without needing a value, so it works correctly for an interface T; this
// is precisely the use case the standard library added it for.
//
// This toolchain (go1.26.4) has a real linker bug where
// reflect.TypeFor[Generic[T]]() called directly in an external _test.go
// file, over a generic type instantiated only by that test file, panics at
// link time (R_USEIFACE ... eqfunc ...) — see the reflect.TypeOf workaround
// notes in optional/optional_repr_test.go and persistent/repr_test.go, and
// Inspect's own doc comment above. Before applying reflect.TypeFor here,
// that risk was probed specifically for THIS call shape — T routed through
// a library function rather than reflected on directly in a test file — by
// TestAllowTypeGenericInstantiationLinkerRisk in check_test.go, which
// instantiates a generic type only in the external test package and passes
// it through AllowType[T](). That test compiles, links, and passes under
// both `go test` and `go test -race`, so the bug does not reach through
// this path, and reflect.TypeFor[T]() is safe to use in AllowType.
func AllowType[T any]() Option {
	return func(c *config) {
		c.types[reflect.TypeFor[T]()] = struct{}{}
	}
}

// AllowPackage marks every type declared in the given import path as an
// immutable leaf. The comparison is by STRING against the type's
// reflect.Type.PkgPath() — it creates no real import edge between
// valueobject and the named package. That string comparison is exactly how
// functional/frozen and functional/persistent are allowed by default
// without this package importing either: a real import would break "no
// public subpackage of functional imports another."
func AllowPackage(importPath string) Option {
	return func(c *config) {
		c.packages = append(c.packages, importPath)
	}
}

// defaultLeafPackages are treated as immutable leaves without any Option:
// both collections already guarantee their own copy on entry and exit, so
// structural inspection has nothing to add.
var defaultLeafPackages = []string{
	"github.com/alvadorncorp/gopackages/functional/frozen",
	"github.com/alvadorncorp/gopackages/functional/persistent",
}

// timeTimeType and timeDurationType are immutable leaves by default.
var (
	timeTimeType     = reflect.TypeOf(time.Time{})
	timeDurationType = reflect.TypeOf(time.Duration(0))
)

// isLeaf reports whether t must not be descended into or reported as a
// violation: a default leaf (time.Time, time.Duration, frozen, persistent),
// an AllowType match, or a type declared under an AllowPackage import path.
func (c *config) isLeaf(t reflect.Type) bool {
	if t == timeTimeType || t == timeDurationType {
		return true
	}
	if _, ok := c.types[t]; ok {
		return true
	}
	pkgPath := t.PkgPath()
	if pkgPath == "" {
		return false
	}
	for _, prefix := range defaultLeafPackages {
		if underPackage(pkgPath, prefix) {
			return true
		}
	}
	for _, prefix := range c.packages {
		if underPackage(pkgPath, prefix) {
			return true
		}
	}
	return false
}

// underPackage reports whether pkgPath is prefix itself or a subpackage of
// it, guarding against a naive strings.HasPrefix false positive such as
// "…/frozenish" matching prefix "…/frozen".
func underPackage(pkgPath, prefix string) bool {
	return pkgPath == prefix || strings.HasPrefix(pkgPath, prefix+"/")
}

// Inspect walks the structure of T and reports every field an external
// caller could use to mutate the value: an exported field, a pointer,
// slice, map, chan, or func alias — including one reached by indexing into
// an array field — or an interface (directly or as an array element) whose
// dynamic type cannot be proven immutable by structural inspection.
//
// T's reflect.Type comes from reflect.TypeOf(zero) on a concrete zero
// value, not reflect.TypeFor[T](). On this toolchain (go1.26.4),
// reflect.TypeFor over a type parameter instantiated only in an external
// test package triggers a real linker bug (panic: R_USEIFACE ... eqfunc
// ...); reflect.TypeOf on a concrete value reaches the same reflect.Type
// without going through that code path.
//
// Known gap — frozen/persistent leaves ignore their element type: a field
// whose type is declared in functional/frozen or functional/persistent is
// treated as an immutable leaf unconditionally (see isLeaf and
// defaultLeafPackages), by matching the field's own PkgPath, WITHOUT
// looking at that generic type's type parameter. That correctly captures
// that the COLLECTION itself cannot be mutated from outside — Values() and
// Items() never leak a reference to the internal slice or map — but it does
// NOT verify that the element type is itself immutable. A
// frozen.Slice[*Widget] field is treated as a leaf and reported with zero
// violations even though every *Widget obtained through Values()/Items()
// remains mutable through the pointer. This is a structural limitation of
// the reflect package, not an omission of this code: reflect exposes no API
// to inspect the type arguments of a generic instantiation (there is no
// reflect.Type.TypeArgs()), so Inspect has no way to discover that Slice
// was instantiated with *Widget and recurse into it. Contrast with
// optional.Optional[*Widget], which holds its type parameter inline rather
// than behind a leaf-by-package check, so Inspect DOES recurse into it and
// correctly reports RuleReferenceField. When a frozen/persistent element
// type is itself a pointer, slice, map, chan, func, or interface, combine
// AllowType or a manual review of that element type, or prefer
// optional.Optional[T] or a scalar field for that element instead.
//
// A non-struct T is treated as a leaf: this package exists to verify value
// objects, which are structs.
func Inspect[T any](opts ...Option) Report {
	c := &config{types: make(map[reflect.Type]struct{})}
	for _, opt := range opts {
		opt(c)
	}

	var zero T
	t := reflect.TypeOf(zero)
	if t == nil || t.Kind() != reflect.Struct {
		return Report{}
	}

	var violations []Violation
	c.inspectStruct(t, t.Name(), &violations)
	return Report{Violations: violations}
}

// inspectStruct visits every field of struct type t, appending a Violation
// for each field an external caller could use to mutate the value. path is
// the dotted path to t itself.
func (c *config) inspectStruct(t reflect.Type, path string, out *[]Violation) {
	for i := range t.NumField() {
		field := t.Field(i)
		fieldPath := path + "." + field.Name

		if field.IsExported() {
			*out = append(*out, Violation{
				Path:   fieldPath,
				Type:   field.Type.String(),
				Rule:   RuleExportedField,
				Reason: "field is exported: a caller can assign to it directly",
			})
			continue
		}

		c.inspectField(field.Type, fieldPath, out)
	}
}

// inspectField applies the rule for one unexported field's type, recursing
// into a struct field that is not a leaf, or into an array field's element
// type once.
func (c *config) inspectField(ft reflect.Type, fieldPath string, out *[]Violation) {
	switch ft.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		if !c.isLeaf(ft) {
			*out = append(*out, Violation{
				Path:   fieldPath,
				Type:   ft.String(),
				Rule:   RuleReferenceField,
				Reason: "field can be mutated through a pointer/slice/map/chan/func alias regardless of export status",
			})
		}
	case reflect.Interface:
		if !c.isLeaf(ft) {
			*out = append(*out, Violation{
				Path:   fieldPath,
				Type:   ft.String(),
				Rule:   RuleInterfaceField,
				Reason: "interface field's dynamic type cannot be proven immutable by structural inspection",
			})
		}
	case reflect.Struct:
		if !c.isLeaf(ft) {
			c.inspectStruct(ft, fieldPath, out)
		}
	case reflect.Array:
		if !c.isLeaf(ft) {
			c.inspectArrayElem(ft, fieldPath, out)
		}
	}
}

// inspectArrayElem applies the rule for an array field's ELEMENT type. An
// array field is not itself a reference — its bytes are copied by value —
// but indexing into it can still hand an external caller a mutable
// pointer/slice/map/chan/func, or an interface whose dynamic type cannot be
// proven immutable, or a nested struct that itself needs inspection, no
// matter that the field itself is unexported. It is the element TYPE that
// decides this, not any one array position, so this recurses at most once
// regardless of the array's length, appending "[]" to fieldPath to mark
// that the descent crossed an array. A scalar element (int, string, bool,
// float, and so on) is immutable by value semantics and reported as a safe
// leaf, same as a nested array element (Kind Array): only one array level
// is inspected, matching the elementary kinds this function already treats
// as a leaf by falling through the switch.
func (c *config) inspectArrayElem(ft reflect.Type, fieldPath string, out *[]Violation) {
	elem := ft.Elem()
	if c.isLeaf(elem) {
		return
	}
	switch elem.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		*out = append(*out, Violation{
			Path: fieldPath,
			Type: ft.String(),
			Rule: RuleReferenceField,
			Reason: "array field can expose a mutable pointer/slice/map/chan/func element " +
				"through indexing regardless of export status",
		})
	case reflect.Interface:
		*out = append(*out, Violation{
			Path:   fieldPath,
			Type:   ft.String(),
			Rule:   RuleInterfaceField,
			Reason: "interface field's dynamic type cannot be proven immutable by structural inspection",
		})
	case reflect.Struct:
		c.inspectStruct(elem, fieldPath+"[]", out)
	}
}

// Check is sugar for a one-line test: Inspect(opts...).Err().
func Check[T any](opts ...Option) error {
	return Inspect[T](opts...).Err()
}
