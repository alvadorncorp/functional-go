---
type: tdd
id: tdd:workflow:composition
context: workflow
title: Workflow composition
status: Accepted
tags: [workflow, composition, inputs, generics, closures, reuse]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-dependency-injection-through-capabilities, adr:workflow:composition-and-resource-safety, tdd:workflow:action]
date: 2026-08-11
---
# TDD: Composição de Workflow

## Escopo técnico

```text
Workflow[I,O] + I       --Provide-->   Action[O]
Workflow[I,O] + (A→I)  --Contramap--> Workflow[A,O]
Workflow[A,B] + Workflow[B,C] --Then--> Workflow[A,C]
```

Dependências entram na construção, input entra na aplicação e contexto entra em
`Run`. Aplicar um workflow pode projetar o input e construir uma Action, mas não
executa o efeito descrito por ela.

## Contratos públicos

```go
func From[I, O any](func(I) action.Action[O]) Workflow[I, O]
func Map[I, A, B any](Workflow[I, A], func(A) B) Workflow[I, B]
func MapFallible[I, A, B any](Workflow[I, A], func(A) (B, error)) Workflow[I, B]
func Then[A, B, C any](Workflow[A, B], Workflow[B, C]) Workflow[A, C]
func Provide[I, O any](Workflow[I, O], I) action.Action[O]
func Contramap[A, I, O any](Workflow[I, O], func(A) I) Workflow[A, O]
```

## Leis

- `Contramap(wf, identity)` é observacionalmente equivalente a `wf`;
- contramaps aninhados equivalem à composição dos mappers;
- `Provide(wf, input)` equivale a aplicar `wf(input)`, incluindo resultado,
  erro, contexto e contagem de execução;
- `Contramap` chama mapper na aplicação, não na construção;
- `Provide` e `Contramap` nunca chamam `Run` nem executam efeitos.

## Limites

Composição permanece binária e type-safe. Não existe lista variádica heterogênea,
reflection, resolução tardia de dependências, DI container, engine ou runtime.
`Provide` fornece input; não registra nem resolve dependências.
