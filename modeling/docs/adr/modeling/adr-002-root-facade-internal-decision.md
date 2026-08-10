---
type: adr
id: adr:modeling:002
context: modeling
title: Root facade over an internal Decision package
status: Aceita
tags: [modeling, decision, packages, compatibility, immutability]
deciders: [alvadorn]
related: [adr:modeling:001, tdd:modeling:decision-composition]
date: 2026-08-10
---
# ADR-002: Root facade over an internal Decision package

## Context and problem

Decision composition needs enough implementation structure to keep persistent
collection mechanics separate from the public metamodel. The existing public
identity, generic order, constructor arity, method set, and valid zero value are
already compatibility surface. A root-only implementation would leave the
module flat, while a public subpackage or alias-backed canonical type would add
an import route and change the conceptual or observable type origin.

## Decision drivers

- preserve `modeling.Decision[S,E]` as the canonical defined public type;
- preserve `Decide[E,S](state)` and every existing Decision method signature;
- make the implementation boundary enforceable by Go import rules;
- keep persistent storage and algorithms reusable by the five root functions;
- expose no functional collection type and add no runtime or infrastructure.

## Options considered

### Keep all Decision files in the root package

Rejected because files alone do not create an enforceable boundary and leave
the module flat.

### Put the canonical type in a public `modeling/decision` package

Rejected because it adds a public import route and either breaks the root type
identity or requires aliases whose canonical origin is no longer the root.

### Alias an internal canonical type from root

Rejected because source compatibility would hide, rather than avoid, an
internal canonical identity and would couple public metadata to that choice.

### Keep the root defined type as a facade over `internal/decision` (chosen)

Chosen because the public identity stays in `modeling`, while Go prevents
external consumers from importing the internal mechanics.

## Decision

`modeling.Decision[S,E]` remains a defined root type with one private
`internal/decision.Value[S,E]` field. The internal value owns state and the
private `immutable.List[E]`, is zero-valid, and implements append, map, and
combine mechanics. Root constructors, methods, and composition functions are
thin facades.

The dependency direction is:

```text
application → modeling → modeling/internal/decision → functional/immutable
```

The reverse edge is forbidden. Architecture tests continue to inspect every
production package for dependency, runtime, and global-state constraints, but
the functional-type public-signature gate examines only externally importable
packages.

The only Decision-composition public additions are `MapState`, `MapEvents`,
`Emit`, `AppendEvents`, and `Combine`. Transition and Invariant remain in the
root; moving them is a separate refactor.

## Consequences

Consumers keep one public import path and their existing Decision source
contracts. Internal algorithms may use immutable collections without leaking
them into external signatures. The root facade and internal value require
explicit compatibility and dependency-direction tests. Another internal
package is justified only by a separately demonstrated ownership boundary;
this decision does not authorize framework layers or public aliases.
