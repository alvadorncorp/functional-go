# Architecture overview

`modeling` provides a metamodel, not an application domain:

```text
application domain
      ├── specializes → modeling root facade
      │                     ↓ delegates privately
      │               modeling/internal/decision
      │                     ↓ uses
      │               functional/immutable
      ├── specializes → modeling/change
      │                     ↓ privately adapts
      │               functional/primitives/function
      ├── specializes → modeling/policy
      │                     ↓ uses
      │               functional/primitives/predicate
      │               functional/values/optional
      ├── specializes → modeling/reducer
                            ↓ uses
                     functional/immutable
      └── specializes → modeling/projection
                            ↓ uses
                     functional/immutable
```

Applications import the root package for Decision, Transition, Invariant, Rule,
and Requirement. They import `modeling/change` when they need named pure state
evolution, `modeling/policy` for context-derived strategies,
`modeling/reducer` for fact-to-state replay, and `modeling/projection` for
fact-to-derived-representation projection. The root does not re-export Change,
Policy, Reducer, Projection, or Project symbols. The internal Decision package
remains an enforceable implementation boundary, not another public API.

## Structural semantics

Value, Identity, Entity, Aggregate, and Model remain application-owned Go
types. The module does not export marker interfaces, required embeddings, base
types, or reflection-based equality. An aggregate is a consistency boundary,
not a database row or repository record.

## Behavioral primitives

The root public contract contains:

- `Decision[S,E]` with its valid zero value, constructor, observations, and
  typed composition functions;
- `Rule[T]` for an explanatory condition over one value, composed by ordered
  `AllRules` or non-empty `AnyRule`;
- `Requirement[S,I]` for an operation precondition over state plus input,
  composed by `AllRequirements` or non-empty `AnyRequirement`;
- `Transition[S,I,O]` and `FallibleTransition[S,I,O]` for next-state computation;
- `Invariant[S]` for resulting-state validity, with fail-fast or accumulating
  error composition.
- `reducer.Reducer[S,E]` for fact-to-state evolution and
  `reducer.Replay` for a synchronous ordered left fold over
  `immutable.List[E]` (see the dedicated reducer section).
- `projection.Projection[P,E]` for fact-to-derived-representation evolution and
  `projection.Project` for the corresponding ordered batch fold.

The `modeling/change` public capability contains exactly `Change[S]`,
`Identity`, `Compose`, `Pipe`, `Apply`, and `ApplyAll`. Change maps `S -> S`,
composes within one state type, and captures any operation context in an
ordinary closure.

Rule and Invariant deliberately remain distinct nominal types even though both
return error. Requirement answers whether an input may be applied to current
state; it is not a Transition hook. Change neither decides applicability nor
validates or emits events. Ordinary functions remain assignable and all domain
errors remain application-owned.

## Policy capability

`modeling/policy` represents `I → O`: a pure strategy that derives a domain
result from one explicit context. `Constant`, `Map`, and `Contramap` reuse and
adapt policies. `When` turns a Policy into `I → Optional[O]`,
`FirstApplicable` selects the first Some in declaration order, and `Fallback`
restores a total Policy.

Policy is not Rule (`T → error`), Requirement (`S × I → error`), Transition
(`S × I → O`), Change (`S → S`), Reducer (`S × E → S`), or Projection
(`P × E → P`). It executes no effects and imports no sibling modeling
capability. Applicability is Optional, never error. Multiple inputs belong in
an application-owned context struct.

## Reducer and Replay

`Reducer[S,E]` is an ordinary `func(S,E) S`: a fact that already occurred is
materialized into the next state. It is deliberately distinct from Transition,
which computes from intention/input, from Decision, which carries state plus
emitted events, and from Policy, which derives a strategy from context.

`reducer.Replay(initial, events, reduce)` applies each `immutable.List` event
exactly once in encounter order. Empty replay returns the initial state, and
replaying concatenated lists is equivalent to replaying the left list and then
the right list. Applications own heterogeneous event unions and may convert a
Decision event sequence with `immutable.CollectList(decisionValue.Events())`
when those events reconstruct state.

