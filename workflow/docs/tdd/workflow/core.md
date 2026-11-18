---
type: tdd
id: tdd:workflow:core
context: workflow
title: Workflow module core
status: Accepted
tags: [workflow, core, module, architecture, archtest]
related: [adr:workflow:workflow-as-reusable-effectful-composition, adr:workflow:functional-integration-without-effect-duplication, adr:workflow:composition-and-resource-safety, adr:workflow:plan-as-explicit-non-executable-workflow-description, adr:workflow:explicit-resilience-as-action-transformation, adr:workflow:ast-evidence-for-exercised-surface-gate, tdd:workflow:action, tdd:workflow:composition, tdd:workflow:plan-resilience]
date: 2026-08-11
---
# TDD: Núcleo do módulo workflow

## Escopo técnico

O módulo mantém cinco pacotes de produção e um pacote interno de gates:
`workflow`, `workflow/action`, `workflow/plan`, `workflow/resilience`,
`workflow/transaction` e `workflow/internal/archtest`.

```text
workflow ──► action ◄── transaction
                ▲
                └────── resilience

plan   (folha, sem dependência intramódulo)
```

`action` é o sumidouro e importa exatamente `context` e `errors`. A segunda
dependência existe somente para preservar duas falhas independentes em `Ensure`
e `Bracket`; ambas continuam dentro da biblioteca padrão. `plan` importa
somente `slices` e `strings`; `resilience` importa exatamente `context`,
`action` e `time`. Os contratos de `plan` e `resilience` estão em
[`plan-resilience.md`](plan-resilience.md). `go.mod` não contém `require` nem
`replace`.

## Contratos públicos

```go
type Workflow[I, O any] func(I) action.Action[O]
type Action[T any] func(context.Context) (T, error)
type Boundary[S any] func(context.Context, func(context.Context, S) error) error
```

As superfícies completas são fixadas por `TestExactPublicSurface`; mudanças de
nome, assinatura, field shape ou arquivo precisam alterar o gate na mesma fase.

## Invariantes impostas por testes

- fecho transitivo = stdlib + este módulo;
- DAG intramódulo exatamente `workflow -> action`, `transaction -> action` e
  `resilience -> action`, com `plan` folha e `action` sumidouro;
- nenhum `require`, `replace`, package novo ou global mutável;
- nenhuma goroutine de produção, chamada a `recover` ou criação, derivação,
  proteção ou leitura de valores de contexto (`Background`, `TODO`,
  `WithCancel`, `WithDeadline`, `WithValue`, `Value`, `WithoutCancel`,
  `AfterFunc`); a única exceção é `context.WithTimeout` no pacote `resilience`;
- `Pair` e `AttemptResult` possuem somente os campos decididos e zero métodos;
- cada export possui consumo exercitado fora do próprio pacote, provado por
  `go/parser` sem tipos (`adr:workflow:ast-evidence-for-exercised-surface-gate`): seletor exato de import path e
  símbolo, com alias resolvido, dot import de pacote rastreado rejeitado, import
  em branco ignorado e identificadores sombreados descartados; comentário,
  string, prefixo e pacote estrangeiro não contam. Uma função consumida alcança
  só os tipos locais de seus resultados e um tipo consumido, os tipos de seus
  campos exportados; constantes exigem seletor direto. Os consumidores são a
  produção do módulo e os `example_test.go` externos, exceto
  `internal/archtest`. `go vet ./...` verifica a compilação dos consumidores;
  `go test ./...` executa os exemplos com saída declarada;
- superfícies e arquivos de produção são allowlists exatas; a lista de
  documentos exigidos é verificada somente por presença, junto com chaves de
  frontmatter e tokens selecionados;
- todo gate possui evidência negativa de que pode falhar.

## Características de execução

Construção e composição são lazy; `Run` é a única passagem para efeitos e
repassa o `ctx` à Action sem alteração. Não há preempção automática: o
cancelamento é cooperativo e depende de a Action observar o contexto.
Actions são reexecutáveis, sem memoização nem runtime. Um closure por combinador
é medido por benchmark, sem promessa universal de zero allocation.
