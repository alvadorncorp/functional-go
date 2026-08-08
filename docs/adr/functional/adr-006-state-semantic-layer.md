---
type: adr
id: adr:functional:006
context: functional
title: State semantic layer — pure State[T], mesh deps, lineage and single-use Tx
status: Aceita
tags: [functional, state, lineage, version, transaction, mesh]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:004, adr:functional:005]
date: 2026-08-07
---
# ADR-006: State semantic layer — pure State[T], mesh deps, lineage and single-use Tx

## Context and problem

`functional` already ships `optional`, `persistent`, and `valueobject` as
public packages under a zero-dependency module boundary (ADR-003), a thin
immutability core (ADR-004), and structure-sharing engines behind
`persistent` (ADR-005). Callers still lack a **semantic evolution layer**:
a way to treat an arbitrary pure value `T` as versioned state with lineage,
linked plans, single-use transactions, and structural diffs — without
importing domain rules into the module and without coupling that layer to
collection engines or to other public packages.

The risk of putting this surface on `persistent` (or inside core/engine) is
clear: state is about **semantic versions of T**, not about List/Map/Set
representation. Mixing those concerns would either force domain-shaped APIs
into the collections package or pull core/engine imports into a package that
should stay mesh-independent.

## Decision drivers

- `State[T]` must be **pure over T**: hold and evolve T without interpreting
  domain fields or embedding domain packages.
- Production **mesh sibling law**: no public subpackage imports another
  (`state` ↛ `persistent` / `optional` / `valueobject`, and the reverse).
- Production `state` must **not** import `internal/core` or
  `internal/engine`; engines remain behind `persistent`.
- Lineage and **Version** identity must support linked plans (compare /
  join only within a shared root).
- **Tx** is single-use: after Commit, further use panics with a stable
  `"state: …"` string.
- **Diff** must not use `reflect` (or `unsafe`) in production code.
- Public panics use the `"state: …"` prefix only.
- Archtests gate the mesh, zero non-stdlib deps, and no reflect/unsafe in
  production `state`.

## Options considered

### Option 1: Extend `persistent` with versioned wrappers over collections

**Pros:** One public package for “immutable data.”
**Cons:** Couples semantic evolution of arbitrary T to List/Map/Set engines;
tempts core/engine imports into a mixed API; blurs collection vs state
contracts.

### Option 2: Place state primitives in `internal/core`

**Pros:** Shared laws next to ownership/cow/transient.
**Cons:** Core is intentionally thin and representation-free (ADR-004);
state is a public semantic product, not a mutation-permission law.

### Option 3: New public package `functional/state` (chosen)

**Pros:** Clear ownership of pure `State[T]`, lineage, Version, Tx, Diff;
mesh isolation from `persistent`; no core/engine dependency in production;
tests may still import `persistent` without weakening production mesh
gates.
**Cons:** Another public package to document and archtest; callers compose
`state` with `persistent` at the application layer rather than via imports
inside `state`.

## Decision

**Chosen: Option 3** — introduce `github.com/alvadorncorp/gopackages/functional/state`
as the semantic evolution layer for `T any`.

### Pure State[T]

- `State[T]` evolves pure values of type T. The package does not define
  domain types, validation, or business rules.
- Semantic immutability of the payload is as deep as T itself (same caveat
  as `persistent` element types).
- Shipped construction and evolution:
  - `Of[T](value T) State[T]` — Version 0, empty plan, fresh lineage token.
  - Zero `State[T]` is valid (zero `T`, Version 0, nil lineage).
  - `Mutation[T] = func(T) T` — pure evolution step.
  - `Apply` records one reverse-linked plan node in O(1) and does **not**
    invoke the mutation; Version advances by one on the returned State.
  - `Materialize` re-applies the full plan oldest→newest on every call
    (**no cache**).
  - `Snapshot` freezes Version and the once-materialized Value.

### Mesh dependencies

- Production imports: stdlib only among external modules; no sibling public
  packages; no `internal/core`; no `internal/engine`.
- `TestNoPublicSiblingDependencies` remains absolute — **no**
  `state` → `persistent` production exception.
- Tests may import `persistent` to exercise composition; archtests continue
  to use production-only `go list` (without `-test`).

### Linked plan, Version, Tx, Diff (shipped)

| Concern | Decision |
|---|---|
| Lineage / Version | Unexported lineage token per `Of` root; `Version` is a monotonic `uint64` counter. Operations across unrelated lineages panic with `"state: unrelated lineages"`. Version is **not** a DB revision. |
| Linked plan | Reverse-linked immutable plan nodes; sibling branches share the parent prefix by pointer. |
| Tx single-use | `Begin` / `Transaction.Apply` / `Commit`. Commit is lazy (no Materialize). After Commit, further use panics with `"state: transaction used after Commit"`. Tx ≠ `persistent.Builder`. |
| Diff | `DiffBetween(from, to)` captures plan steps when `to` is a descendant of `from` in the same lineage; `Diff.Apply` replays lazily. Diff ≠ domain event. No production `reflect` or `unsafe`. |
| Mutation form | Pure `func(T) T` — not in-place mutation of T. |
| Panic prefix | All public panics: `"state: …"`. |

### Explicit non-goals (this ADR)

- No domain model inside `functional/state`.
- No production dependency on `persistent`, `optional`, or `valueobject`.
- No production dependency on core/engine packages.
- No reflect-based Diff.
- No materialize result cache on `State`.

## Consequences

**Positive:** Clear package boundary for versioned pure T; mesh and
zero-dependency gates stay enforceable; `persistent` remains the collection
engine surface; Of/Apply/Materialize/Snapshot/Tx/DiffBetween ship without
reopening mesh or core laws.

**Negative:** Application code must compose `state` with collections or
domain types itself; another public package to maintain.

**Neutral:** Status is **Aceita** with the shipped API above; related ADRs
003/004/005 remain in force for module deps, thin core, and engines.
