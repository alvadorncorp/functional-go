---
type: tdd
id: tdd:workflow:core
context: workflow
title: Workflow module core
status: Aceito
tags: [workflow, core, module, architecture, archtest]
related: [adr:workflow:001, adr:workflow:007, tdd:workflow:action, tdd:workflow:composition]
date: 2026-08-10
---
# TDD: Núcleo do módulo workflow

## Escopo técnico

O módulo embarca três pacotes importáveis e um pacote interno de gates:
`workflow` (raiz), `workflow/action`, `workflow/transaction` e
`workflow/internal/archtest`. O DAG de dependências tem `action` como sumidouro:

```text
workflow ──► action ◄── transaction
```

`action` importa exatamente `context`. Nenhuma outra aresta intramódulo existe, e
as arestas inversas são rejeitadas por nome.

## Contratos públicos

```go
type Workflow[I, O any] func(I) action.Action[O]
type Action[T any] func(context.Context) (T, error)
type Boundary[S any] func(context.Context, func(context.Context, S) error) error
```

## Invariantes impostas por testes

- o fecho transitivo de dependências é a biblioteca padrão mais este módulo;
- `go.mod` não contém linha `require` nem `replace`;
- nenhum arquivo de produção importa `database/sql`, `net/http`, `os`, `reflect`,
  `unsafe`, `sync`, `sync/atomic`, `time`, `log/slog` ou `encoding/json`;
- nenhum arquivo de produção referencia `context.WithValue` nem lê valor de
  contexto;
- não existe `var` de nível de pacote em produção;
- cada pacote exporta exatamente o conjunto de símbolos decidido;
- cada pacote mantém exatamente o conjunto de arquivos decidido, e nenhum arquivo
  excede 500 linhas;
- cada gate tem um caso negativo que prova que pode falhar.

## Características de alocação e profundidade

A composição aloca um closure por combinador; isso é inerente à representação e é
medido em vez de eliminado. Uma cadeia de 10 000 estágios sequenciais executa sem
esgotar a pilha, o que é o limite medido, não uma garantia de profundidade
ilimitada.
