---
type: adr
id: adr:modeling:policy-composition-algebra
context: modeling
title: Closed Policy composition algebra
status: Accepted
tags: [modeling, policy, composition, optional, compatibility]
deciders: [alvadorn]
related: [adr:modeling:policy-capability-boundary, adr:modeling:specification-capability-boundary, tdd:modeling:policy]
date: 2026-08-11
---
# Closed Policy composition algebra

## Context and problem

[The policy capability boundary](policy-capability-boundary.md) established
Policy as a pure `I -> O` strategy with seven exports, inapplicability via
Optional, and eager panics for nil operands. Sequential composition,
combination of independent decisions, total choice, and binary conditional
fallback were still missing. The stable line needs to complete that vocabulary
without turning Policy into a rules engine or a monad.

## Decision drivers

- preserve the seven existing contracts and two dependencies;
- make order and short-circuit an observable part of the contract;
- keep totality separate from inapplicability;
- avoid public tuples, FlatMap, runtime, bridges, or a new nominal abstraction;
- freeze the smallest semantically complete surface.

## Alternatives considered

### FlatMap and monadic algebra

Rejected: it mixes the original input with subsequent decisions and grows the
abstraction without a single domain meaning.

### Public tuple for Zip

Rejected: it would leak `functional` structure; an explicit combiner produces
the domain type the application actually wants.

### Nominal Conditional

Rejected again: `func(I) optional.Optional[O]` remains sufficient and
compatible with When, FirstApplicable, and Fallback.

### Four closed combinators (chosen)

Chosen: add `Compose`, `Zip`, `Choose`, and `OrElse`.

## Decision

`Compose(first, second)` evaluates `first` and feeds its output into `second`.
`Zip(left, right, combine)` evaluates strictly left, then right, then combine.
`Choose(predicate, whenTrue, whenFalse)` evaluates the predicate once and only
the selected branch. `OrElse(left, right)` is the binary composition of
conditional policies; `FirstApplicable` remains the N-ary form.

Choose uses `predicate.Predicate[I]`, same as When. The application adapts a
Specification via a closure or explicit conversion, making the semantic
boundary visible without a sibling dependency. OrElse preserves Optional and
short-circuits left to right.

Compose/Zip/Choose reuse the existing panic for a nil Policy; Choose reuses the
nil-Predicate panic; OrElse reuses the nil-Conditional panic. Zip adds
`modeling/policy: nil combiner`. No new dependency is required and the seven
previous symbols stay unchanged.

## Compatibility and rollback

The change is additive and widens the surface from seven to eleven symbols.
Before publication, each slice can be reverted. After publication, removing any
of the eleven symbols or changing its signature, order, short-circuit,
zero-output, or nil semantics is breaking and requires a new decision and
SemVer. No publication in this work.

## Consequences

Policy is feature-complete for stabilization: Map transforms output, Contramap
adapts input, Compose chains, Zip combines, Choose selects a total strategy,
When introduces inapplicability, OrElse tries an alternative, FirstApplicable
searches among N, and Fallback makes it total. New abstractions require a domain
problem that this vocabulary cannot express.
