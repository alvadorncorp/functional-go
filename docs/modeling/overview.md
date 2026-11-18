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
      │                     ↓ uses
      │               stdlib iter
      ├── specializes → modeling/projection
      │                     ↓ uses
      │               stdlib iter
      └── specializes → modeling/specification
                            (stdlib slices only)
```

Applications import the root package for Decision, Transition, Invariant, Rule,
and Requirement. They import `modeling/change` for named state evolution,
`modeling/policy` for context-derived strategies, `modeling/reducer` for
fact-to-state replay, `modeling/projection` for fact-to-derived-representation
projection, and `modeling/specification` for nominal boolean classification.
The root does not reexport Change, Policy, Reducer, Replay, Projection, Project,
or Specification. The internal Decision package stays an implementation
boundary, not another public API.

## Structural semantics

Value, Identity, Entity, Aggregate, and Model remain Go types of the
application. The module exports no marker interfaces, mandatory embeddings,
base types, or reflection-based equality. An aggregate is a consistency
boundary, not a database row or a repository record.

## Behavioral primitives

The root's public contract contains:

- `Decision[S,E]` with a valid zero value, a constructor, observations, and
  typed composition functions;
- `Rule[T]` for an explanatory condition over a value, composed through
  ordered `AllRules` or non-empty `AnyRule`;
- `Requirement[S,I]` for an operational precondition over state plus input,
  composed through `AllRequirements` or non-empty `AnyRequirement`;
- `Transition[S,I,O]` and `FallibleTransition[S,I,O]` for next-state
  computation;
- `Invariant[S]` for validity of the resulting state, with fail-fast or
  error-accumulating composition.

Capabilities in subpackages:

- `reducer.Reducer[S,E]` and `reducer.Replay`, a synchronous left fold over
  `iter.Seq[E]` (see the Reducer section);
- `projection.Projection[P,E]` and `projection.Project`, the matching fold for
  derived representations over `iter.Seq[E]`;
- `change.Change[S]` with Identity, Compose, Pipe, Apply, and ApplyAll;
- `policy.Policy[I,O]` with composition combinators;
- `specification.Specification[T]` with All/Any/Not/Contramap.

Rule and Invariant stay distinct nominal types even though both return error.
Requirement answers whether an input may be applied to the current state; it is
not a Transition hook. Change neither decides applicability nor validates or
emits events. Ordinary functions stay assignable, and domain errors belong to
the application.

## Policy capability

`modeling/policy` represents `I -> O`: a pure strategy that derives a domain
result from explicit input. Map transforms output, Contramap adapts input,
Compose chains decisions, Zip combines two decisions over the same input,
Choose selects one total strategy, and When/OrElse/FirstApplicable/Fallback
model ordered applicability. Policy neither validates nor runs effects.

`modeling/specification` represents nominal `T -> bool`: boolean domain
classification. All/Any evaluate left-to-right with short-circuit, Not derives
the complementary category, and Contramap adapts input. The package is a leaf
whose only production import is the stdlib `slices`, and it is independent of
Policy and of every sibling.

Policy asks which result or strategy to use. Specification asks whether a
value belongs to a domain category. Rule explains a failed condition,
Requirement authorizes an operation, Invariant validates the resulting state,
and Change performs `S -> S`. Identical underlying signatures never erase
these semantic responsibilities.

## Reducer and Replay

`Reducer[S,E]` is an ordinary `func(S,E) S`: a fact that has already happened
materializes into the next state. It is deliberately distinct from Transition
(intention/input), Decision (state plus emitted events), and Policy (strategy
from context).

`reducer.Replay(initial, events, reduce)` applies each event of an
`iter.Seq[E]` exactly once, in encounter order, in a single pass. Empty Replay
returns the initial state; concatenating sequences (where the application
defines `A ++ B`) is equivalent to replaying the left side and then the right.
The Decision bridge is direct:

```go
state := reducer.Replay(previous, decision.Events(), reduce)
```

`immutable.CollectList` is unnecessary. Anyone needing multiple passes
materializes explicitly before calling Replay.

Whether that Replay must reproduce `decision.State()` is an application law,
verified in the application's tests: see the
[decide/evolve law](patterns/decide-evolve-law.md).
The library makes no such comparison. The
[Decider](patterns/decider.md) pattern shows how the application composes
Transition, Decision, and Reducer into `decide`/`evolve` without a library
type.

Production code of `modeling/reducer` depends only on the stdlib (`iter`).
Replay is a pure fold, not an Event Sourcing runtime: it loads no streams,
persists nothing, serializes no versions, creates no snapshots, registers no
handlers, and does not compare Decision state with reduced state.

## Projection and Project

`Projection[P,E]` is an ordinary `func(P,E) P` that derives the next read
model, summary, index, statistic, or other application-owned representation
from a fact. `projection.Project(initial, events, project)` applies the same
operation to an `iter.Seq[E]` in encounter order, in a single pass, so batch
rebuild and incremental update share one semantic function.

Projection is distinct from Reducer. Reducer explains what the facts mean for
domain state; Projection explains what they mean for a derived representation.
The two packages are siblings: neither imports the other nor the modeling
root, and each consumes application events independently.

The Decision bridge:

```go
view := projection.Project(empty, decision.Events(), project)
```

Empty input returns `initial`; the concatenation law is
`Project(p,A++B,f) = Project(Project(p,A,f),B,f)` where the application defines
the concatenation. Determinism and shallow immutability remain the
responsibility of the supplied function.

Production code of `modeling/projection` depends only on the stdlib (`iter`).
Projection is not a CQRS runtime and does not persist. Repositories, databases,
brokers, subscriptions, offsets, checkpoints, event loading, background
processing, and eventual consistency stay outside modeling.

## Large values and `hugeParam`

The `hugeParam` lint rule (and similar rules) is an architectural guardrail
over the size of concrete values the application chooses to copy by signature,
not a reason to pointerize the metamodel. `Reducer`,
`Projection`, `Change`, and the other capability types keep value semantics
(`func`/values). Large states or projections must be modeled in the
application; `//nolint` and `Ref`/`Box`/`Readonly` are not authorized in the
modeling contract to work around the lint.

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

