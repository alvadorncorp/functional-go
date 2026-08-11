---
type: tdd
id: tdd:workflow:composition
context: workflow
title: Workflow composition
status: Aceito
tags: [workflow, composition, generics, closures, reuse]
related: [adr:workflow:001, adr:workflow:003, tdd:workflow:action]
date: 2026-08-10
---
# TDD: Composição de Workflow

## Escopo técnico

```text
Workflow[I,O]
     │  input I
     ▼
Action[O]

Workflow[A,B] + Workflow[B,C] ──► Workflow[A,C]
```

Dependências sobrevivem sem runtime porque um construtor fecha sobre elas. O
valor de função retornado *é* o grafo de dependências, por isso reatribuir a
variável original depois não muda nada.

## Contratos públicos

```go
func From[I, O any](fn func(I) action.Action[O]) Workflow[I, O]
func Map[I, A, B any](wf Workflow[I, A], mapper func(A) B) Workflow[I, B]
func MapFallible[I, A, B any](wf Workflow[I, A], mapper func(A) (B, error)) Workflow[I, B]
func Then[A, B, C any](first Workflow[A, B], second Workflow[B, C]) Workflow[A, C]
```

## Invariantes impostas por testes

- aplicar um workflow a uma entrada não produz efeito;
- um workflow construído serve a muitas entradas e muitos contextos
  concorrentemente;
- dependências capturadas são estáveis — não existe resolução tardia;
- `Then` é determinístico em ordem e curto-circuita na primeira falha;
- um workflow composto é um `Workflow` ordinário, indistinguível de um primitivo
  para fins de composição.

## Por que a composição é binária

Os genéricos de Go não preservam um tipo intermediário que muda ao longo de uma
lista variádica. `Then(a, b, c)` exigiria `any`, reflection ou asserções de tipo
em runtime. Encadear `Then(Then(a, b), c)` mantém cada tipo checado em
compilação, e a ergonomia nunca supera a segurança de tipos aqui.
