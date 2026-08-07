# Package `persistent`

**Import path:** `github.com/alvadorncorp/gopackages/functional/persistent`

Immutable collections — `List[T]`, `Map[K,V]`, `Set[T]` — that write
functionally: every update method returns a new collection and leaves the
receiver unchanged. There is no in-place mutation anywhere in this API;
"write" always means "derive a new value from an old one."

## Engines (Baseline B)

| Type | Engine | Sharing model |
|---|---|---|
| `Map[K,V]`, `Set[T]` | `internal/engine/hamt` | 32-way HAMT; path-copy structural sharing |
| `List[T]` | `internal/engine/vector` | 32-way persistent vector (trie + tail); path-copy on Set/Append |

- **HAMT seed policy:** each independent empty `hamt.Root` holds its own
  `maphash.Seed`; derived versions preserve that seed. Key hashing uses
  `maphash.Comparable` via `internal/engine/keyhash`.
- **Builders:** open an internal `transient.Window`, edit under the window's
  exclusive `*ownership.Owner` (engine transient paths mutate owned nodes in
  place), and `Freeze` on `Build` — **O(1)** owner invalidation, no tree
  walk. Residual inactive stamps fail `IsOwnedBy`, so no Detach walk is
  required.
- **`List.Delete`:** currently **O(n)** (vector rebuild by iteration). A
  caller MUST NOT depend on a tighter bound for Delete.

Documented log-n complexity bounds below are **upper-bound public-API
headroom**. Decision records:
[ADR-004](../adr/functional/adr-004-thin-immutability-core.md) (thin core),
[ADR-005](../adr/functional/adr-005-persistent-engine-family.md) (engine
family). Historical note: early milestones used full-copy COW over a stdlib
slice/map; that is no longer the engine.

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Ordered collection | `List[T]`, `NewList`, `CollectList`, `ListBuilder[T]` |
| Key-value collection | `Map[K,V]`, `NewMap`, `CollectMap`, `MapBuilder[K,V]` |
| Set collection | `Set[T]`, `NewSet`, `CollectSet`, `SetBuilder[T]` |
| Engines | `internal/engine/hamt` (Map/Set), `internal/engine/vector` (List) |
| Mutation substrate | Builders use `internal/core/transient` + engine exclusive-owner paths |

---

## The contract every type follows

1. **No method accepts or returns the internal representation.** `Items()`
   always copies rather than handing out the backing engine storage.
2. **`Map`/`Set` iteration order is UNSPECIFIED** and may change between
   releases, or even between two calls on the same value. `List` iterates
   in index order — that order is specified and part of the contract.
3. **Documented complexity is an upper bound** the public API reserves room
   for (see the table below). Path-copy updates allocate along a tree path
   (depth curve), not a full collection copy — but callers must still use
   the documented bounds, not today's constants.
4. **None of the three types are `==`-comparable.** Use `Equal`. Comparing
   by `==` would compare identity of the underlying representation.
5. **Calling `Append`/`Set`/`Delete`/`Add`/`Remove` `n` times in a loop still
   allocates `n` intermediate collections.** `Builder()` exists for bulk
   construction: exclusive ownership + O(1) `Freeze`. Builders open a
   `transient.Window` and `Freeze` on `Build`; public panics remain
   `"persistent: ..."`.
6. **Every test in this package is `package persistent_test`** (external),
   asserting only publicly observable behavior — the acceptance criterion
   for any future engine change is this suite passing unchanged.

---

## Documented complexity (upper bound)

| Operation | Documented bound | Notes |
|---|---|---|
| `Len` | `O(1)` | |
| `At`, `Get`, `Has` | `O(log n)` | |
| `Set`, `Delete` (Map), `Add`/`Remove` (Set) | `O(log n)` | Path-copy HAMT |
| `Append`, `Set` (List) | `O(log n)` amortized / path | Path-copy vector |
| `Delete` (List) | `O(log n)` headroom | **Today O(n)** rebuild |

Only the table's documented bound is guaranteed as API headroom for methods
that meet it. `List.Delete`'s current implementation is O(n); do not depend
on a tighter constant. See [ADR-005](../adr/functional/adr-005-persistent-engine-family.md)
and [core.md](core.md).

---

## `List[T]`

| Function / method | Behaviour |
|---|---|
| `NewList[T](items ...T) List[T]` | Copies `items`, in order (via vector append). |
| `CollectList[T](seq iter.Seq[T]) List[T]` | Drains `seq` into a new `List`, in iteration order. |
| `(l List[T]) Len() int` | Element count. |
| `(l List[T]) At(i int) T` | Element at `i`; panics `"persistent: index out of range"` outside `[0, Len)`. |
| `(l List[T]) Append(items ...T) List[T]` | New `List` with `items` appended. `l` unchanged. Path-copy vector. |
| `(l List[T]) Set(i int, v T) List[T]` | New `List` with index `i` replaced. Panics as above. |
| `(l List[T]) Delete(i int) List[T]` | New `List` without index `i`; later elements shift down. Panics as above. **Currently O(n).** |
| `(l List[T]) Values() iter.Seq[T]` | Elements in index order. |
| `(l List[T]) All() iter.Seq2[int, T]` | (index, element) pairs in index order. |
| `(l List[T]) Items() []T` | Fresh mutable copy; caller owns it. |
| `(l List[T]) Equal(other List[T]) bool` | Same elements, same order, via `functional/internal/eq.Equal`. |
| `(l List[T]) Builder() *ListBuilder[T]` | Seeds a `ListBuilder[T]` from `l` and opens a `transient.Window`. |

