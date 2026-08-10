# modeling

`modeling` is an explicit, compile-time-safe Go metamodel for application-owned
domain types and behavior. It sits above the public capabilities in
[`functional`](../functional/) and below concrete application domains.

The module intentionally contains no reflection, runtime registry, dispatcher,
infrastructure, persistence, or concrete business model.

Decision composition is available from the root package through `MapState`,
`MapEvents`, `Emit`, `AppendEvents`, and `Combine`. The canonical public type
remains `modeling.Decision`; persistent mechanics are isolated behind a Go
internal package and are not a second consumer API.

Rules and operation preconditions are explicit root-package functions:
`Rule[T]` composes value conditions with `AllRules` and non-empty `AnyRule`;
`Requirement[S,I]` composes state-plus-input preconditions with
`AllRequirements` and non-empty `AnyRequirement`. Applications explicitly run
Requirement, Transition, Invariant, and Decision; modeling provides no runner.

## Documentation

- [Documentation index](docs/README.md)
- [Architecture overview](docs/overview.md)
- [ADR-001: minimal type-safe domain modeling](docs/adr/modeling/adr-001-minimal-type-safe-domain-modeling.md)
- [ADR-002: root facade over an internal Decision package](docs/adr/modeling/adr-002-root-facade-internal-decision.md)
- [TDD: v0.1 core](docs/tdd/modeling/v0-1-core.md)
- [TDD: v0.2 Decision composition](docs/tdd/modeling/v0-2-decision-composition.md)
- [ADR-003: nominal rules and operation requirements](docs/adr/modeling/adr-003-rules-and-requirements.md)
- [TDD: v0.3 Rules and Requirements](docs/tdd/modeling/v0-3-rules-requirements.md)

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go mod tidy -diff
```
