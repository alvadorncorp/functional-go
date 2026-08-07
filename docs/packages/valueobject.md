# Package `valueobject`

**Import path:** `github.com/alvadorncorp/gopackages/functional/valueobject`

Reflectively verifies that a Go type follows the structural convention of an
immutable value object: no field an external caller can use to mutate an
already-constructed value. **This is a test helper, not a runtime guard.**

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Entry points | `Check[T]`, `Inspect[T]` |
| Configuration | `Option`, `AllowType`, `AllowPackage` |
| Result | `Report`, `Violation`, `Rule` |

---

## What this package proves, and what it cannot

`Inspect` walks `T`'s fields and reports every one an external caller could
use to reach in and mutate the value: an exported field, a
pointer/slice/map/chan/func alias, or an interface whose dynamic type cannot
be proven immutable by structural inspection alone.

It **cannot** prove semantic correctness: it has no way to know whether a
constructor validates its input, or whether a `With*` method actually copies
before it writes. Go gives no way to embed "unexported fields," "a
validating constructor," or "copy-on-write `With*` methods" as a reusable
library contract — those remain decisions each consumer makes for its own
type. `Inspect` checks the narrower, mechanical part: structural reachability
for mutation, not intent.

---

## `Check[T]` / `Inspect[T]`

```go
func Check[T any](opts ...Option) error   // sugar for Inspect[T](opts...).Err()
func Inspect[T any](opts ...Option) Report
```

`T`'s `reflect.Type` is obtained from `reflect.TypeOf(zero)` on a concrete
zero value, not `reflect.TypeFor[T]()` — on go1.26.4, `reflect.TypeFor` over
a type parameter instantiated only in an external test package triggers a
real linker bug (`panic: R_USEIFACE ... eqfunc ...`); `reflect.TypeOf` on a
concrete value reaches the same `reflect.Type` without going through that
code path. A non-struct `T` is treated as a leaf — this package exists to
verify value objects, which are structs.

Typical use, in a table or a single assertion in the type's own test file:

```go
func TestMoney_ValueObject(t *testing.T) {
	if err := valueobject.Check[Money](); err != nil {
		t.Error(err)
	}
}
```

---

## Rules

| Rule | Reported when |
|---|---|
| `RuleExportedField` | The field is exported: a caller can assign to it directly. |
| `RuleReferenceField` | The field is a pointer, slice, map, chan, or func — mutable through the alias regardless of export status. |
| `RuleInterfaceField` | The field is an interface whose dynamic type cannot be proven immutable by structural inspection. |

`Inspect` recurses into an unexported struct field that is not a leaf,
building a dotted `Path` such as `"Money.tags"` or `"Order.lines.items"`.

---

## `Report` / `Violation`

```go
type Violation struct {
    Path   string // dotted field path, e.g. "Money.tags"
    Type   string // the field's Go type name, e.g. "[]string"
    Rule   Rule
    Reason string
}

type Report struct{ Violations []Violation }

func (r Report) OK() bool  // no violation found
func (r Report) Err() error // nil when OK; else one line per violation
```

`Report.Err()` formats each violation as `"<path>: <reason> (<type>)"`, one
per line — this is exactly what a failing `t.Error(err)` prints. Example
(from `ExampleCheck`):

```text
Bad.Name: field is exported: a caller can assign to it directly (string)
Bad.tags: field can be mutated through a pointer/slice/map/chan/func alias regardless of export status ([]string)
```

---

## Leaves: what `Inspect` stops descending into

Default immutable leaves, with no `Option` needed:

| Leaf | Why |
|---|---|
| `time.Time`, `time.Duration` | Stdlib value types treated as opaque. |
| Any type declared in [`functional/persistent`](persistent.md) | Structure is semantically immutable; copy-on-boundary via constructors/`Items`. |

The `persistent` leaf check is a **string comparison** against
`reflect.Type.PkgPath()`, not a real import — `valueobject` does not import
`persistent`. See the module [overview](../overview.md#no-public-subpackage-imports-another)
for why that matters.

**Known gap:** the `persistent` leaf check considers only the container's
own package, never its type parameter. `persistent.List[*Widget]` is treated
as a leaf and reports zero violations even though the `*Widget` elements
obtained through `Values()`/`Items()` remain mutable through the pointer —
`reflect` exposes no API to inspect a generic instantiation's type
arguments, so `Inspect` cannot discover the element type and recurse into
it. Contrast with `optional.Optional[*Widget]`, which holds its type
parameter inline and so IS correctly flagged with `RuleReferenceField`. When
a `persistent` element type is itself a
pointer/slice/map/chan/func/interface, combine `AllowType`/manual review for
that element, or prefer `optional.Optional[T]`/a scalar field instead. See
`Inspect`'s doc comment for the full explanation.

| Option | Effect |
|---|---|
| `AllowType[T]() Option` | Marks the concrete type `T` as an immutable leaf. |
| `AllowPackage(importPath string) Option` | Marks every type declared under `importPath` (by `PkgPath()` prefix) as an immutable leaf. Creates no import edge — same string-comparison mechanism as the default `persistent` leaf. |

`AllowPackage`'s naming follows the functional-options convention used
elsewhere in this repository (`eventbus.Option`, the workopher engine and
adapter packages) — a different concept from the sibling package
`functional/optional`'s `Optional[T]`, which `valueobject` does not import.

---

## Related documents

- [Overview](../overview.md)
- [Value object pattern](../patterns/value-objects.md)
- [Package API: persistent](persistent.md)