`ListBuilder[T]` — **mutable, not safe for concurrent use** — offers
`Append`, `Set`, `Delete` (all chainable, same panic messages as `List`),
`Len`, and `Build() List[T]`. `Build` freezes the builder's transient
window (O(1) owner invalidation), returns an immutable `List`, and
permanently disables the builder: every method afterward, including a
repeated `Build`, panics `"persistent: builder used after Build"`.

## `Map[K,V]`

| Function / method | Behaviour |
|---|---|
| `NewMap[K,V](m map[K]V) Map[K,V]` | Copies `m` into a fresh HAMT root. |
| `CollectMap[K,V](seq iter.Seq2[K,V]) Map[K,V]` | Drains `seq`; last value wins on a repeated key, matching `maps.Collect`. |
| `(m Map[K,V]) Len() int` / `Get(k K) (V, bool)` / `Has(k K) bool` | Read access. |
| `(m Map[K,V]) Set(k K, v V) Map[K,V]` | New `Map` with `k` mapped to `v`. `m` unchanged. Path-copy HAMT. |
| `(m Map[K,V]) Delete(k K) Map[K,V]` | New `Map` without `k`. `m` unchanged. |
| `(m Map[K,V]) All/Keys/Values() iter.Seq(2)[...]` | Unspecified iteration order. |
| `(m Map[K,V]) Items() map[K]V` | Fresh, non-nil copy. |
| `(m Map[K,V]) Equal(other Map[K,V]) bool` | Same key set, each value equal via `functional/internal/eq.Equal`. |
| `(m Map[K,V]) Builder() *MapBuilder[K,V]` | Seeds a `MapBuilder[K,V]` (opens a `transient.Window`). |

`MapBuilder[K,V]` — mutable, not concurrency-safe — offers `Set`, `Delete`
(chainable), `Len`, and `Build() Map[K,V]` with the same freeze,
publish-and-disable contract as `ListBuilder`.

## `Set[T]`

| Function / method | Behaviour |
|---|---|
| `NewSet[T](items ...T) Set[T]` | Deduplicated. |
| `CollectSet[T](seq iter.Seq[T]) Set[T]` | Deduplicated. |
| `(s Set[T]) Len() int` / `Has(v T) bool` | Read access. |
| `(s Set[T]) Add(items ...T) Set[T]` | New `Set` with `items` added; already-present items are a no-op. `s` unchanged. |
| `(s Set[T]) Remove(items ...T) Set[T]` | New `Set` without `items`; absent items are a no-op. `s` unchanged. |
| `(s Set[T]) Values() iter.Seq[T]` | Unspecified iteration order. |
| `(s Set[T]) Items() []T` | Fresh slice, unspecified order. |
| `(s Set[T]) Equal(other Set[T]) bool` | Same membership. `T` is `comparable`, so no `eq.Equal` call is needed. |
| `(s Set[T]) Builder() *SetBuilder[T]` | Seeds a `SetBuilder[T]` (opens a `transient.Window`). |

`SetBuilder[T]` — mutable, not concurrency-safe — offers `Add`, `Remove`
(chainable), `Len`, and `Build() Set[T]` with the same freeze,
publish-and-disable contract.

---

## Measured: single update vs. Builder loop

Path-copy engines allocate along a **depth curve** (roughly O(log₃₂ n)
nodes), not a full `O(n)` collection copy. The property to watch in
`go test -bench` is **allocs/op** (and B/op) **not scaling linearly with n**
for a single `Map.Set` / `Set.Add` / `List.Append` against a pre-built base.

Looping `Append`/`Set`/`Add` on the immutable type still creates **n
intermediate roots**. `Builder` is the bulk path: exclusive ownership,
in-place edits on owned nodes, O(1) `Freeze`. Native mutable baselines
(`[]T` append, `map[K]V` assign) live in `persistent/bench_test.go` for
comparison.

```sh
cd functional
go test ./persistent/ -run '^$' -bench 'BenchmarkMapSet|BenchmarkSetAdd|BenchmarkListAppend|BenchmarkListBuilder' -benchmem
```

Always route loop-building through `<Type>.Builder()`; never call
`Append`/`Set`/`Delete`/`Add`/`Remove` in a loop against the immutable type
itself when linear bulk construction is the goal.

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [Internal: core](core.md)
- [Package API: frozen](frozen.md) (the read-only sibling)
- [ADR-004 — Thin immutability core](../adr/functional/adr-004-thin-immutability-core.md)
- [ADR-005 — Persistent engine family](../adr/functional/adr-005-persistent-engine-family.md)
