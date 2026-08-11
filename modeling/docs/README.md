# modeling documentation

The module documentation is owned here. The repository root remains a module
catalog and must not acquire a `docs/` tree.

## Guides

- [Architecture overview](overview.md)
- [Change package](packages/change.md)
- [Policy package](packages/policy.md)
- [Specification package](packages/specification.md)
- [Projection package](packages/projection.md)

## Architecture decisions

- [ADR-001: minimal type-safe domain modeling](adr/modeling/adr-001-minimal-type-safe-domain-modeling.md)
- [ADR-002: root facade over an internal Decision package](adr/modeling/adr-002-root-facade-internal-decision.md)
- [ADR-003: nominal rules and operation requirements](adr/modeling/adr-003-rules-and-requirements.md)
- [ADR-004: public Change capability boundary](adr/modeling/adr-004-change-capability-boundary.md)
- [ADR-005: additive Policy capability boundary](adr/modeling/adr-005-policy-capability-boundary.md)
- [ADR-006: additive reducer capability boundary](adr/modeling/adr-006-reducer-capability-boundary.md)
- [ADR-007: additive projection capability boundary](adr/modeling/adr-007-projection-capability-boundary.md)
- [ADR-008: nominal Specification capability](adr/modeling/adr-008-specification-capability-boundary.md)
- [ADR-009: completed Policy composition algebra](adr/modeling/adr-009-policy-composition-algebra.md)

## Technical designs

- [TDD: behavioral core](tdd/modeling/core.md)
- [TDD: Decision composition](tdd/modeling/decision-composition.md)
- [TDD: Rules and Requirements](tdd/modeling/rules-requirements.md)
- [TDD: Change and state evolution](tdd/modeling/change.md)
- [TDD: Policy](tdd/modeling/policy.md)
- [TDD: Specification](tdd/modeling/specification.md)
- [TDD: Reducer and Replay](tdd/modeling/reducer-replay.md)
- [TDD: Projection and Project](tdd/modeling/projection.md)

Formal records use the module-owned AKB spine; no copy belongs at repository root.
