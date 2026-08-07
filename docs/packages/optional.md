# Package `optional`

**Import path:** `github.com/alvadorncorp/gopackages/functional/optional`

`Optional[T]`, a value type for an operation that may or may not produce a
`T`. Explicit absence, tracked apart from the value — never inferred by
comparing the value to its zero.

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Core type | `Optional[T]` |
| Construction | `Some`, `None`, `FromPointer`, `FromOK` |
| Presence | `IsSome`, `IsNone`, `IsZero` |
| Extraction | `Get`, `MustGet`, `OrElse`, `OrElseGet`, `OrZero`, `Pointer` |
| Precedence | `Or` |
| Filtering | `Filter` |
| Type-changing (package functions) | `Map`, `FlatMap`, `MapErr`, `Zip`, `EqualFunc` |
| Equality / display | `Equal` (method), `String` |

---

## Why `Map`/`FlatMap`/`Zip` are functions, not methods

Go does not allow a method to introduce its own type parameters — only a
free function can bind a second type parameter beyond the receiver's own
`T`. Every operation on `Optional[T]` that would need to return
`Optional[U]` for some other type `U` is therefore a package-level function
instead of a method. See
[overview.md](../overview.md#the-constraint-that-shapes-the-whole-api-no-type-parameters-on-methods)
and
[ADR-001](../adr/functional/adr-001-optional-representation-and-zero-value.md).

---

## `Optional[T]`

```go
type Optional[T any] struct {
    // unexported: value T; present bool
}
```

The zero `Optional[T]` is a valid `None` — a struct field of this type needs
no constructor and no nil check. `Some(zero)` and `None[T]()` are distinct
values: presence is tracked apart from the held value.

Guards `T` inline, not behind a pointer:

- `Optional[T]` is `==`-comparable exactly when `T` is comparable;
- constructing a `Some` never allocates;
- the held value is a copy — it cannot be mutated through an external alias;
- immutability is **shallow**: `Optional[[]byte]` shares the slice's backing
  array. Compose with an immutable element type (e.g.
  [`persistent.List`](persistent.md)) when the element itself must not be mutable.

---

## Construction

| Function | Behaviour |
|---|---|
| `Some[T](value T) Optional[T]` | Present, holding `value`. `Some(nil)` for a pointer/map/slice/chan/func/interface `T` is present and holds `nil` — that is **not** the same as `None`. |
| `None[T]() Optional[T]` | Absent. Same value as the zero `Optional[T]`. |
| `FromPointer[T](p *T) Optional[T]` | `Some(*p)` when `p != nil`, `None` otherwise. The value is copied — later writes through `p` never reach the result. |
| `FromOK[T](value T, ok bool) Optional[T]` | Adapts any Go comma-ok result. Composes directly: `optional.FromOK(os.LookupEnv("HOME"))`. |

---

## Presence and extraction

| Method | Behaviour |
|---|---|
| `IsSome() bool` | Holds a value. |
| `IsNone() bool` | Holds no value. |
| `IsZero() bool` | Same as `!IsSome()`; enables the stdlib `omitzero` protocol for a struct field of this type. |
| `Get() (T, bool)` | Value (or `T`'s zero) plus presence, comma-ok style. |
| `MustGet() T` | Value, or panics `"functional/optional: Optional has no value"`. For tests and init, not the production path. |
| `OrElse(fallback T) T` | Value, or `fallback` when absent. |
| `OrElseGet(fallback func() T) T` | Value, or `fallback()` — called only when absent; use for an expensive fallback. |
| `OrZero() T` | Value, or `T`'s zero when absent. |
| `Pointer() *T` | Pointer to a **copy** of the value, or `nil` when absent. Caller owns the copy. |
| `Or(other Optional[T]) Optional[T]` | The receiver when present, `other` otherwise — chainable for a precedence order. |
| `Filter(keep func(T) bool) Optional[T]` | The receiver when present and `keep` reports `true`; `None` otherwise. `keep` is never called on an absent `Optional`. |

---

## Type-changing package functions

| Function | Behaviour |
|---|---|
| `Map[T, U](o Optional[T], f func(T) U) Optional[U]` | `Some(f(value))` when present, `None` otherwise. `f` never called when absent. |
| `FlatMap[T, U](o Optional[T], f func(T) Optional[U]) Optional[U]` | `f(value)`'s result when present, `None` otherwise. `f` never called when absent. |
| `MapErr[T, U](o Optional[T], f func(T) (U, error)) (Optional[U], error)` | `(None, nil)` when absent — `f` not called; `(None, err)` on failure; `(Some(v), nil)` on success. |
| `Zip[A, B, R](a Optional[A], b Optional[B], combine func(A, B) R) Optional[R]` | `Some(combine(a, b))` only when both present; `combine` never called otherwise. Takes `combine` so callers don't need an exported `Pair[A, B]` type. |
| `EqualFunc[T, U](a Optional[T], b Optional[U], eq func(T, U) bool) bool` | Whether both hold values `eq` considers equal, or both are absent. `eq` never called unless both are present. |

---

## Equality and display

`Equal` is a **method**, not a function — `go-cmp` only consults a method
named `Equal`, and without one `cmp.Diff` panics on a struct carrying
unexported fields (`Optional[T]`'s `value`/`present` pair). It delegates
element-wise to `functional/internal/eq.Equal`, the same resolution (an
`Equal` method on the dynamic value, then `==` when comparable;
non-comparable types without `Equal` return false — no `reflect`) every
generic `Equal` in this module uses — so it works even when `T` carries no
comparability constraint.

`String() string` renders `"Some(<value>)"` or `"None"` via
`fmt.Sprintf("%v", ...)`. It exists for readable test failures and `%v`; it
is **not** a serialization format — see the module README's non-goals for
the `encoding/json` pitfall.

---

## What this package is not

| Non-goal | Detail |
|---|---|
| Serialization | No `encoding/json`, `database/sql`, or `slog` integration; a struct field marshals as `{}`, losing the value |
| Monad zoo | No `IO`, `Task`, `Reader`, `State` |
| Deep immutability | Shallow only — `T`'s own mutability passes through |

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [ADR-001 — Optional representation and zero value](../adr/functional/adr-001-optional-representation-and-zero-value.md)
