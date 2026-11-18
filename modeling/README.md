# modeling

`modeling` is an explicit, compile-time-safe Go metamodel for application-owned
domain types and behavior. It sits above selected public capabilities in
[`functional`](../README.md) and below concrete application domains.

The module intentionally contains no reflection, runtime registry, dispatcher,
state manager, infrastructure, persistence, or concrete business model.

The root package owns Decision, Transition, Invariant, Rule, and Requirement.
Decision composition is available through `MapState`, `MapEvents`, `Emit`,
`AppendEvents`, and `Combine`; persistent mechanics remain private under
`internal/decision`.

Pure state evolution is available from the public capability package
[`modeling/change`](docs/packages/change.md). `Change[S]` names an endomorphism
`S -> S`; `Identity`, `Compose`, `Pipe`, `Apply`, and `ApplyAll` provide closed
composition and value-first application without a runtime or root re-export.

Rules and operation preconditions remain explicit root-package functions:
`Rule[T]` composes value conditions with `AllRules` and non-empty `AnyRule`;
`Requirement[S,I]` composes state-plus-input preconditions with
`AllRequirements` and non-empty `AnyRequirement`. Applications explicitly run
Requirement, Transition, Invariant, Change, and Decision as their domain flow
requires; modeling provides no runner.

Context-derived domain strategies live in the public
[`modeling/policy`](docs/packages/policy.md) capability. `Policy[I,O]` is a pure
function type; `Map`/`Contramap`, `Compose`/`Zip`, and
`Choose`/`When`/`OrElse`/`FirstApplicable`/`Fallback` provide the complete
closed algebra. No Policy symbol is reexported from the root package.

Boolean domain classifications live in the independent public
[`modeling/specification`](docs/packages/specification.md) capability.
`Specification[T]` is nominal `T -> bool`; `All`, `Any`, `Not`, and `Contramap`
compose classifications without validation, errors, effects, sibling imports,
or root reexports.

The additive [`modeling/reducer`](reducer/) capability defines
`Reducer[S,E] func(S,E) S` and ordered `Replay` over `iter.Seq[E]`.
Reducer materializes facts into state; Replay is a synchronous left fold, not an
Event Sourcing runtime. Every existing root-package, change, and policy contract
remains unchanged.

The orthogonal [`modeling/projection`](projection/) capability defines
`Projection[P,E] func(P,E) P` and ordered `Project` over `iter.Seq[E]`.
Projection derives read models and summaries; it is not a Reducer alias,
persistence adapter, broker consumer, or CQRS runtime. No Projection symbol is
reexported from the root package.

## Documentation

- [Documentation index](docs/README.md)
- [Architecture overview](docs/overview.md)
- [Change package guide](docs/packages/change.md)
- [Policy package guide](docs/packages/policy.md)
- [Specification package guide](docs/packages/specification.md)
- [Reducer package guide](docs/packages/reducer.md)
- [Pattern: decide/evolve law](docs/patterns/decide-evolve-law.md)
- [Pattern: Decider](docs/patterns/decider.md)
- [ADR-001: minimal type-safe domain modeling](docs/adr/modeling/adr-001-minimal-type-safe-domain-modeling.md)
- [ADR-002: root facade over an internal Decision package](docs/adr/modeling/adr-002-root-facade-internal-decision.md)
- [ADR-003: nominal rules and operation requirements](docs/adr/modeling/adr-003-rules-and-requirements.md)
- [ADR-004: public Change capability boundary](docs/adr/modeling/adr-004-change-capability-boundary.md)
- [ADR-005: additive Policy capability boundary](docs/adr/modeling/adr-005-policy-capability-boundary.md)
- [ADR-008: nominal Specification capability](docs/adr/modeling/adr-008-specification-capability-boundary.md)
- [ADR-009: completed Policy composition algebra](docs/adr/modeling/adr-009-policy-composition-algebra.md)
- [TDD: behavioral core](docs/tdd/modeling/core.md)
- [TDD: Decision composition](docs/tdd/modeling/decision-composition.md)
- [TDD: Rules and Requirements](docs/tdd/modeling/rules-requirements.md)
- [TDD: Change and state evolution](docs/tdd/modeling/change.md)
- [TDD: Policy](docs/tdd/modeling/policy.md)
- [TDD: Specification](docs/tdd/modeling/specification.md)
- [ADR-006: additive reducer capability boundary](docs/adr/modeling/adr-006-reducer-capability-boundary.md)
- [TDD: Reducer and Replay](docs/tdd/modeling/reducer-replay.md)
- [Projection package guide](docs/packages/projection.md)
- [ADR-007: additive projection capability boundary](docs/adr/modeling/adr-007-projection-capability-boundary.md)
- [TDD: Projection and Project](docs/tdd/modeling/projection.md)
- [ADR-010: Replay/Project over iter.Seq (stdlib only)](docs/adr/modeling/adr-010-streaming-replay-project-seq.md)

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go mod tidy -diff
```
