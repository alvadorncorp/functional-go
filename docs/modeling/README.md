# modeling documentation

A documentação do módulo é de propriedade daqui. A raiz do repositório
permanece um catálogo de módulos e não deve adquirir uma árvore `docs/`.

## Guides

- [Architecture overview](overview.md)
- [Change package](packages/change.md)
- [Policy package](packages/policy.md)
- [Specification package](packages/specification.md)
- [Reducer package](packages/reducer.md)
- [Projection package](packages/projection.md)
- Modeling-only patterns live in [patterns/](#patterns); effectful or hexagonal recipes stay in [workflow hexagonal patterns](../../workflow/docs/patterns/README.md) (sibling module).

## Patterns

- [Lei decide/evolve](patterns/decide-evolve-law.md)
- [Decider](patterns/decider.md)

## Architecture decisions

- [Minimal explicit and type-safe domain metamodel](../adr/modeling/minimal-type-safe-domain-modeling.md)
- [Root facade over an internal Decision package](../adr/modeling/root-facade-internal-decision.md)
- [Nominal rules and operation requirements](../adr/modeling/rules-and-requirements.md)
- [Public Change capability boundary](../adr/modeling/change-capability-boundary.md)
- [Additive public Policy capability](../adr/modeling/policy-capability-boundary.md)
- [Additive reducer capability boundary](../adr/modeling/reducer-capability-boundary.md)
- [Additive projection capability boundary](../adr/modeling/projection-capability-boundary.md)
- [Nominal and independent Specification capability](../adr/modeling/specification-capability-boundary.md)
- [Closed Policy composition algebra](../adr/modeling/policy-composition-algebra.md)
- [Replay and Project over iter.Seq with a stdlib-only dependency](../adr/modeling/streaming-replay-project-seq.md)

## Technical designs

- [TDD: behavioral core](../tdd/modeling/core.md)
- [TDD: Decision composition](../tdd/modeling/decision-composition.md)
- [TDD: Rules and Requirements](../tdd/modeling/rules-requirements.md)
- [TDD: Change and state evolution](../tdd/modeling/change.md)
- [TDD: Policy](../tdd/modeling/policy.md)
- [TDD: Specification](../tdd/modeling/specification.md)
- [TDD: Reducer and Replay](../tdd/modeling/reducer-replay.md)
- [TDD: Projection and Project](../tdd/modeling/projection.md)

Registros formais usam a spine AKB do módulo; nenhuma cópia pertence à raiz do
repositório.
