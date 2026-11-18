---
type: adr
id: adr:modeling:specification-capability-boundary
context: modeling
title: Nominal and independent Specification capability
status: Accepted
tags: [modeling, specification, classification, packages, architecture]
deciders: [alvadorn]
related: [adr:modeling:minimal-type-safe-domain-modeling, adr:modeling:policy-capability-boundary, adr:modeling:policy-composition-algebra, tdd:modeling:specification]
date: 2026-08-11
---
# Nominal and independent Specification capability

> **Wording correction:** the original decision described the package as
> "zero-import". The delivered implementation imports only the stdlib `slices`
> (`logic.go`); the decision — no `functional`, root, `policy`, or siblings —
> remains in force, and the text below has been aligned with that fact.

## Context and problem

The module distinguishes generic conditions, domain classification, explainable
validation, operational preconditions, invariants, and strategy selection.
`functional.Predicate[T]` and `Specification[T]` share the `func(T) bool`
representation but not the semantic responsibility. The new capability must
name domain classification without widening the root or creating an engine.

## Decision drivers

- preserve `application -> modeling -> functional` and forbid the reverse direction;
- keep sibling capabilities independent and opt-in;
- communicate intent through its own nominal identity;
- offer only algebra proven by domain examples;
- keep nil, order, and short-circuit behavior predictable before the API freezes.

## Alternatives considered

### Re-export Specification from the root

Rejected: it would widen the stable facade and erase the capability boundary.

### Public alias of Predicate

Rejected: an alias would lose nominal identity and confuse generic algebra with
domain vocabulary.

### Adapt Predicate's algebra internally

Rejected: `All` and `Any` would require slice conversions, eager nil checks
would remain local, and the direct loops are smaller. The coupling buys neither
semantics nor meaningful implementation.

### Direct leaf package, nominal, stdlib `slices` only (chosen)

Chosen: `modeling/specification` exposes exactly `Specification`, `All`, `Any`,
`Not`, and `Contramap` and implements the algebra over plain Go functions.

## Decision

`Specification[T]` is a nominal `func(T) bool` that answers whether a value
belongs to a logical domain category. It does not validate, does not return
error, does not explain failure, does not authorize an operation, and does not
select strategy.

`All` and `Any` evaluate left to right and short-circuit. `All()` is the true
identity; `Any()` is the false identity. `Not` negates and `Contramap` adapts
the input. Combinators reject a nil Specification or mapper at construction
with deterministic panics. Invoking the zero Specification keeps Go's natural
nil-function behavior.

The package imports no `functional`, root `modeling`, `policy`, or siblings; its
only production import is the stdlib `slices` (`logic.go`), and `transform.go`
and `specification.go` have no imports.
`Of` is excluded because direct assignment/conversion suffices. `Xor` is
excluded because no domain case justified it. Filter, error, message, reason,
validate, registry, reflection, runtime, and bridges are also excluded.

## Compatibility and rollback

The change is additive. Before publication, rollback reverts the capability
commits. After publication, removing the package or changing any of the five
contracts, empty identities, order, short-circuit, or nil semantics is breaking
and requires a new decision and SemVer handling. This work publishes no release.

## Consequences

Applications gain nominal, composable classification while keeping Policy,
Requirement, Rule, and Invariant semantically distinct. The cost is a small,
intentional duplication of generic boolean algebra in exchange for a stronger
dependency boundary (a leaf with only the stdlib `slices`) and its own nil
contracts.
