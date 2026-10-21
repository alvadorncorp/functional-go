---
type: adr
id: adr:functional:005
context: functional
title: Persistent engine family — HAMT + 32-way vector, seed policy, O(1) Freeze
status: Aceita
tags: [functional, immutable, hamt, vector, ownership, maphash]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:004]
date: 2026-08-07
---
# ADR-005: Persistent engine family — HAMT + 32-way vector, seed policy, O(1) Freeze

> **Public package name:** the collections API lives at `functional/immutable` (formerly `functional/persistent`). This ADR still uses *persistent* for the structure-sharing engine family (HAMT + 32-way vector); that is the representation term, not the import path.

> **Forward note:** ordered maps add a third engine, `internal/engine/wbtree` (a weight-balanced tree behind `immutable.SortedMap`), recorded in [ADR-013](adr-013-sorted-map-weight-balanced-engine.md) (`adr:functional:013`). The HAMT and vector decisions and the O(1) Freeze below are unchanged.

## Context and problem

`immutable` List/Map/Set already expose functional updates, Builders, and
upper-bound log-n complexity as public-API headroom. Early milestones used
full-copy COW over a stdlib slice or map. That representation made every
immutable update `O(n)` and left Builders as the only linear bulk path.

Baseline B replaces those engines with structure-sharing representations
while keeping the thin immutability core (ADR-004) as **laws only** —
ownership, cow, sharing, transient — not as a generic clone/engine library.

## Decision drivers

- Single Map/Set update cost must follow a **depth curve** (path-copy), not
  a full collection copy.
- List needs ordered index access with structural sharing on Append/Set.
- Builder publish must stay **O(1)** (no tree walk on Freeze).
- Hashing must stay stdlib-only (`hash/maphash`); seeds must not invalidate
  structural sharing across a lineage.
- Public panic strings remain `"immutable: ..."`; engines must not leak
  `"vector:"` / `"hamt:"` messages.
- External `package immutable_test` remains the acceptance suite.

## Options considered

### Option 1: Keep full-copy COW

**Pros:** Tiny code; simple mental model.
**Cons:** Every `Set`/`Append` is `O(n)`; allocs scale with collection size;
documented log-n headroom is fiction forever.

### Option 2: Single unified RRB / finger-tree style for all three types

**Pros:** One engine package.
**Cons:** Heavier than needed for Map/Set; poorly matches hash-keyed
collections; higher implementation risk for Baseline B.

### Option 3: HAMT for Map/Set + 32-way vector for List (chosen)

**Pros:** Classic, well-understood pair; path-copy sharing; branching
factor 32 aligns with 5 hash/index bits; pairs cleanly with exclusive-owner
transient mutation and O(1) Invalidate publish.
**Cons:** Two engines to maintain; List.Delete left as O(n) rebuild for now.

## Decision

**Chosen: Option 3.**

### Engine layout

| Public type | Engine package | Representation |
|---|---|---|
| `Map[K,V]`, `Set[T]` | `internal/engine/hamt` | 32-way hash-array-mapped trie |
| `List[T]` | `internal/engine/vector` | 32-way persistent vector (trie + tail ≤ 32) |

- Pure immutable ops path-copy and publish new nodes with `owner == nil`.
- Transient ops take `*ownership.Owner` from a `transient.Window` and mutate
  exclusively owned nodes in place; engines do **not** import `transient`.
- Hashing: `internal/engine/keyhash` wraps `maphash.Comparable`. Each
  independent empty `hamt.Root` holds its own `maphash.Seed`; every derived
  version **preserves** that seed so structural sharing remains valid.

### Freeze / publish

`transient.Window.Freeze` is **O(1)**: `ownership.Owner.Invalidate()` plus
clearing the window owner. Nodes may retain inactive stamps; `IsOwnedBy`
fails for them, so no Detach walk is required at publish time.

### List.Delete

`List.Delete` (and the vector engine's `Delete`) is **currently O(n)** —
rebuild by iteration with correct shift-left semantics. Documented public
headroom may still list O(log n); callers MUST NOT depend on a tighter
constant until a path-aware Delete ships. Correctness over log-n Delete
was the Baseline B trade-off.

### Core remains thin

ADR-004 still holds: core = ownership / cow / sharing / transient laws.
Engines own representation. No `Clone[T]`, no `reflect` in core.

## Consequences

**Positive:** Single Map/Set/List updates allocate along tree depth; Builders
keep exclusive ownership and O(1) Freeze; seed-per-root preserves sharing;
archtests continue to gate zero non-stdlib deps and the core DAG.

**Negative:** Two engines; List.Delete is O(n); maintainers must not re-
introduce full-copy claims in package docs.

**Neutral:** Historical full-copy COW is retired as the engine (docs and
benchmarks track path-copy). ADR-004's thin-core layout is unchanged;
engine-specific claims in older prose are superseded by this ADR.
