---
type: adr
id: adr:workflow:012
context: workflow
title: Sem módulo DDD/hexagonal de runtime
status: Aceita
tags: [workflow, modeling, functional, ddd, hexagonal, boundary]
deciders: [alvadorn]
related: [adr:modeling:001, adr:workflow:003, adr:workflow:005, adr:workflow:006, adr:workflow:008, adr:workflow:011, adr:modeling:008, adr:modeling:009, adr:functional:010]
date: 2026-08-24
---
# ADR-012: Sem módulo DDD/hexagonal de runtime

## Contexto e problema

Aplicações neste workspace já se constroem com `functional` (valores),
`modeling` (decisão pura) e `workflow` (efeitos), compostos pela aplicação,
como demonstra `examples/workflow/checkout`. A tentação seguinte é um quarto
módulo com `Aggregate`, `Entity`, `Repository`, `UnitOfWork` ou `CommandBus`.

## Drivers da decisão

- um conceito, um dono;
- tipos Go da aplicação para identidade, agregado e dinheiro;
- gates executáveis em vez de marker interfaces;
- não reabrir ADRs aceitas para acomodar um framework.

## Decisão

Não se cria módulo `ddd`, `hexagonal`, `aggregate` ou `repository` com tipos
de domínio. A ajuda a aplicações DDD/hexagonais é documentação de padrões,
exemplos, e — só com gatilho próprio — um kit de archtest, combinadores de
compensação, ou slugs de exemplo.

Permanece proibido no núcleo de `workflow` (e não se introduz em `modeling`
nem `functional`) exportar: `Repository`, `CommandBus`, `Outbox`, `Inbox`,
`CommitDecision`, `Handle` como tipo-base, `Clock`, `Saga`, `SagaEngine`,
`DistributedTransaction`, `EventStore`. `Xor` em `modeling/specification` e
circuit breaker / jitter / bulkhead em `workflow/resilience` permanecem fora
(`adr:modeling:008`, `adr:workflow:011`). Money, Coupon e identificadores
ficam na aplicação (`adr:modeling:001`).

Onda futuras nomeadas, não deste ADR: kit `archkit` (test-only); pacote
`workflow/saga` exportando `Compensate` (nunca o identificador `Saga`);
exemplos `starter` e `readmodel`.

## Alternativas consideradas

- **Um módulo runtime DDD.** Rejeitado: cerimônia sem validade de domínio e
  em choque com `adr:modeling:001` e `adr:workflow:003`.
- **Só silêncio.** Rejeitado: a próxima sessão reabriria a questão.

## Consequências

Guias em `workflow/docs/patterns/` e `functional/docs/patterns/` passam a ser
o onboarding. Extrações de código exigem gatilho e ADR própria.

## Não-objetivos

Nenhum `CommitDecision` compartilhado até ≥3 domínios concordarem na forma
(`adr:workflow:005` — hoje não concordam). Nenhuma unificação do outbox da
aplicação com `eventbus/outbox`.
