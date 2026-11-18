---
type: tdd
id: tdd:modeling:reducer-replay
context: modeling
title: Modeling reducer and replay
status: Accepted
tags: [modeling, reducer, replay, iter, generics, laws, testing]
related: [adr:modeling:minimal-type-safe-domain-modeling, adr:modeling:root-facade-internal-decision, adr:modeling:change-capability-boundary, adr:modeling:policy-capability-boundary, adr:modeling:reducer-capability-boundary, adr:modeling:streaming-replay-project-seq, tdd:modeling:policy]
date: 2026-08-12
---
# TDD: Modeling reducer and replay

## Technical scope

This design describes the importable package
`go.alvadorn.org/functional/modeling/reducer`. Decision, Transition,
FallibleTransition, Invariant, Rule, Requirement, composition, methods, the
root's import paths, `modeling/internal/decision`, `modeling/change`, and
`modeling/policy` remain separate contracts.

Reducer models the fact → state transformation. It is deliberately distinct
from Transition (intention/input → output), Decision (next state plus emitted
events), Change (state evolution without events), and Policy (strategy from
context).

The `Replay` events parameter is `iter.Seq[E]` (`adr:modeling:streaming-replay-project-seq`). The earlier
`immutable.List[E]` choice in `adr:modeling:reducer-capability-boundary` is
superseded for this parameter.

## Public contracts

The package exports exactly:

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S
```

Ordinary `func(S, E) S` functions are assignable and invocable without wrappers.
Reducer returns state directly and has no error channel. Replay accepts only
`iter.Seq[E]`; there are no slice, List, Decision, or Apply overloads. Both type
parameters remain unrestricted `any`, including non-comparable state and event
types. The semantics are value semantics: there is no `Reducer[*S, E]`,
`Ref[T]`, `Box`, or `Readonly` in the metamodel.

## Replay semantics and laws

Replay is a synchronous left fold in a single pass over the facts in the
sequence's encounter order:

```text
s1 = reduce(s0, e1)
s2 = reduce(s1, e2)
...
sn = reduce(s(n-1), en)
```

It invokes the reducer exactly once per event; it neither rolls back nor
parallelizes. An empty sequence invokes the reducer zero times and returns the
initial state.

A nil Reducer is not invoked on an empty sequence. For non-empty input,
ordinary Go behavior applies (panic on nil function call); Replay adds neither
a guard nor an alternative error contract.

Concatenation law (where the application defines `A ++ B` observably):

```text
Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)
```

Repeated Replay returns an equal state when the reducer is pure and
deterministic. Purity is the application's responsibility; modeling detects no
clock, randomness, global state, I/O, or side effects.

Because the sequence is single-pass, anyone needing multiple passes
materializes explicitly before calling Replay.

## Event representation and the Decision bridge

`E` remains unrestricted `any`. The application owns heterogeneous unions,
switches, and the unknown-kind policy. modeling adds no marker, reflection,
registry, handler map, schema/version, or event hierarchy.

Decision and Reducer stay orthogonal. When the Decision's events rebuild state,
the bridge is direct:

```go
state := reducer.Replay(previous, decisionValue.Events(), reduce)
```

`immutable.CollectList` is unnecessary. The library does not compare
`Decision.State()` with the Replay result and offers neither `ReplayDecision`
nor `VerifyDecision`; integration events are not universal state rebuilders.

## Implementation boundary and dependencies

Replay iterates `for event := range events` and threads a local state variable.
It materializes no `[]E`, calls neither `Items()` nor `CollectList`, and uses no
`fold.FoldLeft`.

Production imports of the package: only the standard library (`iter`). It
imports no `functional/*`, no modeling root, and no sibling capability.

Architecture tests must enforce:

- exactly two exports, `Reducer` and `Replay`;
- compiling signatures with `iter.Seq[E]`;
- no sibling dependency and no `functional/*` in production;
- the name `Replay` only in `modeling/reducer`;
- a direct range over `events` (single pass, no materialization);
- no functional→modeling reverse edge, no production mutable globals, and no
  magic runtime surface.

## Excluded surface

There is no Apply, FallibleReducer, ReplaySlice, ReplayList, ReplaySeq (beyond
the `iter.Seq` parameter itself), Reduce, ReplayDecision, VerifyDecision,
Event/DomainEvent marker, reflection, registry, dispatcher, stream loader,
event store, snapshot store, projection, serialization, version migration,
persistence, transport, database, filesystem, HTTP, broker, goroutine-based
replay, dependency update, or global configuration.

## Verification

Run focused reducer and architecture tests (including single pass and no
functional import), module tests in the workspace and with `GOWORK=off`, race,
vet, tidy-diff, lint without hidden caps, and the exact package listing. Close
with workspace `make test-unit`, `make test-race`, `make lint`, `make vet`, and
`git diff --check` when the delivery includes code. CI benchmark expansion,
tagging, and publication stay outside this documenting TDD.
