---
type: tdd
id: tdd:modeling:core
context: modeling
title: Modeling behavioral core
status: Aceito
tags: [modeling, decision, transition, invariant, testing]
related: [adr:modeling:001, adr:functional:003, adr:functional:006, adr:functional:007, adr:functional:010]
date: 2026-08-09
---
# TDD: Modeling behavioral core

## Technical scope

One root package provides domain-decision primitives. One internal archtest
package enforces dependency and runtime boundaries. There are no adapters,
registries, subpackages, persistence formats, or infrastructure components.

## Public contracts

```go
type Decision[S, E any] struct { /* private */ }
func Decide[E, S any](state S) Decision[S, E]
func (Decision[S, E]) State() S
func (Decision[S, E]) Emit(events ...E) Decision[S, E]
func (Decision[S, E]) EventCount() int
func (Decision[S, E]) Events() iter.Seq[E]

type Transition[S, I, O any] func(S, I) O
func (Transition[S, I, O]) Apply(S, I) O
type FallibleTransition[S, I, O any] func(S, I) (O, error)
func (FallibleTransition[S, I, O]) Apply(S, I) (O, error)

type Invariant[S any] func(S) error
func FailFast[S any](...Invariant[S]) Invariant[S]
func Accumulate[S any](...Invariant[S]) Invariant[S]
```

Decision stores events in a private `immutable.List[E]`; the zero value is a
valid zero-state/no-event decision. Emit preserves its receiver and argument
order. Events returns a typed iterator. State and event immutability are shallow.

Fallible transitions use ordinary errors. FailFast preserves the first error
identity and stops. Accumulate evaluates all and returns `errors.Join`. Both
composition functions accept zero invariants and reject nil with the exact panic
`modeling: nil invariant`.

## Tests and cost evidence

- external tests cover zero/empty decisions, event typing/order, immutability,
  ordinary function adaptation, error identity, fail-fast, and accumulation;
- executable examples own Order, Money, IDs, commands, and event sum types;
- a `!race` test requires zero allocations only for trivial construction/state/count;
- benchmarks report Empty, EmitOne, EmitBatch, EmitChain, and Events with
  `-benchmem`, without predeclaring arbitrary Emit caps.

## Architecture gates

Production code may directly import only `functional/immutable` from functional,
and no exported signature may contain a functional type;
it cannot directly import reflect, unsafe, any other functional package,
functional/internal, database/sql, net/http, os, or io/fs. Its non-standard
closure is limited to modeling and functional because functional may use its own
implementation packages transitively. functional cannot depend on modeling. Exported runtime
magic names and package-level mutable variables are rejected by AST tests.

## Verification and release

Run module test/race/vet/tidy, root test/lint/race/bench/check, exact release-TOML
validation, and GOWORK=off tidy/list/test against functional v1.0.1. Publication
uses an independent module-scoped tag; publication itself is outside this plan.
