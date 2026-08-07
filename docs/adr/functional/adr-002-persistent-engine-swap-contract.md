---
type: adr
id: adr:functional:002
context: functional
title: Persistent engine swap contract — upper-bound complexity and Builder
status: Aceita
tags: [functional, persistent, complexity, builder]
deciders: [alvadorn]
related: [adr:functional:001]
date: 2026-08-07
---
# ADR-002: Persistent engine swap contract — upper-bound complexity and Builder

## Context and problem

`persistent.List`/`Map`/`Set` write functionally today over a copy-on-write
engine: every update copies the receiver's backing slice or map in full.
That engine is simple and correct, but a full copy per update is `O(n)`,
which is worse than what a structure-sharing engine (a HAMT for `Map`/`Set`,
an RRB tree for `List`) could offer. The package needed a public contract
that does not lock the implementation to copy-on-write forever, without
speculatively building a HAMT/RRB engine before there is a real need for
one.

## Decision drivers

- A future engine swap (copy-on-write → HAMT/RRB) must not be a breaking
  change for any caller of `persistent`.
- Callers must not be able to observe or depend on today's engine — its
  iteration order, its `==` identity, or its actual constant-factor
  performance — because any of those would make a future swap breaking.
- Building a collection element-by-element in a loop is a common pattern and
  must have a real linear (or better) path, not just a warning in a doc
  comment.
- The acceptance criterion for "the swap is safe" must be checkable
  mechanically, not just by code review.

## Options considered

### Option 1: Document today's copy-on-write complexity as the contract

**Pros:** Accurate today; no abstraction gap between docs and code.
**Cons:** Locks the API to copy-on-write's actual bounds forever — a future
HAMT/RRB engine with different (better, but structurally different)
constants could not honor the documented `O(1)` `At` without becoming
copy-on-write again in spirit; callers would rationally start depending on
today's exact behavior since the docs promise exactly that.

### Option 2: Ship a HAMT/RRB engine now

**Pros:** Gets the better asymptotic behavior immediately.
**Cons:** Meaningfully more implementation and testing surface for a module
whose current adopters have not yet demonstrated a workload where
copy-on-write's `O(n)` per update is the bottleneck; premature complexity
for unproven need.

### Option 3: Document an upper-bound contract; keep copy-on-write; add Builder now (chosen)

**Pros:** The public contract already anticipates the swap — complexity
bounds are the ceiling a future engine must honor, not today's floor;
`Builder` gives every caller who needs to build in a loop a real linear
path today, independent of which engine sits underneath; the test suite
itself (external `_test` package, asserting only public behavior) becomes
the acceptance gate for the swap.
**Cons:** Today's engine sometimes beats its own documented bound (`At` is
actually `O(1)`, not `O(log n)`) — callers must be told plainly not to
depend on that gap.

## Decision

**Chosen: Option 3.**

`persistent` documents complexity as an **upper bound** the eventual engine
must honor, not a description of copy-on-write's current constant:

| Operation | Documented bound |
|---|---|
| `Len` | `O(1)` |
| `At`, `Get`, `Has` | `O(log n)` |
| `Set`, `Delete` | `O(log n)` |
| `Append` | `O(log n)` amortized |

Three further rules make the eventual swap possible without a breaking
change:

- **`Map`/`Set` iteration order is unspecified** and may change between
  releases, or even between two calls on the same engine. `List` iteration
  order is specified (index order) and is part of the contract, since
  ordering is intrinsic to what a list is.
- **None of `List`, `Map`, `Set` are `==`-comparable.** Comparing by `==`
  would compare identity of the internal representation — exactly what
  structural sharing would change. `Equal` is the only comparison any of
  the three types offer.
- **`Builder()` is the API that survives the swap.** `<Type>.Builder()`
  seeds a mutable, non-concurrency-safe builder from a snapshot; batch
  writes go through it; `Build()` hands back an immutable collection and
  permanently disables the builder. This is the same shape a HAMT/RRB
  *transient* would expose, so code that already builds through `Builder`
  in a loop does not change at all across the swap — only the constant
  factor underneath it does.

**The acceptance criterion for a future engine swap is the test suite.**
Every test in `persistent` lives in `package persistent_test` (external),
asserting only publicly observable behavior — never the internal
representation. A future swap from copy-on-write to a HAMT/RRB engine is
correct exactly when this suite continues to pass **unchanged**. No test
may be edited to make the swap "pass"; if one needs to change, the swap was
not behavior-preserving.

### Evidence: measured Append-loop vs. Builder

Measured on Apple M2 Max (`go test -run '^$' -bench
'BenchmarkListAppendLoop|BenchmarkListBuilderAppendLoop' -benchmem
./persistent/...`), building a `List[int]` of size `n` from empty:

| n | `Append` in a loop | `Builder` in a loop |
|---:|---|---|
| 16 | 280.5 ns/op, 1,136 B/op, 16 allocs/op | 110.1 ns/op, 248 B/op, 5 allocs/op |
| 1,024 | 436,230 ns/op, 4,470,541 B/op, 1,024 allocs/op | 4,827 ns/op, 25,208 B/op, 12 allocs/op |
| 65,536 | 1,089,195,208 ns/op, 17,435,164,136 B/op, 66,009 allocs/op | 529,275 ns/op, 2,512,152 B/op, 26 allocs/op |

At `n = 65,536`, looping `Append` costs over a second of wall time and 17 GB
of allocation; `Builder` costs 529 microseconds and 2.5 MB — roughly a
2,000× time gap and a 6,900× memory gap, growing with `n` because one path
is quadratic (`1 + 2 + … + n` full copies) and the other is linear. This is
the concrete reason `Builder` exists today, under copy-on-write, independent
of whatever engine eventually sits underneath it. See
[packages/persistent.md](../../packages/persistent.md#measured-append-loop-vs-builder)
for the full table and methodology.

## Consequences

**Positive:** The public contract already has room for a HAMT/RRB engine
without a major version bump or a breaking API change; callers who build in
a loop have a real linear path today (`Builder`), not just a documented
pitfall; the swap's correctness bar is objective and automated (the existing
external test suite passing unchanged).

**Negative:** The documented complexity bounds are sometimes pessimistic
versus today's actual engine (e.g. `At` is `O(1)` today, not the documented
`O(log n)`) — a caller that benchmarks against today's constant and encodes
an assumption on it is depending on something this contract explicitly does
not promise.

**Neutral:** No HAMT/RRB engine is being built now; this ADR only fixes the
contract shape so that work, if and when it happens, is additive rather than
breaking.
