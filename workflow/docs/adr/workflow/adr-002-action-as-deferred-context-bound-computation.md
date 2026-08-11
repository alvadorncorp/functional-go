---
type: adr
id: adr:workflow:002
context: workflow
title: Action as deferred context-bound computation
status: Aceita
tags: [workflow, action, effects, context, laziness]
deciders: [alvadorn]
related: [adr:workflow:001, adr:workflow:007, tdd:workflow:action]
date: 2026-08-10
---
# ADR-002: Action como computação adiada ligada a contexto

## Contexto e problema

Um efeito componível precisa de uma representação. Introduzir um tipo de efeito
sob medida forçaria cada consumidor a aprender um novo canal de falha e um novo
mecanismo de cancelamento, ambos já existentes em Go.

## Drivers da decisão

- zero vocabulário novo para falha e cancelamento;
- construção que, de forma demonstrável, não executa trabalho;
- uma saída de emergência de volta ao Go imperativo ordinário;
- a menor representação que ainda suporte composição.

## Decisão

`Action[T] = func(context.Context) (T, error)`. `Run(ctx, action)` é a única
passagem da descrição à execução. Construção e composição não produzem efeito;
apenas `Run` o faz. O tipo é uma função, não um struct, e portanto não carrega
metadados, identificador nem hooks.

## Alternativas consideradas

- **Um struct com campo `run` mais metadados.** Rejeitado: nada no módulo precisa
  dos campos extras, e cada um alocaria em cada composição.
- **Um canal de falha baseado em `Result`.** Rejeitado: `(T, error)` mantém
  `errors.Is`, `errors.As` e os sentinelas padrão de cancelamento funcionando sem
  mudança.
- **Memoizar o resultado.** Rejeitado: isso é `functional/structural/lazy`, um
  conceito diferente. Uma `Action` é reexecutável por desenho.

## Consequências

`construção ≠ execução` é uma propriedade imposta, não uma convenção: um teste com
contador afirma zero callbacks durante a construção. Executar a mesma `Action`
duas vezes executa o efeito duas vezes. Uma `Action` não é future nem promise e
não inicia goroutine. Pânicos nunca são convertidos em erros, e nenhum combinador
chama `recover`.

## Não-objetivos

`Await`, `Start`, `Spawn`, `Promise`, `Future` e `Memoize` estão excluídos, e um
teste de arquitetura rejeita esses nomes.
