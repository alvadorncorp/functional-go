---
type: adr
id: adr:functional:functional-primitives-and-fold-boundary
context: functional
title: Functional primitives and the fold dependency boundary
status: Accepted
tags: [functional, generics, dependencies, fold]
deciders: [alvadorn]
related: [adr:functional:zero-dependency-boundary, adr:functional:values-family-dependency-boundary]
date: 2026-08-09
---
# Functional primitives and the fold dependency boundary

## Context and problem

The `functional` module needs four small public packages for function
composition, predicate algebra, total ordering, and slice reduction. The
first three can remain stdlib leaves. `fold`, however, has to represent
absence in `Reduce` and minimum cardinality in `ReduceNonEmpty` without
duplicating types already stabilized in `functional/values`.

The zero-dependency boundary ADR and the values family dependency boundary
ADR preserve the dependency-free frontier and record
`values/validated -> values/nonempty` as the only public edge needed at
that point. The new API forces a choice: relax that rule globally, or
authorize two new edges in a directed, testable way.

## Decision drivers

- Keep `function`, `predicate`, and `order` as standard-library leaves.
- Reuse `values/optional.Optional[T]` and `values/nonempty.NonEmpty[T]` in
  signatures instead of creating competing representations.
- Do not couple the primitives to `immutable`, `modeling`, internals,
  `reflect`, or `unsafe`.
- Preserve mechanically the direction and the exact set of public edges.

## Options considered

### Option 1: Keep fold independent and return `(T, bool)`

**Pros:**
- No new edge between public packages.

**Cons:**
- Abandons the typed integration required for `Reduce`.
- Duplicates at the call site semantics that `values/optional` already has.

### Option 2: Relax the rule to allow arbitrary imports between primitives

**Pros:**
- Makes future reuse easier without touching the archtest.

**Cons:**
- Creates coupling with no demonstrated need.
- Weakens the auditability of the DAG and allows future cycles.

### Option 3: Authorize only two fold edges to values

**Pros:**
- Reuses the stabilized absence and non-empty contracts.
- Keeps every other primitive a leaf and preserves an exact allow-list.
- Allows `ReduceNonEmpty` without carrying unneeded absence handling.

**Cons:**
- `fold` carries two types from other packages in its v0 API.
- The archtest and the DAG documentation must be maintained together.

## Decision

**Chosen: Option 3, authorize only two fold edges to values.**

`function`, `predicate`, and `order` depend only on the standard library.
`fold` may import exactly `functional/values/optional` and
`functional/values/nonempty`. No primitive imports another primitive.
`TestPublicDependencyDAG`, `TestPrimitiveDependencyDAG`, and
`TestPrimitivesNoReflectOrUnsafe` make the set executable.

The values family dependency boundary ADR remains the historical record and
the outbound rule of the values family:
only `values/validated` imports `values/nonempty`; no values package starts
importing `fold`. This decision also does not widen the pre-v1 library
signature restraint recorded in the zero-dependency boundary ADR, which stays
specific to the contract decided there; a general policy for all the
primitives would require another explicit decision.

## Consequences

**Positive:**
- `Reduce` represents absence in the type, and `ReduceNonEmpty` removes
  absence handling when the input already guarantees cardinality.
- The DAG stays small, directed, and mechanically verifiable.

**Negative:**
- Future changes to Optional or NonEmpty may require a coordinated change
  to fold's public API while the module is still v0.

**Neutral / Follow-up:**
- Rollback removes the fold package, its two allow-listed edges, and this
  ADR without affecting the three leaf packages.
- Traverse, Sequence, immutable adapters, and algebraic abstractions stay
  out of scope.

The tuples/iterators decision ([structural-primitives-and-iterator-boundary](structural-primitives-and-iterator-boundary.md))
is the later decision for tuple/lazy/iterator/optics. Its
`iterator -> tuple` edge does not relax the rules of this ADR for function,
predicate, order, or fold and must remain explicitly allow-listed.
