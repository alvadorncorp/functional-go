---
type: tdd
id: tdd:modeling:reducer-replay
context: modeling
title: Modeling reducer and replay
status: Aceito
tags: [modeling, reducer, replay, immutable, generics, laws, testing]
related: [adr:modeling:001, adr:modeling:002, adr:modeling:004, adr:modeling:005, adr:modeling:006, tdd:modeling:policy]
date: 2026-08-10
---
# TDD: Modeling reducer and replay

## Technical scope

This delivery adds one externally importable capability package,
`github.com/alvadorncorp/gopackages/modeling/reducer`. Every Decision,
Transition, FallibleTransition, Invariant, Rule, Requirement, composition
function, method, root import path, `modeling/internal/decision` contract, `modeling/change` surface, and
`modeling/policy` surface remain unchanged.

Reducer models a fact-to-state transformation. It is distinct from Transition,
which models intent/input-to-output computation, from Decision, which carries a next state plus emitted events, from Change,
which evolves state without events, and from Policy, which derives a strategy
from context.

## Public contracts

The package adds exactly:

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events immutable.List[E],
    reduce Reducer[S, E],
) S
```

Ordinary functions with signature `func(S,E) S` are assignable and callable
without wrappers. Reducer returns state directly and has no error channel.
Replay accepts only `immutable.List[E]`; there are no slice, sequence, Decision,
or Apply overloads. Both type parameters remain unconstrained `any`,
including non-comparable state and event types.

## Replay semantics and laws

Replay is a synchronous left fold over facts in immutable-list encounter order:

```text
s1 = reduce(s0, e1)
s2 = reduce(s1, e2)
...
sn = reduce(s(n-1), en)
```

It invokes the reducer exactly once per event and never reverses or parallelizes
application. The zero list invokes the reducer zero times and returns the initial
state unchanged.

A nil reducer is not invoked for the zero list. For non-empty input it retains
ordinary Go nil-function invocation behavior and panics; Replay adds no guard or
alternate error contract.

The concatenation law is:

```text
Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)
```

Repeated replay returns equal state when the supplied reducer is pure and
deterministic. Purity is application responsibility; modeling performs no clock,
randomness, global-state, I/O, or side-effect detection.

## Event representation and Decision relationship

`E` remains unconstrained `any`. Applications own heterogeneous event unions,
switches, and unknown-kind behavior. modeling adds no marker interface,
reflection, registry, handler map, schema/version abstraction, or event hierarchy.

Decision and Reducer remain orthogonal. Applications whose Decision events are
state-reconstructing facts may write:

```go
events := immutable.CollectList(decisionValue.Events())
state := reducer.Replay(previous, events, reduce)
```

The library neither compares `Decision.State()` with Replay nor adds
`ReplayDecision`/`VerifyDecision`; integration events are not universally
state-reconstructing.

## Implementation and dependency boundary

Replay iterates `events.Values()` and threads a local state variable. It does
not call `events.Items()` or `fold.FoldLeft`, avoiding slice materialization.
The package imports only `functional/immutable` and standard library packages;
it imports no modeling root or sibling capability.

Architecture tests enforce:

- exactly two reducer exports, `Reducer` and `Replay`;
- the exact compile-time signatures;
- no modeling sibling dependency and no unapproved dependency closure;
- the public functional-type exception for reducer's pinned contract, alongside
  the existing policy Predicate/Optional exception;
- the name `Replay` only in `modeling/reducer`;
- exact root, change, policy, reducer, internal/archtest, and internal/decision
  package enumeration;
- direct `Values()` iteration without `Items()` or `fold.FoldLeft` materialization;
- no reverse functional-to-modeling edge, mutable production globals, or magic runtime surface.

## Excluded surface

There is no Apply, FallibleReducer, ReplaySlice, ReplayList, ReplaySeq, Reduce,
ReplayDecision, VerifyDecision, Event/DomainEvent marker, reflection, registry,
dispatcher, event stream loader, Event Store, snapshot store, projection,
serialization, version migration, persistence, transport, database, filesystem,
HTTP, broker, goroutine-based replay, dependency update, or global configuration.

## Verification

Run focused reducer and architecture tests, then module tests in workspace and
`GOWORK=off` published-dependency modes, race, vet, tidy-diff, uncapped lint, and
exact package listing. Finish with workspace `make test-unit`, `make test-race`,
`make lint`, `make vet`, and `git diff --check`. Database integration,
benchmark expansion, CI changes, dependency updates, tag creation, publication,
and release execution do not belong to this delivery.
