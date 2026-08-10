# Architecture overview

`modeling` provides a metamodel, not an application domain:

```text
application domain
      ↓ specializes
modeling root facade
      ↓ delegates privately
modeling/internal/decision
      ↓ uses public capabilities
functional/immutable
```

Applications import only the root package. The internal Decision package is an
enforceable implementation boundary, not another public API.

## Structural semantics

Value, Identity, Entity, Aggregate, and Model remain application-owned Go
types. The module does not export marker interfaces, required embeddings, base
types, or reflection-based equality. An aggregate is a consistency boundary,
not a database row or repository record.

## Behavioral primitives

The public contract contains:

- `Decision[S,E]` with its valid zero value, constructor, and observation methods;
- `MapState`, `MapEvents`, `Emit`, `AppendEvents`, and `Combine` for typed
  Decision composition;
- `Transition[S,I,O]` and `FallibleTransition[S,I,O]` as first-class domain
  functions;
- `Invariant[S]` with fail-fast or accumulating error composition.

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
              ↓
typed map / emit / append / combine
```

Commands remain intentions. Events remain facts. A Decision knows only its next
state and typed events; it has no topic, offset, outbox, store, transport, or
serialization concern. Fallible transitions return ordinary Go errors and do
not require a modeling-owned Result hierarchy.

`MapState` changes only state; `MapEvents` changes only events in order. `Emit`
and `AppendEvents` preserve the input snapshot. `Combine` selects the right
state and concatenates left-before-right events. The mapping operations obey
identity and composition observationally, and Combine is observationally
associative. FlatMap and alias spellings are not part of v0.2.

`FailFast` returns the first invariant error unchanged. `Accumulate` evaluates
every invariant and uses `errors.Join`, preserving `errors.Is` and `errors.As`.
Nil invariants panic while the composition is built, rather than becoming a
hidden runtime rule.

## Immutability and cost

The public root Decision is a facade over a private internal Value backed by an
immutable list. Events remain exposed only as `iter.Seq[E]`, so modeling stays
independently evolvable. Immutability is shallow: referenced state or event
values remain the caller's responsibility.

MapState reuses event storage. MapEvents and non-empty Combine use a persistent
ListBuilder; Emit and AppendEvents delegate to persistent List append. The
benchmark suite reports allocation and latency observations at 0, 1, 8, 32,
128, and 1024 events. It does not define machine-independent performance caps.

## Non-goals

The module has no reflection, annotations, scanner, registry, service locator,
dependency-injection container, magic dispatch, command bus, event bus, event
store, replay, projection, repository, ORM, database, HTTP, broker, filesystem,
or concrete application-domain type.

`functional/state` and `functional/immutable` are optional capabilities for
applications with larger state. They are not superclasses, and their internal
engines never belong to modeling.
