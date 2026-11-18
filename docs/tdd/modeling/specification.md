---
type: tdd
id: tdd:modeling:specification
context: modeling
title: Modeling Specification
status: Accepted
tags: [modeling, specification, classification, laws, testing]
related: [adr:modeling:specification-capability-boundary, adr:modeling:policy-composition-algebra, tdd:modeling:policy, tdd:modeling:rules-requirements]
date: 2026-08-11
---
# TDD: Modeling Specification

## Technical scope

Create the public leaf package `modeling/specification`, independent of the
root, of Policy, and of the other siblings. The package is a pure metamodel
with no effect, runtime, infrastructure, or concrete domain type.

## Public contracts

```go
type Specification[T any] func(T) bool
func All[T any](specifications ...Specification[T]) Specification[T]
func Any[T any](specifications ...Specification[T]) Specification[T]
func Not[T any](spec Specification[T]) Specification[T]
func Contramap[A, B any](spec Specification[B], mapper func(A) B) Specification[A]
```

This is the entire surface. `Of` and `Xor` were evaluated and rejected; Filter,
Validate, Explain, Reason, Error, Message, and bridges stay excluded.

## Semantics

Specification classifies a value with domain vocabulary. All and Any evaluate
left-to-right with short-circuit; `All()` returns true and `Any()` returns
false. Not negates. Contramap adapts an input before a single evaluation.

The zero Specification is a nil function and panics naturally when invoked.
Combinators reject nil operands eagerly with
`modeling/specification: nil specification`; Contramap rejects a nil mapper with
`modeling/specification: nil mapper`. Nil never means true, false, or skip.

## Dependencies and architecture

Production files import only the stdlib `slices` (`logic.go` only);
`transform.go` and `specification.go` have no imports. Specification imports
neither Predicate, the root, Policy, Change, Reducer, nor Projection. Archtests
pin the eighth package, the import set exactly `["slices"]`, five exports, the
absence of root aliases, and the prohibition of sibling edges.

## Tests and laws

Black-box tests pin nominal identity, ordinary-function assignment, zero/nil,
order, short-circuit, and Contramap. Observational laws cover double negation,
unary forms, empty identities, and true/false. Standard library fuzzing
validates the same properties over arbitrary inputs without comparing functions
directly or adding a dependency.

## Verification

Run focused tests, archtests, module tests, standalone module tests, race, vet,
tidy-diff, bounded fuzzing, the external example, workspace tests, and uncapped
lint.

## Compatibility and exclusions

The capability is additive. No existing API changes. There is no go.mod,
go.work, benchmark, migration, config, release, reflection, error, runner,
engine, registry, dispatcher, persistence, transport, or effect.
