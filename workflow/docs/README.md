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
- [Exemplos](examples.md) — withdraw, crosscontext e CreateOrder, de ponta a ponta
- [Padrões hexagonais](patterns/README.md) — receitas de aplicação (não tipos de biblioteca)

## Decisões de arquitetura

- [ADR-001 — Workflow como composição reutilizável com efeitos](adr/workflow/adr-001-workflow-as-reusable-effectful-composition.md)
- [ADR-002 — Action como computação adiada ligada a contexto](adr/workflow/adr-002-action-as-deferred-context-bound-computation.md)
- [ADR-003 — Injeção funcional de dependências por capacidades](adr/workflow/adr-003-functional-dependency-injection-through-capabilities.md)
- [ADR-004 — Escopos de transação explícitos e capacidades escopadas](adr/workflow/adr-004-explicit-transaction-scopes.md)
- [ADR-005 — Fronteira de integração com modeling](adr/workflow/adr-005-modeling-integration-boundary.md)
- [ADR-006 — Sem abstração de transação distribuída](adr/workflow/adr-006-no-distributed-transaction-abstraction.md)
- [ADR-007 — Integração com functional sem duplicar efeitos](adr/workflow/adr-007-functional-integration-without-effect-duplication.md)
- [ADR-008 — Outbox transacional como composição da aplicação](adr/workflow/adr-008-transactional-outbox-as-application-composition.md)
- [ADR-009 — Composição avançada e resource safety em retornos normais](adr/workflow/adr-009-composition-and-resource-safety.md)
- [ADR-010 — Plan como descrição explícita e não executável de workflow](adr/workflow/adr-010-plan-as-explicit-non-executable-workflow-description.md)
- [ADR-011 — Resiliência explícita como transformação de Action](adr/workflow/adr-011-explicit-resilience-as-action-transformation.md)
- [ADR-012 — Sem módulo DDD/hexagonal de runtime](adr/workflow/adr-012-no-ddd-runtime-module.md)
- [ADR-013 — Traverse sequencial sobre slices](adr/workflow/adr-013-sequential-traverse-over-slices.md)

## Desenhos técnicos

- [Núcleo](tdd/workflow/core.md)
- [Action](tdd/workflow/action.md)
- [Composição](tdd/workflow/composition.md)
- [Transação](tdd/workflow/transaction.md)
- [Integração com modeling](tdd/workflow/modeling-integration.md)
- [Plan e Resilience](tdd/workflow/plan-resilience.md)
