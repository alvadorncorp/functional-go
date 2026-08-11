# Plan

`workflow/plan` representa a estrutura conceitual de uma operação como dado Go
imutável e inspecionável.

> Plan descreve. Workflow constrói. Action adia. Run executa.

## Superfície inicial

```go
var CreateOrderPlan = plan.Sequence(
    plan.Step("load-customer"),
    plan.Step("load-product"),
    plan.Step("decide-order"),
    plan.Transaction(
        plan.Step("save-order"),
        plan.Step("append-outbox"),
    ),
)
```

O valor zero é `KindEmpty`. `Sequence()` também é Empty e `Sequence(p)` devolve
`p`. Com dois ou mais filhos, Sequence preserva ordem e aninhamento.
Transaction preserva o nó declarado porque a fronteira conceitual é informação.

`Step` rejeita somente nome vazio. Nomes são destinados a humanos e tooling;
não são UUIDs nem package paths.

## Inspeção e imutabilidade

`Kind()`, `Name()` e `Children()` são as únicas operações de leitura. Os
construtores copiam slices recebidos, e `Children()` devolve cópia defensiva.
Uma árvore pronta pode ser lida concorrentemente sem locks.

## O que Plan não faz

Plan não executa Action, não recebe `context.Context`, não chama `Run`, não abre
transação, não aplica retry/timeout e não realiza I/O. Plan não é uma
`modeling.Specification`: Specification classifica um valor de domínio; Plan
descreve a estrutura operacional declarada pela aplicação.

Plan não é compilado para Action e Action não é decompilada para Plan. A
aplicação pode declarar os dois artefatos lado a lado, ou usar qualquer um
isoladamente.
