---
type: adr
id: adr:modeling:policy-capability-boundary
context: modeling
title: Additive public Policy capability
status: Accepted
tags: [modeling, policy, generics, optional, predicate, architecture]
deciders: [alvadorn]
related: [adr:modeling:minimal-type-safe-domain-modeling, adr:modeling:root-facade-internal-decision, adr:modeling:rules-and-requirements, adr:modeling:policy-composition-algebra, tdd:modeling:policy]
date: 2026-08-10
---
# Additive public Policy capability

> **Partial supersession
> ([policy composition algebra](policy-composition-algebra.md)):** the surface of
> `modeling/policy` grew from seven to eleven exports (`Compose`, `Zip`,
> `Choose`, `OrElse`). The seven original symbols, their signatures, evaluation
> order, and panics remain in force, and this ADR's status is still Accepted.

## Context and problem

The shipped modeling baseline exposes primitives in the root package. A Policy
answers which result or strategy derives from a context; it does not validate,
does not evolve state, and does not run effects. The request called for a
dedicated `modeling/policy` capability, no root aliases, and conditional
selection via Predicate and Optional. This deliberately contradicts the
root-only limit and the general prohibition of functional types in public APIs.

The capability slot had been reserved, at the time of this decision, for a
planned Reducer/Replay predecessor; in the effective chronology that slot holds
Change ([the change capability boundary](change-capability-boundary.md)) and
Reducer/Replay arrives separately ([the reducer capability
boundary](reducer-capability-boundary.md)). Policy took the slot that follows
without rewriting or anticipating Change or Reducer/Replay.

## Decision drivers

- keep pure functions, first-class types, and a single generic input;
- give the capability its own value through transformation and conditional selection;
- represent inapplicability as Optional, never as error;
- preserve all existing APIs and forbid runtime/dynamic dispatch;
- limit the functional exception to the two types actually needed.

## Options considered

### Re-export Policy from the root

Rejected: it widens the root, erases the capability as a unit, and contradicts
the request.

### Create a nominal Conditional type

Rejected for now: plain `I -> Optional[O]` functions suffice and avoid
crystallizing a second abstraction with no demonstrated independent use.

### Use error for inapplicability

Rejected: it mixes absence of selection with domain failure.

### New package with a narrow functional exception (chosen)

Chosen: `modeling/policy` exposes exactly Policy, Constant, Map, Contramap,
When, FirstApplicable, and Fallback. Only their conditional signatures may use
`predicate.Predicate` and `optional.Optional`.

## Decision

The dependency is application → modeling/policy → the public packages
`functional/primitives/predicate` and `functional/values/optional`, plus the
stdlib `slices`, used in FirstApplicable's defensive copy. The root `modeling`
package, siblings, and `functional` do not depend on policy.

FirstApplicable requires a first operand, evaluates in declaration order, and
stops at the first Some. Fallback evaluates its total policy only on None.
Combinators reject nil operands at construction with deterministic private
panics. Apply, a nominal Conditional, multi-arity, methods, aliases, and
runtime are excluded.

## Compatibility and rollback

The change is additive and opt-in. No existing name, signature, behavior, or
import path changes; `functional v1.1.0` preserves the required contracts and is
the pinned release.
Before adoption, rollback reverts the phase commits. After publication and
adoption, removing or changing any of the seven signatures is breaking and
requires a new ADR and release; this work publishes no version.

## Consequences

Consumers gain policy composition without an engine. The gates become
package-aware: they keep the general prohibition and allow Predicate/Optional
only in policy. The pending Reducer/Replay plan shares these gates and cannot
run in parallel; whichever plan runs second must be deepened.
