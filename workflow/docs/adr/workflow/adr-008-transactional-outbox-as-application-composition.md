---
type: adr
id: adr:workflow:008
context: workflow
title: Transactional outbox as application composition
status: Aceita
tags: [workflow, outbox, integration-events, transaction, examples]
deciders: [alvadorn]
related: [adr:workflow:004, adr:workflow:006, tdd:workflow:transaction]
date: 2026-08-10
---
# ADR-008: Outbox transacional como composição da aplicação

## Contexto e problema

Publicar um evento de integração de dentro de uma transação de banco é um padrão
comum e incorreto: uma chamada a um broker não pode ser revertida. O outbox
transacional corrige isso, e a questão é se o módulo deve fornecê-lo.

## Drivers da decisão

- provar que o núcleo é expressivo o bastante para um padrão real;
- evitar uma forma universal de outbox com a qual duas aplicações não concordam.

## Decisão

O outbox é um exemplo, não um primitivo. `Outbox`, `OutboxMessage`,
`OutboxRepository` e `OutboxDispatcher` são nomes exportados proibidos no núcleo.
O exemplo define seu próprio `OutboxMessage` e sua própria capacidade
`EventEncoder`, e `Workflow` mais um escopo de transação mais capacidades da
aplicação bastam para implementar o padrão corretamente.

## Alternativas consideradas

- **Um `CommitDecision[S,E,O]` genérico no núcleo.** Rejeitado após comparar os
  domínios de saque e cross-context: se toda decisão é persistida, se todo evento
  é externo, se o estado é escrito antes dos eventos e qual é o tipo de saída
  diferem. Um ajudante "genérico" teria codificado um estilo.
- **Publicar em um broker dentro da fronteira.** Rejeitado como o próprio erro
  que o padrão existe para prevenir.

## Consequências

Estado e mensagens compartilham uma transação física e confirmam ou revertem
juntos, provado contra PostgreSQL real. A ordem de eventos de
`modeling.Decision` é preservada porque a sequência é iterada diretamente e os
appends são sequenciais. Uma decisão que emite zero eventos não escreve linhas.
Persistir o outbox e entregá-lo são preocupações separadas, e nenhuma garantia
exatamente-uma-vez ponta a ponta é alegada em lugar algum.

## Não-objetivos

Nenhum runtime de dispatcher, nenhuma abstração de broker, nenhum framework de
inbox e nenhuma garantia de entrega além de at-least-once do transporte que a
aplicação escolher.
