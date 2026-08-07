# Package `frozen`

**Import path:** `github.com/alvadorncorp/gopackages/functional/frozen`

Read-only, allocation-copying snapshots of a slice or map. `Slice[T]` and
`Map[K,V]` copy on construction and copy again on every read that could
otherwise leak the backing array or map — no caller can reach or mutate what
either type holds. Neither type has an update operation and never will: for
a collection that evolves through functional updates, use
[`persistent`](persistent.md).

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Immutable slice | `Slice[T]`, `NewSlice`, `CollectSlice` |
| Immutable map | `Map[K,V]`, `NewMap`, `CollectMap` |
| Element-wise equality | `SliceEqualFunc`, `MapEqualFunc` |

---

## `Slice[T]`

```go
type Slice[T any] struct {
    // unexported: items []T
}
```

The zero `Slice[T]` is a valid, empty `Slice` — a struct field of this type
needs no constructor. `Slice` is **not** `==`-comparable (it holds a slice
header internally); compare with `Equal`.

| Function / method | Behaviour |
|---|---|
| `NewSlice[T](items ...T) Slice[T]` | Copies `items`. Passing an existing slice with `xs...` still copies — later writes to `xs` never reach the `Slice`. |
| `CollectSlice[T](seq iter.Seq[T]) Slice[T]` | Copies the elements `seq` produces. |
| `(s Slice[T]) Len() int` | Number of elements. |
| `(s Slice[T]) At(i int) T` | Element at `i`; panics `"frozen: index out of range"` outside `[0, Len)`. |
| `(s Slice[T]) Values() iter.Seq[T]` | Iterates without allocating. Prefer this over `Items()` for a read-only walk. |
| `(s Slice[T]) All() iter.Seq2[int, T]` | Index/value pairs, mirroring `slices.All`. |
| `(s Slice[T]) Items() []T` | Fresh mutable copy. **Allocates on every call.** |
| `(s Slice[T]) Equal(other Slice[T]) bool` | Same elements, same order, compared element-wise via `functional/internal/eq.Equal`. |

---

## `Map[K,V]`

```go
type Map[K comparable, V any] struct {
    // unexported: items map[K]V
}
```

The zero `Map[K,V]` is a valid, empty `Map`, same copy-on-ingress/egress
contract as `Slice`.

| Function / method | Behaviour |
|---|---|
| `NewMap[K,V](m map[K]V) Map[K,V]` | Copies `m`. Later writes to `m` never reach the `Map`. |
| `CollectMap[K,V](seq iter.Seq2[K,V]) Map[K,V]` | Copies the pairs `seq` produces. |
| `(m Map[K,V]) Get(k K) (V, bool)` | Comma-ok, **deliberately not** `optional.Optional[V]` — see below. |
| `(m Map[K,V]) Has(k K) bool` | Membership test. |
| `(m Map[K,V]) Len() int` | Number of entries. |
| `(m Map[K,V]) All() iter.Seq2[K,V]` | Key/value pairs, unspecified order. |
| `(m Map[K,V]) Keys() iter.Seq[K]` | Keys, unspecified order. |
| `(m Map[K,V]) Values() iter.Seq[V]` | Values, unspecified order. |
| `(m Map[K,V]) Items() map[K]V` | Fresh, non-nil copy. **Allocates on every call.** |
| `(m Map[K,V]) Equal(other Map[K,V]) bool` | Same key/value pairs, values compared via `functional/internal/eq.Equal`. Order does not matter. |

### Why `Get` returns comma-ok, not `optional.Optional[V]`

`frozen` does not import `optional` — no public subpackage of `functional`
imports another (see the module [overview](../overview.md)). Bridge at the
call site instead: `optional.FromOK(m.Get(k))`.

---

## Element-wise equality helpers

| Function | Behaviour |
|---|---|
| `SliceEqualFunc[T, U](a Slice[T], b Slice[U], eq func(T, U) bool) bool` | Same length, `eq` true for every index pair. |
| `MapEqualFunc[K, V1, V2](a Map[K,V1], b Map[K,V2], eq func(V1, V2) bool) bool` | Same length, and for every key in `a`, `b` has that key with an `eq`-equal value. |

`Slice.Equal` and `Map.Equal` are implemented in terms of these, seeded with
`functional/internal/eq.Equal[T]` — the same three-tier equality
(`Equal` method, then `==`, then `reflect.DeepEqual`) used across the
module.

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [Package API: persistent](persistent.md) (the mutable-write sibling)
- [Value object pattern](../patterns/value-objects.md)
