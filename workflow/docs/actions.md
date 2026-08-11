# Actions

## Execução adiada

```go
program := loadUser(id)
// No I/O has happened.

user, err := action.Run(ctx, program)
// Now it has.
```

## Propagação de contexto

O contexto dado a `Run` chega a toda `Action` derivada. Quando um cliente se
desconecta, o cancelamento viaja até o driver do banco inalterado: este módulo
nunca engole nem reclassifica `context.Canceled` ou `context.DeadlineExceeded`.

## Composição

`Pure` eleva um valor. `Map` transforma um resultado. `MapFallible` aplica uma
função pura falível — a costura onde uma decisão de domínio entra em uma cadeia
com efeitos. `FlatMap` sequencia um efeito dependente. `Recover` substitui uma
falha.

Uma falha curto-circuita tudo a jusante, de modo que uma cadeia se lê de cima
para baixo e para no primeiro problema.

## Saída de emergência

Combinadores são opcionais. Quando Go ordinário é mais claro, escreva-o:

```go
func complexOperation(deps Deps) action.Action[Output] {
    return action.Do(func(ctx context.Context) (Output, error) {
        // ordinary imperative Go, including early returns
    })
}
```

O resultado ainda é uma `Action` e ainda se compõe com todo o resto. É isso que
impede a biblioteca de se tornar um DSL fechado.

## Não é future, não é valor lazy

`functional/structural/lazy` é adiado *e memoizado*. Uma `Action` é adiada,
reexecutável, ligada a contexto e falível. Executá-la duas vezes executa o efeito
duas vezes, de propósito.
