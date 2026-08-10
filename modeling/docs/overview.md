# Architecture overview

`modeling` provides a metamodel, not an application domain:

```text
application domain
      ├── specializes modeling root primitives
      ├── composes modeling/change values
      └── uses modeling/policy strategies

modeling root facade
      ↓ delegates privately
modeling/internal/decision
      ↓ uses public capabilities
functional/immutable

modeling/change
      ↓ privately adapts
functional/primitives/function.Transform v1.1.0

modeling/policy
      ↓ uses public capabilities
functional/values/optional
functional/primitives/predicate
```

Applications import the root package for Decision, Transition, Invariant, Rule,
and Requirement. They import `modeling/change` when they need named pure state
evolution and `modeling/policy` for context-derived strategies. The root does
not re-export Change or Policy symbols. The internal Decision package remains
an enforceable implementation boundary, not another public API.

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
(`S × I → O`), Change (`S → S`), or Reducer (`S × E → S`). It executes no
effects and imports no sibling modeling capability. Applicability is Optional,
never error. Multiple inputs belong in an application-owned context struct.

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
store, projection, repository, ORM, database, HTTP, broker, filesystem, or
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
