---
type: tdd
id: tdd:modeling:projection
context: modeling
title: Modeling Projection and Project
status: Aceito
tags: [modeling, projection, read-model, immutable, generics, laws, testing]
related: [adr:modeling:006, adr:modeling:007, tdd:modeling:reducer-replay]
date: 2026-08-10
---
# TDD: Modeling Projection and Project

## Technical scope

Create one externally importable `modeling/projection` package. Existing root,
change, policy, reducer, internal, module, dependency, and release contracts
remain unchanged.

Projection derives an application-owned representation from facts. Reducer evolves domain state from facts. Their computational shape is equal, but their semantic ownership is not; neither package imports the other.

## Public contracts

The package exports exactly:

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events immutable.List[E],
    project Projection[P, E],
) P
```

Ordinary `func(P,E) P` values are directly assignable and callable. P and E are
unconstrained `any`; the library defines no Event, ReadModel, View, or marker
interface.

## Semantics and laws

Projection is the incremental operation:

```text
P x E -> P
```

Project is repeated incremental application:

```text
p1 = project(p0, e1)
p2 = project(p1, e2)
...
pn = project(p(n-1), en)
```

Project is synchronous, invokes the Projection once per event, preserves
immutable-list encounter order, and never parallelizes. Empty input returns the
initial value without invocation. A non-empty input with nil Projection has
ordinary Go nil-function panic behavior.

The concatenation law is:

```text
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

Batch and incremental use are equivalent because Project only repeats the same
Projection operation. Repeated calls return equal results when the supplied
Projection is deterministic. The package does not inspect clocks, randomness,
I/O, global state, mutation, or purity.

Returning the current P means an event does not affect the representation.
Unknown-event policy belongs to the application's event union and switch.

## Implementation and dependency boundary

Project iterates `events.Values()` and threads a local projected value. It does
not call `Items()`, import a fold package, materialize a slice, or delegate to
reducer. Its complexity is `O(n * cost(projection))` with no automatic event
copying or reflection dispatch.

Production projection imports only `functional/immutable`. Architecture tests
enforce exactly two exports, exactly two production and two test files, no
modeling sibling dependency or root re-export, exact seven-package topology,
and the direct Values implementation.

## Tests and benchmarks

Black-box tests pin ordinary-function assignment, struct and primitive P,
persistent `immutable.Map` P, non-comparable E, irrelevant-event identity,
empty/one/many events, encounter order, previous-result threading, manual
sequential equivalence, concatenation, deterministic repetition, unchanged
events, unchanged immutable initial value, and batch/incremental equivalence.

Benchmarks report `ns/op`, `B/op`, and `allocs/op` at sizes 0, 1, 8, 32, 128,
1024, and 8192 for an int projection, an immutable.Map projection, and a direct
Values-loop baseline. Results are evidence, not machine-independent pass/fail
thresholds. Comparing to `FoldLeft(events.Items())` would measure an avoidable
copy and is deliberately excluded.

## Projection versus infrastructure

Projection returns a value. The application loads events and current views,
invokes Projection or Project, and persists the result in its own layer. This
package knows no repository, database, broker, event store, subscription,
offset, checkpoint, consumer group, background worker, command/query handler,
or eventual-consistency runtime.

## Excluded surface

There is no FallibleProjection, ContramapEvent, Compose, Zip, Map,
ContramapState, Replay alias, parallel API, ReadModel/View/Event marker,
registry, dispatcher, handler map, reflection, persistence, transport,
serialization, version migration, dependency change, or root alias.

## Verification

Run focused projection and architecture tests, module tests in workspace and
GOWORK=off modes, race, vet, tidy-diff, uncapped lint, exact package listing,
the benchmark matrix, and the repository-wide test/race/lint/vet gates from the
implementation plan.
