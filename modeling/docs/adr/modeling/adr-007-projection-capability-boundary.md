---
type: adr
id: adr:modeling:007
context: modeling
title: Additive projection capability boundary
status: Aceita
tags: [modeling, projection, read-model, packages, compatibility, immutability]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:004, adr:modeling:005, adr:modeling:006, adr:modeling:010, tdd:modeling:projection]
date: 2026-08-10
---
# ADR-007: Additive projection capability boundary

> **Supersessão parcial (ADR-010):** o parâmetro `events` de `Project` passou de
> `immutable.List[E]` para `iter.Seq[E]`. A dependência de produção de
> `modeling/projection` é somente stdlib (`import "iter"`). A ponte com Decision
> é `Project(initial, decision.Events(), project)` sem materializar List. O
> restante desta ADR — boundary do pacote, tipo `Projection[P,E]`, exclusões de
> CQRS/runtime e independência de reducer — permanece em vigor. Ver
> [ADR-010](adr-010-streaming-replay-project-seq.md).

## Context and problem

Applications need to derive read models, summaries, indexes, and statistics
from facts that have already occurred. This is pure modeling behavior, not
persistence, subscription handling, message dispatch, or a CQRS runtime.

`Reducer[S,E]` and the requested `Projection[P,E]` have the same computational
shape, but not the same responsibility. Reducer answers how a fact changes
domain state. Projection answers how the fact changes a derived
representation. Coupling the capabilities would erase that distinction.

The settled Project contract accepts `immutable.List[E]`. Existing architecture
gates prohibit a public Projection name and functional types by default, so the
new package needs a narrow, mechanically checked exception without weakening
the root, change, policy, or reducer boundaries.

## Decision drivers

- expose an ordinary assignable `P x E -> P` function type;
- preserve event order and previous-result threading in batch projection;
- keep incremental and batch use on the same semantic operation;
- keep every event union and read-model type application-owned;
- avoid collection materialization and sibling-capability dependencies;
- exclude reflection, registries, effects, persistence, and CQRS runtime APIs;
- keep the public surface minimal until real consumers prove another operation.

## Options considered

### Reuse `reducer.Reducer` and `reducer.Replay`

Rejected. It would encode read-model derivation as domain-state reconstruction
and create an unnecessary dependency between semantic siblings.

### Delegate through `fold.FoldLeft(events.Items(), ...)`

Rejected. `FoldLeft` accepts a slice, while `immutable.List.Items()` allocates
and copies. Direct Values iteration preserves order without materialization.

### Add a lower iterator fold to functional

Rejected for this delivery. It expands another module's public surface to save
a trivial local loop and is not required by the Projection contract.

### Add composition or `ContramapEvent`

Rejected for now. No concrete consumer demonstrates repeated event adaptation,
and generic Projection composition has no unambiguous state transformation.

### Add one public Projection package with a narrow List exception (chosen)

Chosen because it preserves domain language, keeps capabilities orthogonal,
and delivers the complete minimum API without a runtime.

## Decision

Add public package `go.alvadorn.org/functional/modeling/projection` with
exactly:

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events immutable.List[E],
    project Projection[P, E],
) P
```

Project ranges over `events.Values()` exactly once and threads the projected
value left to right. Empty input returns `initial` without invoking `project`.
A non-empty list with a nil Projection retains ordinary Go nil-function panic
behavior. Purity and determinism are documented application responsibilities.

The dependency direction is:

```text
application -> modeling/projection -> functional/immutable
application -> modeling/reducer    -> functional/immutable
```

Projection and reducer do not import one another or any modeling sibling. The
root package re-exports neither Projection nor Project.

Architecture tests permit the immutable List signature only in the approved
capability, restrict exports to Projection and Project, pin an exact four-file
package envelope, require direct Values iteration with no Items/fold/reducer
dependency, and keep every existing reverse-edge and runtime prohibition.

## Compatibility and rollback

The change is additive and opt-in. Every existing import path, type identity,
signature, behavior, module requirement, and release group remains unchanged.
Before publication or adoption, rollback is a reverse-order revert of the plan
commits. After publication, removing the package or changing either signature
is breaking and requires a separate compatibility decision.

## Consequences

Applications gain a small synchronous algebra for derived representations and
may use the same Projection incrementally or across a history. They retain
ownership of event routing, unknown-event behavior, repositories, brokers,
checkpoints, and persistence.

This decision does not authorize FallibleProjection, ContramapEvent, Compose,
Zip, Map, ContramapState, marker interfaces, registries, handlers, reflection,
parallel projection, persistence, message consumption, or CQRS infrastructure.
