---
type: adr
id: adr:modeling:streaming-replay-project-seq
context: modeling
title: Replay and Project over iter.Seq with a stdlib-only dependency
status: Accepted
tags: [modeling, reducer, projection, iter, streaming, compatibility, stdlib]
deciders: [alvadorn]
related: [adr:modeling:reducer-capability-boundary, adr:modeling:projection-capability-boundary, tdd:modeling:reducer-replay, tdd:modeling:projection]
date: 2026-08-12
---
# Replay and Project over iter.Seq with a stdlib-only dependency

## Context and problem

[The reducer capability boundary](reducer-capability-boundary.md) and
[the projection capability boundary](projection-capability-boundary.md)
established the `modeling/reducer` and `modeling/projection` capabilities with
synchronous folds over facts that have already occurred. Both accepted
`immutable.List[E]` in the `events` parameter and imported
`functional/immutable` only for that contract.

`Decision[S,E].Events()` already exposes `iter.Seq[E]`. To replay or project
events from a Decision, the application had to materialize the sequence with
`immutable.CollectList` before calling `Replay`/`Project`. That step:

- allocates an unnecessary intermediate list for a single-pass fold;
- couples reducer and projection to `functional/immutable` with no semantic gain;
- blocks using any `iter.Seq[E]` directly (lazy streams, on-demand
  concatenation, test generators) without prior materialization.

Shipped code already adopts `iter.Seq[E]` and removes the production dependency
on `functional/immutable` in both packages. This ADR records the public-contract
decision and the partial supersession of the reducer and projection capability
boundaries regarding the `events` parameter.

## Decision drivers

- preserve synchronous left-fold semantics and encounter order;
- eliminate intermediate materialization in `Replay` and `Project`;
- allow `reducer.Replay(initial, decision.Events(), reduce)` and the projection
  equivalent without `CollectList`;
- keep `Reducer` and `Projection` as `func` types with value semantics;
- restrict production imports of reducer/projection to the stdlib (`iter`);
- not weaken lint (`hugeParam` and the other rules) or introduce pointers into
  the metamodel to sidestep the copy cost of large values;
- document the public-contract break already present in `modeling` tags up to
  `v0.6.0`.

## Alternatives considered

### Keep `immutable.List[E]` (the previous status quo)

Rejected. It forces materialization when the source is already `iter.Seq[E]`
and keeps unnecessary coupling to `functional/immutable`.

### Accept `[]E` or List + Seq overloads

Rejected. It multiplies the public surface, favors copying into a slice, and
dilutes the single one-pass contract.

### `fold.FoldLeft` over a materialized slice

Rejected. `FoldLeft` operates on `[]T` and requires `Items()` or an equivalent
conversion, reintroducing the allocation and dependency this decision removes.

### `Reducer`/`Projection` as pointer or `immutable.Ref[T]` / `Box` / `Readonly`

Rejected. The metamodel stays with functions and values; the cost of large
states is the application's problem (its own representation), not a reason to
pointerize the capability type. `hugeParam` remains a lint guardrail over the
size of concrete values, not a justification for changing the metamodel's shape
or for `//nolint`.

### Disabling or suppressing `hugeParam` / lint rules

Rejected. Lint stays effective; fixing large values happens in the
application's modeling, not by suppression in the package.

### `iter.Seq[E]` with a single-pass fold (chosen)

Chosen: one sequence parameter, exactly-once consumption, zero intermediate
collection, and a direct bridge with `Decision.Events()`.

## Decision

The public contracts keep exactly two exports per package, with the events
parameter as `iter.Seq[E]`:

```go
// modeling/reducer
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S

// modeling/projection
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Semantics:

- synchronous single-pass left fold: each event is consumed exactly once, in
  the sequence's encounter order;
- an empty sequence returns `initial` without invoking the reducer/projection;
- no materialization of `[]E`, `immutable.List[E]`, or any other intermediate
  collection inside `Replay`/`Project`;
- callers needing multiple passes materialize explicitly **before** calling
  the API;
- Decision bridge: `Replay(s, decision.Events(), r)` and
  `Project(p, decision.Events(), f)` without `CollectList`.

Production dependency:

```text
application → modeling/reducer    → stdlib (iter)
application → modeling/projection → stdlib (iter)
```

Reducer and projection import no `functional/*`, no modeling root, and no
sibling capabilities. The direction `application → modeling root →
modeling/internal/decision → functional/immutable` remains for Decision;
only the `events` parameter of Replay/Project stops requiring a List.

This ADR **supersedes** the `immutable.List[E]` choice in the reducer and
projection capability boundaries for the `events` parameter. Still in force:
package boundaries, nominal identity of `Reducer`/`Projection`, ES/CQRS runtime
exclusions, absence of root re-exports, and independence between the two
capabilities.

## Compatibility and rollback

The switch from `immutable.List[E]` to `iter.Seq[E]` and the removal of the
`functional/immutable` import in reducer/projection are a **public-contract
break**. `modeling` tags up to `modeling/v0.6.0` carry the previous List-based
contract; consumers of those tags must adapt their calls (provide
`iter.Seq[E]`, typically via `Events()` or `list.Values()`).

Rollback after adoption is a new compatibility decision and a major/minor
SemVer release per the module's policy. No dual List+Seq API is authorized to
"soften" the break on this line.

## Consequences

- Applications avoid `CollectList` on the Decision → Replay/Project path.
- Any `iter.Seq[E]` (lazy or materialized via `.Values()`) is valid input; the
  encounter order is the supplied sequence's order.
- Sequence concatenation and multi-pass are the application's responsibility;
  concatenation laws remain valid when the application defines `A ++ B`
  observably.
- `hugeParam` and the rest of lint stay active: large state or projection values
  are handled in the application's modeling, without `Reducer[*S, E]`,
  `Ref[T]`, or `//nolint` in the metamodel.
- The surface excluded by the reducer and projection capability boundaries
  stays excluded (Apply, Fallible*, registries, persistence, ES/CQRS runtime,
  etc.).
