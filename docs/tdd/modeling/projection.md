---
type: tdd
id: tdd:modeling:projection
context: modeling
title: Modeling Projection and Project
status: Accepted
tags: [modeling, projection, read-model, iter, generics, laws, testing]
related: [adr:modeling:reducer-capability-boundary, adr:modeling:projection-capability-boundary, adr:modeling:streaming-replay-project-seq, tdd:modeling:reducer-replay]
date: 2026-08-12
---
# TDD: Modeling Projection and Project

## Technical scope

Describes the importable package `modeling/projection`. Existing contracts of
the root, change, policy, reducer, internal, module, dependencies, and release
stay separate; projection does not change them.

Projection derives an application-owned representation from facts. Reducer
evolves domain state from facts. The computational shape is the same; the
semantic responsibility is not. Neither package imports the other.

The `Project` events parameter is `iter.Seq[E]` (`adr:modeling:streaming-replay-project-seq`). The earlier
`immutable.List[E]` choice in `adr:modeling:projection-capability-boundary` is
superseded for this parameter.

## Public contracts

The package exports exactly:

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Ordinary `func(P, E) P` functions are assignable and invocable directly. P and
E are unrestricted `any`; the library defines no Event, ReadModel, View, or
marker. Value semantics: no pointer in the capability type and no
`Ref`/`Box`/`Readonly`.

## Semantics and laws

Projection is the incremental operation:

```text
P x E -> P
```

Project is the ordered, repeated application (a single-pass fold):

```text
p1 = project(p0, e1)
p2 = project(p1, e2)
...
pn = project(p(n-1), en)
```

Project is synchronous, invokes the Projection once per event, preserves the
sequence's encounter order, and does not parallelize. Empty input returns the
initial value without invocation. Non-empty input with a nil Projection panics
with ordinary Go nil-function behavior.

Concatenation law (where the application defines `A ++ B` observably):

```text
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

Batch and incremental use are equivalent because Project only repeats the same
operation. Repeated calls return equal results when the Projection is
deterministic. The package inspects no clocks, randomness, I/O, global state,
mutation, or purity.

Returning the current P means the event does not affect the representation.
The unknown-event policy belongs to the application's union/switch.

Because the sequence is single-pass, multi-pass requires explicit
materialization before Project.

## Decision bridge

When a Decision's events feed a read model:

```go
view := projection.Project(empty, decisionValue.Events(), projectFn)
```

`immutable.CollectList` is unnecessary. The library offers no `ProjectDecision`
and does not compare Decision state with the projected value.

## Implementation boundary and dependencies

Project iterates `for event := range events` and threads a local projected
value. It calls no `Items()`, imports no fold, materializes no slice, and
delegates to no reducer. Complexity is `O(n * cost(projection))` with no
automatic event copying and no reflection dispatch.

Production imports: only the standard library (`iter`). Architecture tests must
enforce exactly two exports, no sibling dependency and no `functional/*`, a
direct range over `events`, and no reexport in the root.

## Tests and benchmarks

Black-box tests pin: ordinary-function assignment; struct and primitive P; P
with an application-owned persistent collection; non-comparable E; identity on
an irrelevant event; zero/one/many events; encounter order; chaining of the
previous result; equivalence with manual sequential application; concatenation;
deterministic repetition; shallow immutability of the initial value;
batch/incremental equivalence; single pass (no implicit second sweep); and a
direct bridge from a Decision sequence where applicable.

Benchmarks report `ns/op`, `B/op`, and `allocs/op` at sizes 0, 1, 8, 32,
128, 1024, and 8192 for int projection, index projection, and a direct-loop
baseline over the sequence. Results are evidence, not a machine-independent
threshold. Comparing against `FoldLeft` over a materialized slice would measure
avoidable copying and stays out.

## Projection versus infrastructure

Projection returns a value. The application loads facts and current views,
invokes Projection or Project, and persists the result in its own layer. This
package knows no repository, database, broker, event store, subscription,
offset, checkpoint, consumer group, background worker, command/query handler,
or eventual-consistency runtime.

## Excluded surface

There is no FallibleProjection, ContramapEvent, Compose, Zip, Map,
ContramapState, Replay alias, parallel API, ReadModel/View/Event marker,
registry, dispatcher, handler map, reflection, persistence, transport,
serialization, version migration, non-stdlib dependency change, or root alias.

## Verification

Run focused projection and architecture tests (single pass and stdlib only),
module tests in the workspace and with `GOWORK=off`, race, vet, tidy-diff,
lint without hidden caps, package listing, the benchmark matrix, and the
implementation plan's repository-wide gates when the delivery includes code.
