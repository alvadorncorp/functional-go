# Actions

## Execução adiada

```go
program := loadUser(id)
// No I/O has happened.

user, err := action.Run(ctx, program)
// Now it has.
```

Somente `Run` inicia a execução. `Do` adapta funções com resultado
`func(context.Context) (T, error)`; `Exec` adapta efeitos sem resultado
`func(context.Context) error` e devolve `Action[struct{}]`:

```go
notify := action.Exec(func(ctx context.Context) error {
    return mailer.Send(ctx, receipt)
})
// Nenhum envio ocorreu.
_, err := action.Run(ctx, notify)
```

`Exec` não é outra entrada de execução: valida a função na construção e adia
o efeito até `Run`. Função nil provoca panic imediato com
`workflow/action: nil function`, como em `Do`. Cada `Run` chama a função de novo;
o contexto, o erro e qualquer panic seguem inalterados, sem recovery nem
checagem adicional de cancelamento.

## Propagação de contexto

O contexto dado a `Run` chega a toda `Action` derivada. Quando um cliente se
desconecta, o cancelamento viaja até o driver do banco inalterado: este módulo
nunca engole nem reclassifica `context.Canceled` ou `context.DeadlineExceeded`.

## Composição

`Pure` eleva valor; `Map`/`MapFallible` transformam; `FlatMap` sequencia uma
Action dependente. `Zip` executa duas Actions independentes em ordem e preserva
ambos os resultados; `Both` executa na mesma ordem e chama um combiner. Não há
paralelismo implícito.

`Traverse(items, binder)` vincula cada item de um slice a uma Action e as
executa uma depois da outra, em ordem de índice, coletando um valor por item.
A primeira falha interrompe o percurso e devolve o erro inalterado; os itens
restantes não são vinculados. O slice é copiado na construção, e cada `Run`
percorre os itens de novo.

`Tap` executa um efeito apenas depois de sucesso e preserva o valor. `Attempt`
entrega o par normal valor/erro como `AttemptResult`, sem recuperar panic.
`Ensure` executa um finalizer depois de retorno normal. `Bracket` associa
explicitamente acquire, use e release; `AcquireUseRelease` é o mesmo algoritmo
com nome descritivo.

Erros únicos permanecem idênticos. Dupla falha de trabalho e finalização usa
`errors.Join`, preservando ambos. O mesmo contexto de `Run`, ainda que cancelado,
chega a toda etapa; shielding pertence ao adapter.

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
