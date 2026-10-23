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

`Kind()`, `Name()`, `Children()` e `String()` são as operações de leitura de
Plan. `Kind.String()` devolve `empty`, `step`, `sequence` ou `transaction`;
valores de Kind não definidos devolvem `unknown`.
Os construtores copiam slices recebidos, e `Children()` devolve cópia defensiva.
Uma árvore pronta pode ser lida concorrentemente sem locks; `String()` não
modifica a árvore nem snapshots obtidos por `Children()`.

`Plan.String()` descreve a estrutura declarada em uma forma canônica:

```go
fmt.Println(CreateOrderPlan.String())
// sequence(load-customer, load-product, decide-order, transaction(save-order, append-outbox))
```

O valor zero e `Sequence()` produzem `empty`; um Step produz seu nome sem
alterações. Sequence mantém as regras do construtor: um único filho é reduzido
ao próprio filho, e sequências aninhadas com dois ou mais filhos permanecem
aninhadas. Transaction preserva sua fronteira, inclusive `transaction()` sem
filhos. Nomes aparecem literalmente, sem escape; a descrição não promete parsing
ou round-trip e não acrescenta semântica de execução.

## O que Plan não faz

Plan não executa Action, não recebe `context.Context`, não chama `Run`, não abre
transação, não aplica retry/timeout e não realiza I/O. Plan não é uma
`modeling.Specification`: Specification classifica um valor de domínio; Plan
descreve a estrutura operacional declarada pela aplicação.

Plan não é compilado para Action e Action não é decompilada para Plan. A
aplicação pode declarar os dois artefatos lado a lado, ou usar qualquer um
isoladamente.
