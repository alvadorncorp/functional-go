---
type: adr
id: adr:workflow:003
context: workflow
title: Functional dependency injection through capabilities
status: Aceita
tags: [workflow, dependency-injection, capabilities, composition-root]
deciders: [alvadorn]
related: [adr:workflow:001, adr:workflow:004, tdd:workflow:composition]
date: 2026-08-10
---
# ADR-003: Injeção funcional de dependências por capacidades

## Contexto e problema

A injeção de dependências em Go costuma ser resolvida com contêiner, registro ou
service locator. Cada um move erros de fiação do tempo de compilação para o de
execução e torna o grafo de dependências de uma operação ilegível a partir da
assinatura.

## Drivers da decisão

- grafo de dependências visível na assinatura do construtor;
- dublês de teste sem framework de mocking;
- detecção em tempo de compilação de dependência ausente;
- nenhum estado mutável global.

## Decisão

Dependências são fornecidas como funções, closures e pequenos structs de
capacidade, montados por uma composition root da aplicação. Uma capacidade é um
struct de funções como `Accounts{ Load func(AccountID) action.Action[Account] }`.
O módulo não define porta universal: `Repository`, `Store`, `CommandBus` e
parentes são nomes exportados proibidos.

## Alternativas consideradas

- **Um contêiner com `Register` / `Resolve`.** Rejeitado: falhas de fiação em
  runtime, reflection e registro global, todos proibidos por nome no archtest.
- **Interfaces como mecanismo mandatório de injeção.** Rejeitado como
  *mandatório*, não como opção: uma interface continua correta onde há
  polimorfismo genuíno. Simplesmente não é exigida para injeção.
- **Passar dependências por `context.Context`.** Rejeitado: contexto carrega
  ciclo de vida, não capacidades, e valores de contexto derrotam a checagem em
  compilação.

## Consequências

Construir um workflow com capacidades falsas não exige framework: um literal de
struct com closures é o dublê. Como as dependências são capturadas por closure na
construção, o grafo é estável — nenhuma reatribuição posterior altera o que um
workflow já construído usa.

## Não-objetivos

Nenhum pacote `workflow/di`, nenhuma anotação, nenhuma geração de código e nenhuma
fiação automática.
