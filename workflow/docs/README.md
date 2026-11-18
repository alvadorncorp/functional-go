# Documentação do workflow

## Guias

- [Visão geral](overview.md) — a tríade, os três momentos e o glossário
- [Arquitetura](architecture.md) — camadas, propriedade e núcleo funcional / casca com efeitos
- [Actions](actions.md) — execução adiada, contexto, composição, saída de emergência
- [Plan](plan.md) — descrição estrutural explícita, imutável e não executável
- [Resilience](resilience.md) — Retry e Timeout como transformações explícitas de Action
- [Composição](composition.md) — construindo um vocabulário reutilizável de workflows
- [Injeção de dependências](dependency-injection.md) — capacidades e a composition root
- [Transações](transactions.md) — fronteiras, escopos e as quatro coisas chamadas "transação"
- [Integração com modeling](modeling.md) — carregar, decidir, confirmar
- [Exemplos executáveis](examples.md) — composição, transação em memória e resiliência
- [Padrões hexagonais](patterns/README.md) — receitas de aplicação (não tipos de biblioteca)

## Decisões de arquitetura

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

## Desenhos técnicos

- [Núcleo](tdd/workflow/core.md)
- [Action](tdd/workflow/action.md)
- [Composição](tdd/workflow/composition.md)
- [Transação](tdd/workflow/transaction.md)
- [Integração com modeling](tdd/workflow/modeling-integration.md)
- [Plan e Resilience](tdd/workflow/plan-resilience.md)
