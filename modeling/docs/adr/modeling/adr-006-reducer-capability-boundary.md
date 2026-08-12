---
type: adr
id: adr:modeling:006
context: modeling
title: Additive reducer capability boundary
status: Aceita
tags: [modeling, reducer, replay, packages, compatibility, immutability]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:002, adr:modeling:003, adr:modeling:004, adr:modeling:005, adr:modeling:010, tdd:modeling:reducer-replay]
date: 2026-08-10
---
# ADR-006: Additive reducer capability boundary

> **Supersessão parcial (ADR-010):** o parâmetro `events` de `Replay` passou de
> `immutable.List[E]` para `iter.Seq[E]`. A dependência de produção de
> `modeling/reducer` é somente stdlib (`import "iter"`). A conversão via
> `immutable.CollectList(decision.Events())` não é mais necessária; use
> `Replay(initial, decision.Events(), reduce)` diretamente. O restante desta
> ADR — boundary do pacote, tipo `Reducer[S,E]`, exclusões de runtime e
> independência de siblings — permanece em vigor. Ver
> [ADR-010](adr-010-streaming-replay-project-seq.md).

## Context and problem

Applications need to evolve state from facts that have already occurred and to
reconstruct state by applying an ordered event collection. This behavior is a
pure domain-modeling capability, not a decision, transition, validation hook,
policy strategy, or Event Sourcing runtime.

The shipped module keeps Decision, Transition, Invariant, Rule, and Requirement
as canonical root-package contracts, keeps pure state evolution in
`modeling/change` (ADR-004), and keeps Policy in `modeling/policy` (ADR-005).
ADR-001 also keeps functional collection types out of public modeling signatures
by default, while the settled Replay contract is explicitly defined over
`immutable.List[E]`. The new capability therefore needs a bounded package and a
narrow, mechanically enforced exception without moving or weakening existing
contracts. This decision is ADR-006 because ADR-004 already records Change.

## Decision drivers

- preserve every shipped root, change, and policy import path, type identity, signature, and behavior;
- model `State × Event → State` with an ordinary assignable Go function type;
- replay immutable events strictly in encounter order without copying to a slice;
- expose exactly one coherent capability package rather than one package per symbol;
- keep Decision and heterogeneous event representation application-owned;
- exclude reflection, registries, dispatch, persistence, transport, and Event Sourcing runtime concerns.

## Options considered

### Move every modeling capability to a public subpackage

Rejected. That source-document direction was corrected as a mistake and would
break accepted root contracts, import paths, and defined type identities.

### Keep Reducer and Replay in the root package

Rejected. The root remains the compatibility surface for shipped capabilities;
adding replay there would resume flat-package growth and blur the distinct
fact-to-state capability.

### Accept `iter.Seq[E]` instead of `immutable.List[E]`

Rejected for this delivery. It would integrate directly with `Decision.Events()`,
but it would change the settled Replay signature. Applications may explicitly
convert with `immutable.CollectList(decision.Events())` when that relationship
applies.

### Add one public reducer package with a narrow immutable-list exception (chosen)

Chosen because it adds the requested capability without changing existing
types. `immutable.List.Values()` supplies ordered iteration directly; using
`fold.FoldLeft` would first copy the list through `Items()`.

## Decision

Add public package `github.com/alvadorncorp/gopackages/modeling/reducer` with
exactly:

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events immutable.List[E],
    reduce Reducer[S, E],
) S
```

Replay ranges over `events.Values()`, threading state left to right. The zero
list returns the initial state without invoking the reducer. Reducer purity and
determinism are documented application responsibilities, not runtime checks.

The dependency direction is:

```text
application → modeling/reducer → functional/immutable
application → modeling root → modeling/internal/decision → functional/immutable
application → modeling/change → functional/primitives/function
application → modeling/policy → functional/primitives/predicate + functional/values/optional
```

Neither reducer, change, policy, nor the root imports the other. Reducer imports
no Decision, Transition, Invariant, Rule, Requirement, Change, Policy,
reflection, registry, persistence, or transport capability.

Architecture tests permit a public functional type only in the approved
capability packages (policy for Predicate/Optional; reducer for the two exact,
compile-time-pinned reducer signatures), keep change's private function import
without leaking functional types, permit the exported name `Replay` only in
`modeling/reducer`, require that package in module enumeration, and cap its
exports to `Reducer` and `Replay`. Existing root and internal gates remain in
force everywhere else.

## Compatibility and rollback

The change is additive. Every shipped root, change, and policy contract remains
byte-for-byte and behaviorally unchanged; the existing `functional v1.1.0`
dependency and release topology do not change in this plan.
Before release or consumer adoption, rollback is a reverse-order revert of the
three Phase 1 commits. After release, removing the package or changing either
signature is a breaking public-contract decision requiring a separate release.

## Consequences

Applications gain a small, synchronous replay algebra with explicit laws and no
Event Sourcing infrastructure. They retain ownership of event unions and any
Decision-to-Replay conversion. The exception couples this one public signature
to `functional/immutable`, so its exact shape is a compatibility surface guarded
by published-dependency tests. No Apply helper, fallible reducer, slice overload,
marker Event, decision verifier, registry, event store, projection, or runtime is
authorized.
