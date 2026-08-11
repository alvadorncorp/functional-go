---
type: tdd
id: tdd:workflow:action
context: workflow
title: Action representation and execution
status: Aceito
tags: [workflow, action, laziness, context, errors, panic]
related: [adr:workflow:002, tdd:workflow:core, tdd:workflow:composition]
date: 2026-08-10
---
# TDD: Representação e execução de Action

## Escopo técnico

`Action[T]` é um tipo função, não um struct: não carrega identificador, metadados
nem hooks, e a composição custa um closure por combinador.

```text
Action creation
     │
     X  no execution
     │
     ▼
Action value
     │
     ▼
Run(ctx)
     │
     ▼
callback(ctx)
```

## Contratos públicos

```go
func Run[T any](ctx context.Context, action Action[T]) (T, error)
func Do[T any](fn func(context.Context) (T, error)) Action[T]
func Pure[T any](value T) Action[T]
func Fail[T any](err error) Action[T]
func Map[A, B any](action Action[A], mapper func(A) B) Action[B]
func MapFallible[A, B any](action Action[A], mapper func(A) (B, error)) Action[B]
func FlatMap[A, B any](action Action[A], binder func(A) Action[B]) Action[B]
func Recover[T any](action Action[T], recovery func(error) Action[T]) Action[T]
```

## Invariantes impostas por testes

- **Laziness.** Construir ou compor uma `Action` não executa callback.
- **Propagação de contexto.** Toda `Action` derivada recebe o contexto dado a
  `Run`, inalterado; nenhum combinador deriva um novo contexto.
- **Curto-circuito.** Em uma cadeia sequencial, uma falha impede que qualquer
  callback a jusante rode.
- **Identidade de erro.** Nenhum combinador embrulha um erro, então `errors.Is` e
  `errors.As` continuam funcionando de ponta a ponta.
- **Cancelamento.** `context.Canceled` e `context.DeadlineExceeded` chegam ao
  caller inalterados e nunca são reclassificados.
- **Reexecução.** Executar uma `Action` duas vezes executa o efeito duas vezes;
  nada é memoizado.
- **Política de pânico.** Um pânico propaga intocado; `Recover` trata apenas
  `error` e nenhum combinador chama `recover()`.

## Comportamento com nil

Todo combinador rejeita `Action` ou callback nil ansiosamente na construção com
mensagem estável: `workflow/action: nil action`, `nil function`, `nil error`,
`nil mapper`, `nil binder`, `nil recovery`. Um binder
ou recovery que retorne nil na execução entra em pânico com
`workflow/action: nil action`.
