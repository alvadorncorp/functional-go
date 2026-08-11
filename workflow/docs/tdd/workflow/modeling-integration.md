---
type: tdd
id: tdd:workflow:modeling-integration
context: workflow
title: Modeling integration pattern
status: Aceito
tags: [workflow, modeling, decision, domain, outbox]
related: [adr:workflow:005, adr:workflow:008, tdd:workflow:transaction]
date: 2026-08-10
---
# TDD: Padrão de integração com modeling

## Escopo técnico

A forma canônica é aquisição com efeito, decisão pura, confirmação com efeito:

```text
Action[Account]
      ↓
Account
      ↓
domain.Withdraw          pure, fallible
      ↓
Decision[Account, Event]
      ↓
Action[Receipt]          inside one transaction boundary
```

Nenhum pacote no módulo `workflow` importa `modeling`. A aplicação é a camada de
composição que importa ambos.

## Propriedade

| Conceito | Dono |
|---|---|
| `Rule`, `Requirement`, `Transition`, `Invariant`, `Policy`, `Decision` | `modeling` |
| regra ou limiar de negócio concreto | domínio da aplicação |
| capacidades `Load`, `Save` | aplicação |
| `OutboxMessage`, `EventEncoder` | aplicação |
| `Workflow`, `Action`, fronteira de transação | `workflow` |
| SQL, HTTP, broker | adaptador |

## Invariantes demonstradas pelos exemplos

- o domínio compila e é testado sem `context`, `workflow`, banco ou transporte;
- uma decisão recusada retorna antes de qualquer transação ser aberta;
- a ordem de eventos de `Decision.Events()` é preservada quando as mensagens são
  persistidas, porque o `iter.Seq` é iterado diretamente e os appends são
  sequenciais;
- uma decisão que emite zero eventos não produz linhas de outbox;
- um ramo escolhido por um `policy.Policy` é executado pelo workflow, enquanto o
  limiar que o selecionou permanece no domínio;
- `modeling` não precisou de mudança para ser utilizável aqui — a fronteira
  estaria errada se tivesse precisado.
