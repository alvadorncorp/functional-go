# modeling

`modeling` is an explicit, compile-time-safe Go metamodel for application-owned
domain types and behavior. It sits above selected public capabilities in
[`functional`](../functional/) and below concrete application domains.

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
[`modeling/policy`](policy/) capability. `Policy[I,O]` is a pure function type;
`Constant`, `Map`, and `Contramap` adapt it, while `When`,
`FirstApplicable`, and `Fallback` provide ordered Optional-based selection.
No Policy symbol is reexported from the root package.

## Documentation

- [Documentation index](docs/README.md)
- [Architecture overview](docs/overview.md)
- [Change package guide](docs/packages/change.md)
- [ADR-001: minimal type-safe domain modeling](docs/adr/modeling/adr-001-minimal-type-safe-domain-modeling.md)
- [ADR-002: root facade over an internal Decision package](docs/adr/modeling/adr-002-root-facade-internal-decision.md)
- [ADR-003: nominal rules and operation requirements](docs/adr/modeling/adr-003-rules-and-requirements.md)
- [ADR-004: public Change capability boundary](docs/adr/modeling/adr-004-change-capability-boundary.md)
- [ADR-005: additive Policy capability boundary](docs/adr/modeling/adr-005-policy-capability-boundary.md)
- [TDD: v0.1 core](docs/tdd/modeling/v0-1-core.md)
- [TDD: v0.2 Decision composition](docs/tdd/modeling/v0-2-decision-composition.md)
- [TDD: v0.3 Rules and Requirements](docs/tdd/modeling/v0-3-rules-requirements.md)
- [TDD: v0.5 Change and state evolution](docs/tdd/modeling/v0-5-change-state-evolution.md)
- [TDD: Policy](docs/tdd/modeling/policy.md)

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go mod tidy -diff
```
