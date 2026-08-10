---
type: tdd
id: tdd:modeling:v0-2-decision-composition
context: modeling
title: Modeling v0.2 Decision composition
status: Aceito
tags: [modeling, decision, composition, generics, immutability, testing]
related: [adr:modeling:001, adr:modeling:002, tdd:modeling:v0-1-core]
date: 2026-08-10
---
# TDD: Modeling v0.2 Decision composition

## Technical scope

The root `modeling` package remains the only externally importable product
package. `modeling/internal/decision` owns the zero-valid persistent value and
composition mechanics. `modeling/internal/archtest` enforces dependencies and
public-surface constraints. Transition and Invariant stay unchanged.

## Public contracts

The v0.1 Decision type, constructor, and methods retain their exact signatures:

```go
type Decision[S, E any] struct { /* private */ }
func Decide[E, S any](state S) Decision[S, E]
func (Decision[S, E]) State() S
func (Decision[S, E]) Emit(events ...E) Decision[S, E]
func (Decision[S, E]) EventCount() int
func (Decision[S, E]) Events() iter.Seq[E]
```

v0.2 adds exactly:

```go
func MapState[S, T, E any](Decision[S, E], func(S) T) Decision[T, E]
func MapEvents[S, E, F any](Decision[S, E], func(E) F) Decision[S, F]
func Emit[S, E any](Decision[S, E], E) Decision[S, E]
func AppendEvents[S, E any](Decision[S, E], ...E) Decision[S, E]
func Combine[S, E any](Decision[S, E], Decision[S, E]) Decision[S, E]
```

There is no FlatMap, alias API, new method, exported representation, or public
subpackage.

## Semantics and laws

- MapState applies its mapper once to state and preserves ordered events.
- MapEvents preserves state and maps every event once in source order.
- Emit appends one event; AppendEvents appends zero or more in argument order.
- Combine returns the right Decision state and left events followed by right
  events.
- Every operation preserves its input snapshots.
- MapState and MapEvents satisfy identity and composition by State/Events
  observation.
- Combine is associative by State/Events observation; empty-side cases still
  follow right-state semantics.

Decision is intentionally not required to be Go-comparable. Tests assert only
its public observations.

## Internal design

```text
modeling facade
      ↓
modeling/internal/decision.Value[S,E]
      ↓
functional/immutable.List[E]
```

MapState creates a new typed Value while reusing the persistent event list.
MapEvents builds typed output storage in order with `ListBuilder`. Append
delegates to `List.Append`. Combine seeds a Builder from the left list and
appends the right iterator; empty sides reuse safe persistent storage. No path
uses a slice round-trip.

The root public-signature architecture gate skips Go internal packages only;
all dependency, direct-import, runtime-name, global-state, and reverse-edge
gates continue to enumerate every production package.

## Cost and evidence

The trivial Decision construction/state/count path retains its zero-allocation
test outside race mode. MapState performs no event traversal. MapEvents and the
non-empty Combine path materialize output through a Builder; Append follows the
documented persistent List behavior. Benchmarks report, rather than promise,
ns/op, B/op, and allocs/op for Emit, AppendEvents, Combine, and MapEvents at
0, 1, 8, 32, 128, and 1024 events.

## Verification

Run module test, race, vet, tidy-diff, lint with unlimited issue caps, exact
package listing, and the composition benchmark matrix. Then run workspace
`make test-unit`, `make test-race`, `make lint`, `make vet`, and `make bench`,
followed by `git diff --check`. No integration database target or release step
is part of this contract.
