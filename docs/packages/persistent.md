# Package `persistent`

**Import path:** `github.com/alvadorncorp/gopackages/functional/persistent`

Immutable collections — `List[T]`, `Map[K,V]`, `Set[T]` — that write
functionally: every update method returns a new collection and leaves the
receiver unchanged. There is no in-place mutation anywhere in this API;
"write" always means "derive a new value from an old one."

The engine underneath is **full-copy** copy-on-write over a stdlib slice or
map **today**. Builders open an internal `transient.Window` (see
[core.md](core.md)) and freeze it on `Build`. A future structure-sharing
engine may use more of the thin core (`ownership` / `cow` / `sharing`);
documented log-n complexity bounds below are **aspirational public-API
headroom**, not a description of the current engine. Decision record:
[ADR-004](../adr/functional/adr-004-thin-immutability-core.md).

---

## Responsibilities

| Area | Types / APIs |
|---|---|
| Ordered collection | `List[T]`, `NewList`, `CollectList`, `ListBuilder[T]` |
| Key-value collection | `Map[K,V]`, `NewMap`, `CollectMap`, `MapBuilder[K,V]` |
| Set collection | `Set[T]`, `NewSet`, `CollectSet`, `SetBuilder[T]` |
| Internal substrate | Builders use `internal/core/transient`; full-copy engines do not yet stamp node owners |

---

## The contract every type follows

1. **No method accepts or returns the internal representation.** `Items()`
   always copies rather than handing out the backing slice or map.
2. **`Map`/`Set` iteration order is UNSPECIFIED** and may change between
   releases, or even between two calls on the same value. `List` iterates
   in index order — that order is specified and part of the contract.
3. **Documented complexity is an upper bound** the public API reserves room
   for, not a description of what full-copy COW does today (see the table
   below).
4. **None of the three types are `==`-comparable.** Use `Equal`. Comparing
   by `==` would compare identity of the underlying representation.
