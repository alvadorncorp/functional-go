# Workflow documentation

## Guides

- [Overview](overview.md) — the triad, the three moments, and the glossary
- [Architecture](architecture.md) — layers, ownership, and the functional core / effectful shell
- [Actions](actions.md) — deferred execution, context, composition, and the escape hatch
- [Plan](plan.md) — an explicit, immutable, non-executable structural description
- [Resilience](resilience.md) — Retry and Timeout as explicit Action transformations
- [Composition](composition.md) — building a reusable workflow vocabulary
- [Dependency injection](dependency-injection.md) — capabilities and the composition root
- [Transactions](transactions.md) — boundaries, scopes, and the four things called "transaction"
- [modeling integration](modeling.md) — load, decide, confirm
- [Runnable examples](examples.md) — composition, an in-memory transaction, and resilience
- [Hexagonal patterns](patterns/README.md) — application recipes (not library types)

## Architecture decisions

- [Workflow as reusable effectful composition](adr/workflow/workflow-as-reusable-effectful-composition.md)
- [Action as deferred context-bound computation](adr/workflow/action-as-deferred-context-bound-computation.md)
- [Functional dependency injection through capabilities](adr/workflow/functional-dependency-injection-through-capabilities.md)
- [Explicit transaction scopes and scoped capabilities](adr/workflow/explicit-transaction-scopes.md)
- [Modeling integration boundary](adr/workflow/modeling-integration-boundary.md)
- [No distributed transaction abstraction](adr/workflow/no-distributed-transaction-abstraction.md)
- [Functional integration without effect duplication](adr/workflow/functional-integration-without-effect-duplication.md)
- [Transactional outbox as application composition](adr/workflow/transactional-outbox-as-application-composition.md)
- [Advanced composition and normal-return resource safety](adr/workflow/composition-and-resource-safety.md)
- [Plan as explicit non-executable workflow description](adr/workflow/plan-as-explicit-non-executable-workflow-description.md)
- [Explicit resilience as Action transformation](adr/workflow/explicit-resilience-as-action-transformation.md)
- [No DDD/hexagonal runtime module](adr/workflow/no-ddd-runtime-module.md)
- [Sequential Traverse over slices](adr/workflow/sequential-traverse-over-slices.md)
- [AST evidence for the exercised-surface gate](adr/workflow/ast-evidence-for-exercised-surface-gate.md)

## Technical designs

- [Core](tdd/workflow/core.md)
- [Action](tdd/workflow/action.md)
- [Composition](tdd/workflow/composition.md)
- [Transaction](tdd/workflow/transaction.md)
- [modeling integration](tdd/workflow/modeling-integration.md)
- [Plan and Resilience](tdd/workflow/plan-resilience.md)
