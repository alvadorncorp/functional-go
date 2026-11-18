---
type: adr
id: adr:modeling:004
context: modeling
title: Public Change capability boundary
status: Aceita
tags: [modeling, change, state, composition, generics, package-boundary]
deciders: [alvadorn]
related: [adr:modeling:001, adr:modeling:002, adr:modeling:003, tdd:modeling:change]
date: 2026-08-10
---
# ADR-004: Public Change capability boundary

## Context and problem

Applications need to name, return, store, and compose pure state evolutions of
the form `S -> S`. Transition already models explicit input and output,
Invariant validates a result, Rule and Requirement express conditions, and
Decision carries state plus events. Reusing any of those names would erase the
meaning of a captured state change.

The accepted architecture exposed product primitives from the root package.
The Change specification instead requires a capability-oriented public package
and forbids root re-exports. That new import route is an additive public
contract and therefore needs an explicit boundary decision.

## Decision drivers

- preserve `Change[S]` as domain-semantic state evolution;
- keep ordinary `func(S) S` values directly usable;
- provide closed composition with a lawful empty identity;
- keep the root Decision facade and every existing root signature unchanged;
- consume only an independently resolvable functional release and prevent workspace-only dependency success;
- add no runtime, manager, validation, events, patch model, or deep-copy policy.

## Options considered

### Add Change to the root package

Rejected because the approved capability organization explicitly places the
surface under `modeling/change` and forbids new root symbols.

### Publicly alias functional Transform

Rejected because a public alias would leak mathematical implementation
vocabulary and couple future contracts.

### Duplicate Transform composition locally

Rejected after `functional/v1.1.0` was released and independently resolved.
Explicit private conversions can reuse Identity, ComposeTransforms, and
PipeTransforms without exposing their types or duplicating their loops.

### Create a second module or internal implementation package

Rejected because Change shares modeling's release, ownership, and dependency
policy, and its implementation has no private subsystem that justifies another
boundary.

### Create one public capability leaf (chosen)

Chosen because `modeling/change` gives the domain concept its own import and
documentation surface while remaining a tiny leaf over the already-required
functional module. Existing root consumers are unaffected.

## Decision

Create `go.alvadorn.org/functional/modeling/change` with exactly:

```go
type Change[S any] func(S) S
func Identity[S any]() Change[S]
func Compose[S any](changes ...Change[S]) Change[S]
func Pipe[S any](changes ...Change[S]) Change[S]
func Apply[S any](state S, transformation Change[S]) S
func ApplyAll[S any](state S, changes ...Change[S]) S
```

`Compose` evaluates right-to-left; `Pipe` and `ApplyAll` evaluate left-to-right.
Empty composition is Identity and empty ApplyAll returns its state. Identity,
not nil, represents no change. The package adds no eager nil validation or
custom panic; nil invocation retains ordinary Go behavior.

Change is conceptually immutable but only shallowly enforceable. Value state can
be copied before modification; slices, maps, pointers, and shared references
remain the Change author's responsibility. The package performs no reflection
or deep copy.

Implementation converts `Change[S]` values to and from
`functional/primitives/function.Transform[S]` privately. modeling upgrades its
existing functional requirement from v1.0.1 to v1.1.0 and permits direct imports
only from `functional/immutable` and `functional/primitives/function`.

The repository had no Reducer/Replay capability when this decision was written.
Reducer/Replay was introduced later, independently, by
[ADR-006](adr-006-reducer-capability-boundary.md); ADR-010 only amended it
partially, changing `Replay` events to `iter.Seq`
([ADR-010](adr-010-streaming-replay-project-seq.md)). Change remains
structurally independent of it, so this decision neither introduces nor claims
those concepts, and it authorizes no tag or release action.

## Consequences

Consumers gain one additive import route and six additive exports. The root
modeling package, internal Decision boundary, functional public API, module
topology, and existing modeling signatures remain unchanged. The existing
functional requirement advances to v1.1.0. Exact AST and file
gates cap the new package while generic architecture tests continue to reject
functional public types, forbidden imports, reverse dependencies, runtime APIs,
and mutable globals.

Before consumer adoption, rollback reverts the Change commits and dependency
bump. After a version containing the package is published, removal or incompatible behavior
requires a separately approved corrective or breaking release; publication is
never undone by deleting a tag from consumers.
