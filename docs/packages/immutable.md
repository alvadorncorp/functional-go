# Package `immutable`

**Import path:** `go.alvadorn.org/functional/immutable`

Immutable collections — `List[T]`, `Map[K,V]`, `Set[T]`, `SortedMap[K,V]` —
that write functionally: every update method returns a new collection and
leaves the receiver unchanged. There is no in-place mutation anywhere in this
API; "write" always means "derive a new value from an old one."

Constructors are **`ListOf` / `MapOf` / `SetOf` / `EmptySortedMap`** (not
`New*`). See the [name glossary](#name-glossary) for historical renames.

## Engines (Baseline B)

| Type | Engine | Sharing model |
|---|---|---|
| `Map[K,V]`, `Set[T]` | `internal/engine/hamt` | 32-way HAMT; path-copy structural sharing |
| `List[T]` | `internal/engine/vector` | 32-way persistent vector (trie + tail); path-copy on Set/Append and tail removal on last-element Delete |
| `SortedMap[K,V]` | `internal/engine/wbtree` | Weight-balanced binary search tree (Adams; Straka's `delta = 3`, `gamma = 2` on weight `size+1`); path-copy structural sharing; subtree size per node; see [ADR-013](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md) |

- **HAMT seed policy:** each independent empty `hamt.Root` holds its own
  `maphash.Seed`; derived versions preserve that seed. Key hashing uses
  `maphash.Comparable` via `internal/engine/keyhash`. The zero `Map`/`Set`
  is a valid empty collection for reads; the **first write**
  (`Set`/`Update`/`Delete`/`Add`/`Remove`) or `Builder()` **materializes**
  a proper `hamt.Empty` seed so the lineage never freezes a zero
  `maphash.Seed`. `MapOf`, `SetOf`, `CollectMap`, and `CollectSet` with empty
  input return the zero value and create no seed; it materializes on the
  first write or `Builder()`.
  Independent zero receivers each get their own seed on first materialization.
  `Set.Union`, `Set.Intersection`, and `Set.Difference` accept operands of
  any lineage (independent seeds or one derived lineage); which operand's
  structure and seed lineage the result continues is **UNSPECIFIED**.
- **wbtree comparator policy:** each `wbtree.Root` carries the comparator it
  was constructed with, and every derived version keeps it, including one
  emptied by `Delete`. Rotations look only at subtree sizes; every comparator
  call on a path runs while descending, before any node is edited, so a
  comparator panic leaves a builder's prior contents intact. `immutable`
  checks for a comparator before calling the engine, so `"wbtree: ..."`
  panics never reach callers.
- **Builders:** open an internal `transient.Window`, edit under the window's
  exclusive `*ownership.Owner` (engine transient paths mutate owned nodes in
  place), and `Freeze` on `Build` — **O(1)** owner invalidation, no tree
  walk. Residual inactive stamps fail `IsOwnedBy`, so no stamp-clearing walk
  is required. Builders are **not** safe for concurrent use.
  Bulk constructors, collectors, and multi-item `Append`/`Add`/`Remove` use the
  transient edit path internally; their receivers remain immutable. Empty
  `ListOf`/`MapOf`/`SetOf` return the zero collection with no edit window and
  no allocation. An empty `Collect*` sequence also returns the zero, unseeded
  collection, but that is not an allocation guarantee. Single-item writes
  (including the first item of a `Collect*` call and a one-entry `MapOf`) are
  pure, with no transient window. That is not zero total allocation: it only
  skips the builder and window setup.
- **`List.Delete`:** middle-index deletion rebuilds the vector in **O(n)**;
  last-element deletion uses internal tail removal (O(1) with a non-singleton
  tail, otherwise O(log₃₂ n)). The public O(n) upper bound remains unchanged;
  there is no public `List.Pop`.
- **No `reflect` in production:** `immutable` production code does not import
  `reflect`. Equality goes through `functional/internal/eq` (method, then
  `==`; never `reflect.DeepEqual`).
  `List.Equal`, `Map.Equal`, and `SortedMap.Equal` select the
  value-comparison strategy (`eq.For[V]()`) once per call. Only
  `List.Equal`, `Map.Equal`, and `Set.Equal` also recognize a shared root, so
  the same version is reflexive even with NaN values or keys, or a
  non-reflexive custom `Equal` method (the shared-root check bypasses it).
  `SortedMap.Equal` has no identity shortcut and keeps its comparator-based
  lookup; `EqualFunc` on any type never skips the caller's comparison.
- **Semantic immutability:** the collection structure never mutates in place.
  Stored values are not deep-copied; mutable referents shared by the caller
  remain shared across versions.
- **Concurrency:** published (immutable) collections are safe for concurrent
  read, including while another goroutine derives new versions from the same
  base.
- **Engine micro-benches:** optional alloc/path-copy probes live in
  `internal/engine/hamt/benchmark_test.go`,
  `internal/engine/vector/benchmark_test.go`, and
  `internal/engine/wbtree/benchmark_test.go` (not part of the public API).

Documented log-n complexity bounds below are **upper-bound public-API
headroom**. Decision records:
[ADR-004](../adr/functional/adr-004-thin-immutability-core.md) (thin core),
[ADR-005](../adr/functional/adr-005-persistent-engine-family.md) (engine
family), [ADR-013](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md)
(SortedMap on wbtree). Historical note: early milestones used full-copy COW
over a stdlib slice/map; that is no longer the engine.

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Ordered collection | `List[T]`, `ListOf`, `CollectList`, `ListBuilder[T]` |
| Key-value collection | `Map[K,V]`, `MapOf`, `CollectMap`, `MapBuilder[K,V]` |
| Set collection | `Set[T]`, `SetOf`, `CollectSet`, `SetBuilder[T]` |
| Sorted key-value collection | `SortedMap[K,V]`, `EmptySortedMap`, `CollectSortedMap`, `SortedMapBuilder[K,V]` |
| Engines | `internal/engine/hamt` (Map/Set), `internal/engine/vector` (List), `internal/engine/wbtree` (SortedMap) |
| Mutation substrate | Builders use `internal/core/transient` + engine exclusive-owner paths |

---

## The contract every type follows

1. **No method accepts or returns the internal representation.** `Items()`
   always copies rather than handing out the backing engine storage.
2. **`Map`/`Set` iteration order is UNSPECIFIED** and may change between
   releases, or even between two calls on the same value. `List` iterates
   in index order and `SortedMap` in ascending comparator order — those
   orders are specified and part of the contract.
3. **Documented complexity is an upper bound** the public API reserves room
   for (see the table below). Path-copy updates allocate along a tree path
   (depth curve), not a full collection copy — but callers must still use
   the documented bounds, not today's constants.
4. **None of the four types (`List`, `Map`, `Set`, `SortedMap`) are
   `==`-comparable.** Use `Equal` or `EqualFunc`. Comparing by `==` would
   compare identity of the underlying representation.
5. **Calling `Append`/`Set`/`Delete`/`Update`/`Add`/`Remove` `n` times in a
   loop still allocates `n` intermediate collections.** `Builder()` exists
   for bulk construction: exclusive ownership + O(1) `Freeze`. Builders open a
   `transient.Window` and `Freeze` on `Build`; public panics remain
   `"immutable: ..."`.
6. **Zero `List`, `Map`, and `Set` values are valid empty collections.**
   Map/Set first write or `Builder()` materializes an `hamt.Empty` seed; empty
   `MapOf`/`SetOf`/`Collect*` input stays the unseeded zero value.
   **`SortedMap` is the deliberate exception:** `K` has no default order, so
   the zero `SortedMap` has no comparator. It reads as empty and `Delete` on
   it is a no-op, but `Set` and `Builder()` panic with `"immutable: zero
   SortedMap has no comparator; construct with EmptySortedMap or
   CollectSortedMap"`. A zero-safe getter for a `SortedMap` field must
   return `EmptySortedMap(cmp)`, never the zero `SortedMap`.
7. **Every test in this package is `package immutable_test`** (external),
   asserting only publicly observable behavior — the acceptance criterion
   for any future engine change is this suite passing unchanged.

---

## Documented complexity (upper bound)

| Operation | Documented bound | Notes |
|---|---|---|
| `Len`, `IsEmpty` | `O(1)` | |
| `At`, `Get`, `Has` | `O(log n)` | |
| `Set`, `Delete`, `Update` (Map), `Add`/`Remove` (Set) | `O(log n)` | Path-copy HAMT; `Update` is one `Get` plus one `Set` |
| `Set`, `Delete` (SortedMap) | `O(log n)` | Path-copy weight-balanced tree; `O(log n)` comparator calls, at most one single or double rotation per level |
| `Union`, `Intersection`, `Difference` (Set) | `O(min(m,n)·log max(m,n))` | Walks the smaller operand (size `m` or `n`), probes the larger |
| `Append`, `Set` (List) | `O(log n)` amortized / path | Path-copy vector |
| `Delete` (List) | `O(n)` | Middle index rebuilds by iteration; last index uses tail removal (faster), but the public bound is `O(n)` |

Only the table's documented bound is guaranteed as API headroom for methods
that meet it. `List.Delete`'s public bound is O(n): the last index is
faster, but do not depend on a tighter constant. See
[ADR-005](../adr/functional/adr-005-persistent-engine-family.md)
and [core.md](core.md).

---

## Name glossary

| Prefer | Avoid / historical |
|---|---|
| `ListOf` / `MapOf` / `SetOf` | `NewList` / `NewMap` / `NewSet` |
| `EmptySortedMap` / `CollectSortedMap` | `NewSortedMap` / `SortedMapOf` (`Of` builds from contents; a `SortedMap` starts from a comparator) |
| `Has` | `Contains` |
| `Build` (public builder publish) | `Freeze` (internal `transient.Window` op) |

---

## `List[T]`

| Function / method | Behaviour |
|---|---|
| `ListOf[T](items ...T) List[T]` | Copies `items`, in order. Empty: zero `List`. One item: pure `Append`. Two or more: one transient `ListBuilder` window. |
| `CollectList[T](seq iter.Seq[T]) List[T]` | Drains `seq` into a new `List`, in iteration order. First item is a pure `Append`; later items go through one builder. Empty `seq`: zero `List`. |
| `(l List[T]) Len() int` | Element count. |
| `(l List[T]) IsEmpty() bool` | `Len() == 0`. Zero `List` is empty. |
| `(l List[T]) At(i int) T` | Element at `i`; panics `"immutable: index out of range"` outside `[0, Len)`. |
| `(l List[T]) Append(items ...T) List[T]` | New `List` with `items` appended. `l` unchanged. Path-copy vector. |
| `(l List[T]) Set(i int, v T) List[T]` | New `List` with index `i` replaced. Panics as above. |
| `(l List[T]) Delete(i int) List[T]` | New `List` without index `i`; later elements shift down. Panics as above. Middle indexes are O(n); the last index uses tail removal. Public bound **O(n)**. |
| `(l List[T]) Values() iter.Seq[T]` | Elements in index order. |
| `(l List[T]) All() iter.Seq2[int, T]` | (index, element) pairs in index order. |
| `(l List[T]) Items() []T` | Fresh mutable copy; caller owns it. |
| `(l List[T]) Equal(other List[T]) bool` | Same elements, same order. A shared root returns `true`. Otherwise values compare via `eq.For[T]()`, selected once per call (no `DeepEqual`). |
| `(l List[T]) EqualFunc(other List[T], eq func(T, T) bool) bool` | Same length; pairs compared with `eq`. |
| `(l List[T]) Builder() *ListBuilder[T]` | Seeds a `ListBuilder[T]` from `l` and opens a `transient.Window`. |

`ListBuilder[T]` — **mutable, not safe for concurrent use** — offers
`Append`, `Set`, `Delete` (all chainable, same panic messages as `List`),
`Len`, `At`, and `Build() List[T]`. `Build` freezes the builder's transient
window (O(1) owner invalidation), returns an immutable `List`, and
permanently disables the builder: every method afterward, including a
repeated `Build`, panics `"immutable: builder used after Build"`.

## `Map[K,V]`

| Function / method | Behaviour |
|---|---|
| `MapOf[K,V](m map[K]V) Map[K,V]` | Copies `m`. Empty `m`: zero `Map`, no seed. One entry: pure `Set`. Two or more: one transient `MapBuilder` window on a fresh seeded root. |
| `CollectMap[K,V](seq iter.Seq2[K,V]) Map[K,V]` | Drains `seq`; on a repeated key the last key and last value win, matching `maps.Collect`. Empty `seq`: zero `Map`, no seed. |
| `(m Map[K,V]) Len() int` | Entry count. |
| `(m Map[K,V]) IsEmpty() bool` | `Len() == 0`. Zero `Map` is empty. |
| `(m Map[K,V]) Get(k K) (V, bool)` / `Has(k K) bool` | Read access. |
| `(m Map[K,V]) Set(k K, v V) Map[K,V]` | New `Map` with `k` mapped to `v`. `m` unchanged. Path-copy HAMT. Zero `Map` materializes seed first. |
| `(m Map[K,V]) Update(k K, fn func(old V, present bool) V) Map[K,V]` | New `Map` with `k` mapped to `fn(old, present)` (`old` is the zero value when `k` is absent). `fn` runs exactly once, before any write; its result is always stored. `m` unchanged. Nil `fn` is not special-cased. Zero `Map` materializes seed first. |
| `(m Map[K,V]) Delete(k K) Map[K,V]` | New `Map` without `k`. `m` unchanged. Zero `Map` materializes seed first. |
| `(m Map[K,V]) All/Keys/Values() iter.Seq(2)[...]` | Unspecified iteration order. |
| `(m Map[K,V]) Items() map[K]V` | Fresh, non-nil copy. |
| `(m Map[K,V]) Equal(other Map[K,V]) bool` | Same key set, each value equal. A shared root returns `true`. Otherwise values compare via `eq.For[V]()`, selected once per call (no `DeepEqual`). |
| `(m Map[K,V]) EqualFunc(other Map[K,V], eq func(V, V) bool) bool` | Same keys; values compared with `eq`. |
| `(m Map[K,V]) Builder() *MapBuilder[K,V]` | Seeds a `MapBuilder[K,V]` (opens a `transient.Window`). Zero `Map` materializes seed first. |

`MapBuilder[K,V]` — mutable, not concurrency-safe — offers `Set`, `Delete`
(chainable), `Len`, `Get`, `Has`, and `Build() Map[K,V]` with the same freeze,
publish-and-disable contract as `ListBuilder`.

## `Set[T]`

| Function / method | Behaviour |
|---|---|
| `SetOf[T](items ...T) Set[T]` | Deduplicated; for repeated equal items the last is the stored representative. Empty input: zero `Set`, no seed. One item: pure `Add`. Two or more: one transient `SetBuilder` window on a fresh seeded root. |
| `CollectSet[T](seq iter.Seq[T]) Set[T]` | Deduplicated; the last equal element is the stored representative. Empty `seq`: zero `Set`, no seed. |
| `(s Set[T]) Len() int` | Element count. |
| `(s Set[T]) IsEmpty() bool` | `Len() == 0`. Zero `Set` is empty. |
| `(s Set[T]) Has(v T) bool` | Membership. |
| `(s Set[T]) Add(items ...T) Set[T]` | New `Set` with `items` added. An item equal to a held one replaces the stored representative (observable for `+0`/`-0` and shared strings); membership and `Len` are unchanged. `s` unchanged. Zero `Set` materializes seed first. |
| `(s Set[T]) Remove(items ...T) Set[T]` | New `Set` without `items`; absent items are a no-op. `s` unchanged. Zero `Set` materializes seed first. |
| `(s Set[T]) Values() iter.Seq[T]` | Unspecified iteration order. |
| `(s Set[T]) Items() []T` | Fresh slice, unspecified order. |
| `(s Set[T]) Equal(other Set[T]) bool` | Same membership. A shared root returns `true`. `T` is `comparable`, so no `eq` strategy is involved. |
| `(s Set[T]) Union(other Set[T]) Set[T]` | Every element of `s` or `other`. Both unchanged. Shared structure and seed lineage UNSPECIFIED. |
| `(s Set[T]) Intersection(other Set[T]) Set[T]` | Every element of `s` that `other` also holds. Both unchanged. Shared structure and seed lineage UNSPECIFIED. |
| `(s Set[T]) Difference(other Set[T]) Set[T]` | `s` minus `other`. Both unchanged. Shared structure and seed lineage UNSPECIFIED. |
| `(s Set[T]) Builder() *SetBuilder[T]` | Seeds a `SetBuilder[T]` (opens a `transient.Window`). Zero `Set` materializes seed first. |

`SetBuilder[T]` — mutable, not concurrency-safe — offers `Add` (an equal item
replaces the stored representative, as in `Set.Add`), `Remove` (both
chainable), `Len`, `Has`, and `Build() Set[T]` with the same freeze,
publish-and-disable contract.

## `SortedMap[K,V]`

The comparator `cmp(a, b)` returns a negative, zero, or positive result and
must be a strict weak order that stays deterministic for the map's lifetime.
Keys for which `cmp` returns 0 are the same key. A nil comparator panics
`"immutable: nil comparator"`; a comparator panic propagates to the caller.

| Function / method | Behaviour |
|---|---|
| `EmptySortedMap[K,V](cmp func(a, b K) int) SortedMap[K,V]` | Empty `SortedMap` ordered by `cmp`. |
| `CollectSortedMap[K,V](cmp func(a, b K) int, seq iter.Seq2[K,V]) SortedMap[K,V]` | Drains `seq` through a builder; for equivalent keys the last key and the last value win. |
| `(m SortedMap[K,V]) Len() int` | Entry count. |
| `(m SortedMap[K,V]) IsEmpty() bool` | `Len() == 0`. Zero `SortedMap` is empty. |
| `(m SortedMap[K,V]) Get(k K) (V, bool)` / `Has(k K) bool` | Read access by an equivalent key. |
| `(m SortedMap[K,V]) Set(k K, v V) SortedMap[K,V]` | New `SortedMap` with `k` mapped to `v`; an equivalent stored key is replaced by `k`, and its value by `v`. `m` unchanged. Panics on the zero `SortedMap`. |
| `(m SortedMap[K,V]) Delete(k K) SortedMap[K,V]` | New `SortedMap` without `k`. `m` unchanged. The result keeps `m`'s comparator, even when emptied. No-op on the zero `SortedMap`. |
| `(m SortedMap[K,V]) All/Keys/Values() iter.Seq(2)[...]` | Ascending comparator order (`Values` in the order of their keys). |
| `(m SortedMap[K,V]) Equal(other SortedMap[K,V]) bool` | Same entry count, every key found (by `other`'s comparator) with a value equal via `eq.For[V]()`, selected once per call. No shared-root shortcut: even the same version compares its values. |
| `(m SortedMap[K,V]) EqualFunc(other SortedMap[K,V], eq func(V, V) bool) bool` | Same keys; values compared with `eq`. |
| `(m SortedMap[K,V]) Builder() *SortedMapBuilder[K,V]` | Seeds a `SortedMapBuilder[K,V]` with `m`'s root and comparator (opens a `transient.Window`). Panics on the zero `SortedMap`. |

[ADR-013](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md)
keeps `First`/`Last`/`Floor`/`Ceiling`, range queries and reverse iteration,
`SortedSet`, `SortedMap.Update`, and the engine's rank, split and join
operations out until a consumer needs them.

`SortedMapBuilder[K,V]` — mutable, not concurrency-safe — offers `Set` (same
key-replacement rule) and `Delete` (both chainable), `Len`, `Get`, `Has`, and
`Build() SortedMap[K,V]` with the same freeze, publish-and-disable contract.
A comparator panic during a builder `Set` or `Delete` leaves the builder's
prior contents intact.

### Lighter recipe: sort on read

`SortedMap` pays off when a map stays ordered across many versions. For a
one-off deterministic emission of a `Map` or `Set`, sort the iterator
instead: `slices.Sorted(s.Values())` when the element type is
`cmp.Ordered`, or `slices.SortedFunc(m.Keys(), cmp)` with any comparator.
That costs `O(n log n)` per read and needs no new collection.

---

## Measured: single update vs. Builder loop

Path-copy engines allocate along a **depth curve** (roughly O(log₃₂ n)
nodes), not a full `O(n)` collection copy. The property to watch in
`go test -bench` is **allocs/op** (and B/op) **not scaling linearly with n**
for a single `Map.Set` / `Set.Add` / `List.Append` / `SortedMap.Set`
against a pre-built base (the weight-balanced tree follows `log₂ n` rather
than `log₃₂ n`).

Looping `Append`/`Set`/`Add` on the immutable type still creates **n
intermediate roots**. `Builder` is the bulk path: exclusive ownership,
in-place edits on owned nodes, O(1) `Freeze`. Native mutable baselines
(`[]T` append, `map[K]V` assign / get) live in `immutable/bench_test.go` for
comparison.

```sh
cd functional
go test ./immutable/ -run '^$' -bench 'BenchmarkMapSet|BenchmarkMapUpdate|BenchmarkSetAdd|BenchmarkListAppend|BenchmarkListBuilder|BenchmarkMapBuilder|BenchmarkMapGet|BenchmarkSortedMap|BenchmarkNativeMap|BenchmarkSet(Union|Intersection|Difference)' -benchmem
```

For the Set algebra benches, the no-change cases (`Union` and
`Intersection` with the smaller operand a subset of the larger, `Difference`
with disjoint operands) report the same allocs/op for every size: they open
no builder and return an operand untouched.

`BenchmarkSortedMapGet`, `BenchmarkSortedMapSet`, and
`BenchmarkSortedMapBuilderSet` mirror the `Map` benches at the same sizes, so
the two engines compare directly.

Always route loop-building through `<Type>.Builder()`; never call
`Append`/`Set`/`Delete`/`Update`/`Add`/`Remove` in a loop against the
immutable type itself when linear bulk construction is the goal.

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [Internal: core](core.md)
- [ADR-004 — Thin immutability core](../adr/functional/adr-004-thin-immutability-core.md)
- [ADR-005 — Persistent engine family](../adr/functional/adr-005-persistent-engine-family.md)
- [ADR-013 — SortedMap sobre um motor de árvore balanceada por peso (wbtree)](../adr/functional/adr-013-sorted-map-weight-balanced-engine.md)
