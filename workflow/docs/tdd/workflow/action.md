---
type: tdd
id: tdd:workflow:action
context: workflow
title: Action representation and execution
status: Aceito
tags: [workflow, action, composition, resources, context, errors, panic]
related: [adr:workflow:002, adr:workflow:009, adr:workflow:013, tdd:workflow:core, tdd:workflow:composition]
date: 2026-08-11
---
# TDD: Representação e execução de Action

## Escopo técnico

`Action[T] = func(context.Context) (T, error)`. É um valor reexecutável e sem
estado interno; somente `Run` executa o closure. Todo combinador encaminha o
mesmo contexto e nunca cria goroutine, runtime, memoização ou recovery.

## Contratos públicos

```go
func Run[T any](context.Context, Action[T]) (T, error)
func Do[T any](func(context.Context) (T, error)) Action[T]
func Pure[T any](T) Action[T]
func Fail[T any](error) Action[T]
func Map[A, B any](Action[A], func(A) B) Action[B]
func MapFallible[A, B any](Action[A], func(A) (B, error)) Action[B]
func FlatMap[A, B any](Action[A], func(A) Action[B]) Action[B]
func Recover[T any](Action[T], func(error) Action[T]) Action[T]
type Pair[A, B any] struct { First A; Second B }
func Zip[A, B any](Action[A], Action[B]) Action[Pair[A, B]]
func Both[A, B, C any](Action[A], Action[B], func(A, B) C) Action[C]
func Traverse[A, B any]([]A, func(A) Action[B]) Action[[]B]
func Tap[T any](Action[T], func(T) Action[struct{}]) Action[T]
func Ensure[T any](Action[T], Action[struct{}]) Action[T]
type AttemptResult[T any] struct { Value T; Err error }
func Attempt[T any](Action[T]) Action[AttemptResult[T]]
func Bracket[R, T any](Action[R], func(R) Action[T], func(R) Action[struct{}]) Action[T]
func AcquireUseRelease[R, T any](Action[R], func(R) Action[T], func(R) Action[struct{}]) Action[T]
```

## Ordem e curto-circuito

`Zip` e `Both` executam primeiro a Action esquerda e depois a direita. `Both`
chama o combiner somente após dois sucessos. `Tap` chama o efeito somente após o
sucesso da source. Independência de valores não significa paralelismo.

`Traverse` vincula e executa um item por vez, em ordem de índice: o binder do
item seguinte só é chamado depois do sucesso da Action do item anterior. Cada
`Run` percorre os itens de novo. Sucesso devolve um slice novo, não nil, com
`len(items)` valores na ordem dos itens; entrada nil ou vazia devolve slice vazio
não nil sem chamar o binder. A primeira falha devolve `(nil, err)` com o erro
inalterado, sem chamar o binder para os itens restantes. Não há checagem de
contexto entre itens. A construção copia `items` (cópia rasa: elementos ponteiro
continuam compartilhados).

Perfil de alocação de `Traverse`: uma cópia de `items` na construção, um slice de
resultado por `Run` e as alocações feitas pelo binder e pelas Actions que ele
devolve.

## Erros e finalização

Erros únicos preservam identidade. `Ensure` executa o finalizer depois de todo
retorno normal da source. `Bracket` executa release se e somente se acquire
retornar sucesso e também depois de um erro normal de use. Quando source/use e
finalizer/release falham, `errors.Join(primaryErr, finalizerErr)` preserva ambos
para `errors.Is`/`errors.As`; qualquer erro produz zero `T`.

`Attempt` transforma o par normal `(value, err)` em `AttemptResult` e retorna
erro externo nil. Não recupera panic nem implementa política best-effort.

## Contexto, panic e nil

O contexto de `Run`, inclusive cancelado, chega inalterado a toda etapa.
Shielding pertence ao adapter. Nenhum combinador chama `recover`; `Ensure` não
promete finalizer depois de panic e `Bracket` não promete release depois de
panic. Actions/callbacks nil falham na construção; callbacks que retornam Action
nil falham durante execução com `workflow/action: nil action`. O binder nil de
`Traverse` falha na construção com `workflow/action: nil binder`, mesmo com
`items` nil.