The application invokes each step explicitly. modeling has no pipeline,
runner, command handler, rules engine, registry, dispatcher, state manager,
persistence, or transport.

AllRules and AllRequirements preserve declaration order, stop at the first
failure, preserve error identity, and accept zero operands. AnyRule and
AnyRequirement require a first operand, stop at the first success, and return
the last evaluated error when all fail. Nil constituents panic when the
composition is built, aligned with the Invariant programmer-error policy.

Change uses a separate contract: Identity represents the absence of change;
Compose runs right-to-left; Pipe and ApplyAll run left-to-right; empty
composition returns Identity; empty ApplyAll returns its input. A nil Change
has ordinary Go nil-function behavior, not Rule/Requirement-style eager
validation.

`MapState` changes only the Decision state; `MapEvents` changes only the
events, in order. `Emit` and `AppendEvents` preserve the input snapshot.
`Combine` selects the right state and concatenates events left-before-right.
FlatMap, negation, predicate bridges, rule accumulation,
conditional/guarded changes, validation hooks, and aliases stay out of
Change's scope.

## Immutability and cost

The root's public Decision is a facade over a private internal Value backed by
an immutable list. Events stay exposed only as `iter.Seq[E]`, which keeps
modeling independently evolvable. Decision immutability is shallow: referenced
state or event values are the caller's responsibility.

Combinators copy the retained variadic slices before nil validation; later
changes to the caller's slice do not affect the built composition.

Change is also conceptually and shallowly immutable. Structs may be copied
before updating; slices, maps, pointers, and shared references remain the
responsibility of the Change author. There is no reflection and no deep copy.
Composition is a linear pass over the supplied functions and defines no
machine-independent performance ceiling.

## Non-goals

The module has no reflection, annotations, scanner, registry, service locator,
DI container, magic dispatch, command bus, event bus, event store, worker, or
projection runtime, repository, ORM, database, HTTP, broker, filesystem, or
concrete application domain type.

Library Entity/Aggregate/Handle types stay out; see
adr:workflow:no-ddd-runtime-module and functional/docs/patterns/cow-handles.md
for the handle pattern in the application.

Change is not a Patch, a Diff, a state manager, a Reducer implementation, a
Replay engine, a Transition replacement, an event producer, a validator, a
guard, a runtime, a builder, or a fluent API.

`functional/state` and `functional/immutable` are optional capabilities for
applications with larger state. They are not superclasses, and their internal
engines never belong to modeling. `functional/primitives/function.Transform`
supplies Change's private composition algorithms but is never aliased,
reexported, or exposed by the modeling capability. Reducer and projection, in
turn, do not depend on `functional` in production.
