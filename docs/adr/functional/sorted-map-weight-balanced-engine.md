---
type: adr
id: adr:functional:sorted-map-weight-balanced-engine
context: functional
title: SortedMap on a weight-balanced tree engine (wbtree)
status: Accepted
tags: [functional, immutable, sorted-map, wbtree, engine, ownership]
deciders: [alvadorn]
related: [adr:functional:thin-immutability-core, adr:functional:persistent-engine-family]
date: 2026-10-21
---
# SortedMap on a weight-balanced tree engine (wbtree)

## Context and problem

`immutable.Map` and `immutable.Set` use the HAMT from the persistent engine
family ADR: the iteration order is unspecified and changes across lineages,
because each root holds its own `maphash.Seed`. There is no immutable
collection with a user-defined key order.

Item 3b (a sorted immutable map) was approved by the owner. Being honest
about the demand: today no consumer of `Map` needs order, and the case cited
in the proposal (emitting the elements of a `Set` in deterministic order) is
already served by `slices.Sorted(s.Values())`, or `slices.SortedFunc` when
`T` is not `cmp.Ordered`. That remains the lightest path for a one-off
emission. `SortedMap` exists for whoever keeps a map sorted across many
versions, without re-sorting on every read.

## Decision drivers

- `O(log n)` update and read, with path-copy and structural sharing, like
  `Map`.
- Builders with transient editing under `*ownership.Owner` and `O(1)` Freeze
  (persistent engine family), with no tree walk to publish.
- Thin core untouched (thin immutability core): the engine imports only
  `ownership` and `cow`.
- Zero dependencies outside the stdlib, no `reflect`, no `unsafe`.
- The smallest balanced tree whose balancing has a published proof.

## Options considered

### Option 1: Persistent B-tree

Rejected for now: better cache locality, but wide-node split/merge with a
transient owner is considerably more code, and each path-copy copies whole
nodes. It stays as a possible engine swap, without changing the API, only if
`SortedMap` benchmarks show `Get`/`Set` more than roughly 3x slower than
`Map` with 10k keys.

### Option 2: Persistent red-black or AVL

Rejected: persistent deletion in red-black has many cases, and neither tree
stores the subtree size. A weight-balanced tree stores it for free, which
gives `Len` in `O(1)` and keeps rank, split, and join possible in the future.

### Option 3: Sort only on read

Keep `Map` and sort on every iteration. Costs `O(n log n)` per read and
creates no new type. Rejected as the only answer to item 3b, but it remains
the recommended recipe for one-off emissions (`slices.Sorted`).

### Option 4: Weight-balanced tree with Straka's parameters (chosen)

## Decision

Add `immutable.SortedMap[K, V]` and `immutable.SortedMapBuilder[K, V]` on a
new internal engine `internal/engine/wbtree`.

### The `wbtree` engine

A persistent weight-balanced binary search tree (Adams), with the parameters
proven by Straka ("Adams' Trees Revisited", 2011) over the weight `size+1`,
as named constants:

```
delta = 3
gamma = 2

isBalanced(a, b) := delta*(size(a)+1) >= size(b)+1
isSingle(a, b)   := size(a)+1 < gamma*(size(b)+1)
```

Every node satisfies `isBalanced(left, right)` and
`isBalanced(right, left)`. After an insertion or deletion, each level of the
path is repaired with a single rotation (when
`isSingle(inner, outer)` holds for the grandchildren on the heavy side) or a
double one. Rotations look only at sizes, never at the comparator.

- Pure operations path-copy and publish nodes with `owner == nil`.
- `TransientSet`/`TransientDelete` receive a `*ownership.Owner`, edit in
  place the nodes actively owned by it, and clone the others via
  `cow.EnsureEditable`. The engine does not import `transient`; publishing is
  `Invalidate` in `O(1)`.
- Every comparison along a path happens on the way down, before any node is
  edited on the way back. A comparator that panics leaves no half-edited
  owned nodes behind, and the builder keeps the previous content.
- Deleting an absent key returns the same node pointer.
- Allowed imports: `ownership` and `cow`. The `functional` archtest
  registers the engine and that allow-list.

### Comparator contract

`cmp func(a, b K) int` returns negative, zero, or positive and must be a
strict weak order: deterministic, transitive, and stable for the life of the
map. Keys with `cmp(a, b) == 0` are the same key. The comparator is fixed at
construction and inherited by every derived version, including one emptied
by `Delete`. `EmptySortedMap(nil)` and `CollectSortedMap(nil, seq)` panic
with `immutable: nil comparator`. A comparator panic propagates to the
caller. Iteration (`All`, `Keys`, `Values`) is ascending by the comparator;
the order of `Map` and `Set` remains unspecified.

### Zero-value exception

`K any` has no default order, so a zero `SortedMap` has no comparator. It
reads as empty (`Len`, `Get`, `Has`, iteration) and `Delete` on it is a
no-op, but `Set` and `Builder()` panic with exactly:

```
immutable: zero SortedMap has no comparator; construct with EmptySortedMap or CollectSortedMap
```

This is a deliberate exception to parity with `Map` and `Set`, whose zero
value accepts writes. A zero-value-safe getter needs to return
`EmptySortedMap(cmp)`. `wbtree:` messages never reach the user:
`immutable` validates before calling the engine.

### Key replacement rule

`Set` replaces the stored key and the value when an equivalent key already
exists, on every path (pure, builder, and `CollectSortedMap`). The last key
and the last value win, and that is observable through `Keys()` with a
case-insensitive comparator, for example.

### API scope

Constructors `EmptySortedMap(cmp)` and `CollectSortedMap(cmp, seq)`; methods
`Len`, `IsEmpty`, `Get`, `Has`, `Set`, `Delete`, `All`, `Keys`, `Values`,
`Equal`, `EqualFunc`, and `Builder`; the builder has `Set`, `Delete`, `Len`,
`Get`, `Has`, and `Build`, which is `O(1)`. Out of scope: `First`, `Last`,
`Floor`, `Ceiling`, range queries, and reverse iteration; `SortedSet`;
`SortedMap.Update`; and the engine's rank, split, and join operations. They
are additive and have no consumer today.

## Consequences

**Positive:** a sorted collection with `O(log n)` path-copy, structural
sharing, and `O(1)` Freeze, without touching the core or the existing
engines. The key rule stays unique and tested. The per-node size keeps rank
and split possible without swapping engines.

**Negative:** a third engine to maintain. The zero value of `SortedMap`
breaks parity with `Map` and `Set` on writes. Inconsistent comparators
corrupt the order with no diagnostic. A binary tree has worse cache locality
than a HAMT or a B-tree.

**Verification:** an invariant checker with the exact predicates after every
operation, ascending and descending sequential tests with roughly 10k keys,
`FuzzOps` against a model, transient and sharing tests, an allocation cap on
one `Set`, acceptance tests in `immutable_test`, the engine dependency
archtest, `go test ./...`, and lint.
