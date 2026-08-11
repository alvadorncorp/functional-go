---
type: tdd
id: tdd:workflow:core
context: workflow
title: Workflow module core
status: Aceito
tags: [workflow, core, module, architecture, archtest]
related: [adr:workflow:001, adr:workflow:007, adr:workflow:009, tdd:workflow:action, tdd:workflow:composition]
date: 2026-08-11
---
# TDD: Núcleo do módulo workflow

## Escopo técnico

O módulo mantém três pacotes de produção e um pacote interno de gates:
`workflow`, `workflow/action`, `workflow/transaction` e
`workflow/internal/archtest`.

```text
workflow ──► action ◄── transaction
```

`action` importa exatamente `context` e `errors`. A segunda dependência existe
somente para preservar duas falhas independentes em `Ensure` e `Bracket`; ambas
continuam dentro da biblioteca padrão. `go.mod` não contém `require` nem
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
- DAG intramódulo exatamente `workflow -> action <- transaction`;
- nenhum `require`, `replace`, package novo ou global mutável;
- nenhuma goroutine de produção, chamada a `recover` ou uso de
  `context.Background`, `TODO`, `WithValue`, `Value` ou `WithoutCancel`;
- `Pair` e `AttemptResult` possuem somente os campos decididos e zero métodos;
- cada export possui um consumidor compilado fora do próprio pacote;
- superfícies, arquivos e documentos são allowlists exatas;
- todo gate possui evidência negativa de que pode falhar.

## Características de execução

Construção e composição são lazy; `Run` é a única passagem para efeitos.
Actions são reexecutáveis, sem memoização nem runtime. Um closure por combinador
é medido por benchmark, sem promessa universal de zero allocation.
