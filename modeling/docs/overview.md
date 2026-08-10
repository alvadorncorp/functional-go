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

Applications import the root package for Decision, Rule, Requirement,
Transition, and Invariant. Context-derived strategies live in the distinct
`modeling/policy` capability; the root does not reexport Policy symbols. The
internal Decision package remains an enforceable implementation boundary.

## Structural semantics

Value, Identity, Entity, Aggregate, and Model remain application-owned Go
types. The module does not export marker interfaces, required embeddings, base
types, or reflection-based equality. An aggregate is a consistency boundary,
not a database row or repository record.

## Behavioral primitives

The public contract contains:

- `Decision[S,E]` with its valid zero value, constructor, observations, and
  typed composition functions;
- `Rule[T]` for an explanatory condition over one value, composed by ordered
  `AllRules` or non-empty `AnyRule`;
- `Requirement[S,I]` for an operation precondition over state plus input,
  composed by ordered `AllRequirements` or non-empty `AnyRequirement`;
- `Transition[S,I,O]` and `FallibleTransition[S,I,O]` for next-state computation;
- `Invariant[S]` for resulting-state validity, with fail-fast or accumulating
  error composition.

Rule and Invariant deliberately remain distinct nominal types even though both
return error. Requirement answers whether an input may be applied to current
state; it is not a Transition hook. Ordinary functions remain assignable and
all returned errors remain application-owned.

## Policy capability

`modeling/policy` represents `I → O`: a pure strategy that derives a domain
result from one explicit context. `Constant`, `Map`, and `Contramap` reuse and
adapt policies. `When` turns a Policy into `I → Optional[O]`,
`FirstApplicable` selects the first Some in declaration order, and `Fallback`
restores a total Policy.

Policy is not Rule (`T → error`), Requirement (`S × I → error`), Transition
(`S × I → O`), Change (`S → S`), or Reducer (`S × E → S`). It executes no
effects and imports no sibling modeling capability. Applicability is Optional,
never error. Multiple inputs belong in an application-owned context struct.

## Explicit decision flow

```text
Current State + Input
          ↓
Requirement[S,I] ── failure → application error
          ↓ valid
Transition[S,I,O]
          ↓
Invariant[O] ────── failure → application error
          ↓ valid
Decision[O,E]
```

The application explicitly invokes each step. modeling has no pipeline,
runner, command handler, rules engine, registry, dispatcher, persistence, or
transport responsibility.

AllRules and AllRequirements preserve declaration order, stop at the first
failure, preserve its identity, and accept zero operands. AnyRule and
AnyRequirement require a first operand, stop at the first success, and return
the final evaluated error when all operands fail. Nil constituents panic while
composition is built, matching the existing Invariant programmer-error policy.

`MapState` changes only Decision state; `MapEvents` changes only events in
order. `Emit` and `AppendEvents` preserve the input snapshot. `Combine` selects
the right state and concatenates left-before-right events. FlatMap, negation,
predicate bridges, rule accumulation, and alias spellings are not part of v0.3.

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