Replay is a pure fold, not an Event Sourcing runtime. It does not load streams,
persist events, serialize versions, create snapshots/projections, register
handlers, dispatch messages, or compare Decision state with reduced state.

## Projection and Project

`Projection[P,E]` is an ordinary `func(P,E) P` that derives the next read model,
summary, index, statistic, or other application-owned representation from a
fact. `projection.Project(initial, events, project)` applies the same operation
to an `immutable.List` in encounter order, so batch rebuild and incremental
updates share one semantic function.

Projection is deliberately distinct from Reducer. Reducer explains what facts
mean for domain state; Projection explains what they mean for a derived
representation. The two packages are siblings, import neither one another nor
the modeling root, and independently consume application-owned events.

Project iterates `events.Values()` directly. Empty input returns initial; the
concatenation law is `Project(p,A++B,f) = Project(Project(p,A,f),B,f)`.
Determinism and shallow immutability remain the supplied function's
responsibility.

Projection is not a CQRS runtime and performs no persistence. Repositories,
databases, brokers, subscriptions, offsets, checkpoints, event loading,
background processing, and eventual consistency stay outside modeling.

## Explicit domain flow

```text
Current State + Input
          ↓
Requirement[S,I] ── failure → application error
          ↓ valid
Transition[S,I,O] or captured Change[S]
          ↓
Invariant[O] ────── failure → application error
          ↓ valid
Decision[O,E]
```

The application explicitly invokes each step. modeling has no pipeline,
runner, command handler, rules engine, registry, dispatcher, state manager,
persistence, or transport responsibility.

AllRules and AllRequirements preserve declaration order, stop at the first
failure, preserve its identity, and accept zero operands. AnyRule and
AnyRequirement require a first operand, stop at the first success, and return
the final evaluated error when all operands fail. Nil constituents panic while
composition is built, matching the existing Invariant programmer-error policy.

Change uses a separate contract: Identity represents no change; Compose runs
right-to-left; Pipe and ApplyAll run left-to-right; empty composition returns
Identity; empty ApplyAll returns its input. A nil Change gets ordinary Go
nil-function behavior rather than Rule/Requirement eager validation.

`MapState` changes only Decision state; `MapEvents` changes only events in
order. `Emit` and `AppendEvents` preserve the input snapshot. `Combine` selects
the right state and concatenates left-before-right events. FlatMap, negation,
predicate bridges, rule accumulation, conditional changes, guarded changes,
validation hooks, and alias spellings are out of scope for Change.

## Immutability and cost

The public root Decision is a facade over a private internal Value backed by an
immutable list. Events remain exposed only as `iter.Seq[E]`, so modeling stays
independently evolvable. Decision immutability is shallow: referenced state or
event values remain the caller's responsibility.

Change is also conceptually and shallowly immutable. Struct values can be
copied before update, while slices, maps, pointers, and shared references remain
the Change author's responsibility. No reflection or deep copy is attempted.
Composition is a linear pass over the supplied functions and defines no
machine-independent performance cap.

## Non-goals

The module has no reflection, annotations, scanner, registry, service locator,
dependency-injection container, magic dispatch, command bus, event bus, event
store, projection worker/runtime, repository, ORM, database, HTTP, broker, filesystem, or
concrete application-domain type.

Change is not a Patch, Diff, state manager, Reducer implementation, Replay
engine, Transition replacement, event producer, validator, guard, runtime,
builder, or fluent API. The Change capability does not add Reducer/Replay or
release/tag work.

`functional/state` and `functional/immutable` are optional capabilities for
applications with larger state. They are not superclasses, and their internal
engines never belong to modeling. `functional/primitives/function.Transform`
supplies the private composition algorithms for Change, but it is never aliased,
re-exported, or exposed by the modeling capability.
