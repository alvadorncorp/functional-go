# Architecture overview

`modeling` provides a metamodel, not an application domain:

```text
application domain
      ↓ specializes
modeling
      ↓ uses public capabilities
functional
```

## Structural semantics

Value, Identity, Entity, Aggregate, and Model remain application-owned Go
types. The module does not export marker interfaces, required embeddings, base
types, or reflection-based equality. An aggregate is a consistency boundary,
not a database row or repository record.

## Behavioral primitives

The v0.1 contract is deliberately limited to:

- `Decision[S,E]` carries the next state and ordered typed domain events;
- `Transition[S,I,O]` and `FallibleTransition[S,I,O]` are first-class domain
  functions;
- `Invariant[S]` composes state validity checks with fail-fast or accumulating
  error semantics.

Rule and Policy remain ordinary application functions until a concrete generic
behavior justifies another public type. Command means intent and Event means an
occurred fact; neither requires a marker interface or runtime bus.

## Decision pipeline

```text
Current State + Command/Intent
              ↓
Transition / FallibleTransition
              ↓
application Rules and Policies
              ↓
Next State → Invariants
              ↓ valid
Decision[State, Event]
```

Commands remain intentions. Events remain facts. A Decision knows only its next
state and typed events; it has no topic, offset, outbox, store, transport, or
serialization concern. Fallible transitions return ordinary Go errors and do
not require a modeling-owned Result hierarchy.

`FailFast` returns the first invariant error unchanged. `Accumulate` evaluates
every invariant and uses `errors.Join`, preserving `errors.Is` and `errors.As`.
Nil invariants panic while the composition is built, rather than becoming a
hidden runtime rule.

## Immutability and cost

Decision uses a private immutable list so `Emit` returns a new value and keeps
typed order. Although functional v1 is stable, the public API exposes
`iter.Seq[E]` so modeling remains independently evolvable.
Immutability is shallow: referenced state or event values remain the caller's
responsibility. Repeated single-event emits may incur persistent path-copy
intermediates; benchmarks therefore report empty, single, batch, chain, and
iteration costs. Only trivial construction/state/count carries a zero-allocation assertion.

## Non-goals

The module has no reflection, annotations, scanner, registry, service locator,
dependency-injection container, magic dispatch, command bus, event bus, event
store, replay, projection, repository, ORM, database, HTTP, broker, filesystem,
or concrete application-domain type.

`functional/state` and `functional/immutable` are optional capabilities for
applications with larger state. They are not superclasses, and their internal
engines never belong to modeling.
