---
type: adr
id: adr:modeling:001
context: modeling
title: Minimal explicit and type-safe domain metamodel
status: Aceita
tags: [modeling, generics, immutability, domain, architecture]
deciders: [alvadorn]
related: [adr:functional:003, adr:functional:004, adr:functional:005, adr:functional:006, adr:functional:007, adr:functional:010, tdd:modeling:v0-1-core]
date: 2026-08-09
---
# ADR-001: Minimal explicit and type-safe domain metamodel

## Context and problem

Applications need reusable compile-time primitives for domain decisions without
turning Go into a reflective enterprise runtime. `functional` already owns
technical values, immutable collections, and state evolution; modeling must own
only semantic governance above those capabilities.

## Decision drivers

- ordinary functions and generics must remain first-class;
- decisions preserve typed events, order, empty semantics, and immutability;
- functional internals and sibling-module types must not leak into the public contract;
- application domains own concrete values, identities, entities, aggregates,
  commands, events, rules, and policies;
- reflection, registries, dispatchers, buses, storage, and infrastructure are
  mechanically excluded.

## Options considered

### Universal structural wrappers and marker interfaces

Rejected: they add nominal ceremony without enforcing domain validity and make
application models inherit a library shape.

### Public functional collection/state types

Rejected: even stable sibling types would couple modeling's compatibility
surface to functional's collection/state API. Public signatures stay
modeling-owned or standard-library-only.

### Minimal behavioral core with documented structural semantics (chosen)

Chosen: export Decision, Transition, FallibleTransition, Invariant, FailFast,
and Accumulate. Keep immutable.List private behind iter.Seq. Demonstrate the
rest of the vocabulary with application-owned plain Go types and functions.

## Decision

The dependency direction is application domain → modeling → public functional.
functional never imports modeling. modeling imports no functional/internal path.
No reflection, runtime discovery, global registry, magic dispatch, infrastructure,
or concrete business model is part of the module.

The exact public v0.1 contract is recorded in `tdd:modeling:v0-1-core`. Its
signature order, zero values, event order, shallow immutability, error identity,
and invariant evaluation behavior are compatibility surface.

## Consequences

Applications retain idiomatic Go models and may use functional collections/state
when their size justifies it. New framework types such as Rule or Policy require
demonstrated compile-time behavior in a future separately approved change.