5. **A single update is a full copy: `O(n)` with today's engine.** Calling
   `Append`/`Set`/`Delete` `n` times in a loop is therefore `O(n²)`.
   `Builder()` exists for that case — see [the benchmark](#measured-append-loop-vs-builder)
   below. Builders open a `transient.Window` and `Freeze` on `Build`; public
   panics remain `"persistent: ..."`.
6. **Every test in this package is `package persistent_test`** (external),
   asserting only publicly observable behavior — the acceptance criterion
   for any future engine change is this suite passing unchanged.

---

## Documented complexity (upper bound, not today's constant)

| Operation | Documented bound |
|---|---|
| `Len` | `O(1)` |
| `At`, `Get`, `Has` | `O(log n)` |
| `Set`, `Delete` | `O(log n)` |
| `Append` | `O(log n)` amortized |

The current full-copy engine beats several of these today (`At` is `O(1)`
right now, for instance) — a caller **must not** depend on that. Only the
bound above is guaranteed as API headroom. The thin core does not implement
these bounds; it only provides ownership/window/publish primitives for
engines that might. See [ADR-004](../adr/functional/adr-004-thin-immutability-core.md)
and [core.md](core.md).

---

## `List[T]`

| Function / method | Behaviour |
|---|---|
| `NewList[T](items ...T) List[T]` | Copies `items`, in order. |
| `CollectList[T](seq iter.Seq[T]) List[T]` | Drains `seq` into a new `List`, in iteration order. |
| `(l List[T]) Len() int` | Element count. |
| `(l List[T]) At(i int) T` | Element at `i`; panics `"persistent: index out of range"` outside `[0, Len)`. |
| `(l List[T]) Append(items ...T) List[T]` | New `List` with `items` appended. `l` unchanged. Full copy: `O(n)` per call. |
| `(l List[T]) Set(i int, v T) List[T]` | New `List` with index `i` replaced. Panics as above. |
| `(l List[T]) Delete(i int) List[T]` | New `List` without index `i`; later elements shift down. Panics as above. |
| `(l List[T]) Values() iter.Seq[T]` | Elements in index order. |
| `(l List[T]) All() iter.Seq2[int, T]` | (index, element) pairs in index order. |
| `(l List[T]) Items() []T` | Fresh mutable copy; caller owns it. |
| `(l List[T]) Equal(other List[T]) bool` | Same elements, same order, via `functional/internal/eq.Equal`. |
| `(l List[T]) Builder() *ListBuilder[T]` | Seeds a `ListBuilder[T]` with a copy of `l`'s elements and opens a `transient.Window`. |

`ListBuilder[T]` — **mutable, not safe for concurrent use** — offers
`Append`, `Set`, `Delete` (all chainable, same panic messages as `List`),
`Len`, and `Build() List[T]`. `Build` freezes the builder's transient
window, transfers ownership of the builder's buffer to the returned `List`,
and permanently disables the builder: every method afterward, including a
repeated `Build`, panics `"persistent: builder used after Build"`.

## `Map[K,V]`

| Function / method | Behaviour |
|---|---|
| `NewMap[K,V](m map[K]V) Map[K,V]` | Copies `m`. |
| `CollectMap[K,V](seq iter.Seq2[K,V]) Map[K,V]` | Drains `seq`; last value wins on a repeated key, matching `maps.Collect`. |
| `(m Map[K,V]) Len() int` / `Get(k K) (V, bool)` / `Has(k K) bool` | Read access. |
| `(m Map[K,V]) Set(k K, v V) Map[K,V]` | New `Map` with `k` mapped to `v`. `m` unchanged. |
| `(m Map[K,V]) Delete(k K) Map[K,V]` | New `Map` without `k`. `m` unchanged. |
| `(m Map[K,V]) All/Keys/Values() iter.Seq(2)[...]` | Unspecified iteration order. |
| `(m Map[K,V]) Items() map[K]V` | Fresh, non-nil copy. |
| `(m Map[K,V]) Equal(other Map[K,V]) bool` | Same key set, each value equal via `functional/internal/eq.Equal`. |
| `(m Map[K,V]) Builder() *MapBuilder[K,V]` | Seeds a `MapBuilder[K,V]` (opens a `transient.Window`). |

`MapBuilder[K,V]` — mutable, not concurrency-safe — offers `Set`, `Delete`
(chainable), `Len`, and `Build() Map[K,V]` with the same freeze,
transfer-and-disable contract as `ListBuilder`.

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
transfer-and-disable contract.

---

## Measured: Append-loop vs. Builder

This is the argument for using `Builder` in a loop, not `Append` repeated.
Measured on Apple M2 Max (`go test -run '^$' -bench
'BenchmarkListAppendLoop|BenchmarkListBuilderAppendLoop' -benchmem
./persistent/...`), each benchmark builds a `List[int]` of size `n` from
empty, one element at a time:

| n | `Append` in a loop | `Builder` in a loop | Time ratio | Memory ratio |
|---:|---|---|---:|---:|
| 16 | 280.5 ns/op, 1.1 KB/op, 16 allocs/op | 110.1 ns/op, 248 B/op, 5 allocs/op | ~2.5× | ~4.6× |
| 1,024 | 436.2 µs/op, 4.47 MB/op, 1,024 allocs/op | 4.8 µs/op, 25.2 KB/op, 12 allocs/op | ~90× | ~177× |
| 65,536 | 1.089 s/op, 17.4 GB/op, 66,009 allocs/op | 529.3 µs/op, 2.51 MB/op, 26 allocs/op | ~2,058× | ~6,941× |

**This is the `O(n²)` cliff, made concrete.** `Append` in a loop does
`1 + 2 + … + n` full copies of the growing `List`; at `n = 65,536` that is
over a second of wall time and 17 GB of allocation to build one collection —
against 529 microseconds and 2.5 MB through `Builder`. The gap grows with
`n` precisely because one path is quadratic and the other is linear. Always
route loop-building through `<Type>.Builder()`; never call
`Append`/`Set`/`Delete`/`Add`/`Remove` in a loop against the immutable type
itself.

---

## Related documents

- [Overview](../overview.md)
- [Examples](../examples.md)
- [Internal: core](core.md)
- [Package API: frozen](frozen.md) (the read-only sibling)
- [ADR-004 — Thin immutability core](../adr/functional/adr-004-thin-immutability-core.md)
