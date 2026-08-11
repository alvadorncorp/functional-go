---
type: adr
id: adr:workflow:007
context: workflow
title: Functional integration without effect duplication
status: Aceita
tags: [workflow, functional, dependencies, values, boundary]
deciders: [alvadorn]
related: [adr:workflow:002, adr:workflow:005, tdd:workflow:core]
date: 2026-08-10
---
# ADR-007: Integração com functional sem duplicar efeitos

## Contexto e problema

`functional` já fornece `Optional`, `Result`, `Either`, `Validated`,
`Transform`, `Predicate`, coleções imutáveis e `state`. Um novo módulo
funcional-like corre o risco de duplicá-los ou forçar todo consumidor a passar
por eles.

## Drivers da decisão

- um dono por conceito;
- nenhuma cerimônia imposta a operações triviais;
- um núcleo cujo fecho de dependências permanece mínimo.

## Decisão

O núcleo de `workflow` é apenas stdlib: `workflow/go.mod` declara nenhum
`require` e nenhum `replace`, controlado por um teste que analisa o arquivo
nomeado por `go env GOMOD`. O módulo introduz apenas o que precisa de ciclo de
vida de execução — `Action`, `Workflow` e fronteiras de transação — e não
duplica nada de `functional`.

`Action` deliberadamente não é movida para `functional`: carrega
`context.Context` e semântica de execução da aplicação, enquanto `functional`
permanece genérico e livre de qualquer sistema de efeitos.

## Alternativas consideradas

- **Retornar `result.Result[T, E]` de `Run`.** Rejeitado: `(T, error)` é o que
  todo caller Go, adaptador e linter já entende, e `Result` continua disponível
  para aplicações que o queiram.
- **Pontes automáticas `FromResult` / `ToResult`.** Rejeitado: o próprio
  `functional` evita pontes universais entre suas famílias de valor, e
  reintroduzir esse acoplamento lateralmente seria pior.

## Consequências

Um consumidor pode usar `workflow` sem importar `functional`. Comparar os
conceitos vizinhos os mantém distintos: `functional.Transform[T]` é `T -> T`,
`modeling.Change[S]` é `S -> S`, `workflow.Workflow[I,O]` é `I -> Action[O]`, e
`action.Action[O]` é `Context -> (O, error)`.

## Não-objetivos

Nenhuma família de monads, nenhuma emulação higher-kinded e nenhum equivalente de
`Optional`/`Result`/`Either` dentro de `workflow`.
