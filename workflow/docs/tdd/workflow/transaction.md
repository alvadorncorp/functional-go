---
type: tdd
id: tdd:workflow:transaction
context: workflow
title: Transaction boundary and scoped capabilities
status: Aceito
tags: [workflow, transaction, boundary, scope, outbox, postgres]
related: [adr:workflow:004, adr:workflow:006, adr:workflow:008, tdd:workflow:core]
date: 2026-08-10
---
# TDD: Fronteira de transação e capacidades escopadas

## Escopo técnico

```text
transaction provider (adapter)
        ↓
concrete transaction
        ↓
scope construction  (application-defined S)
        ↓
scoped capabilities
        ↓
Action[T]
        ↓
commit / rollback
```

A abstração — `Boundary[S]` e `Within[S,T]` — pertence a este módulo. A
implementação pertence a um adaptador, e a forma de `S` pertence à aplicação.

## Contratos públicos

```go
type Boundary[S any] func(context.Context, func(context.Context, S) error) error
func Within[S, T any](boundary Boundary[S], fn func(S) action.Action[T]) action.Action[T]
```

`Within` carrega `T` para fora através de um callback que só retorna `error`
atribuindo-o no closure interno, o que permite a um resultado tipado atravessar
uma fronteira cuja assinatura devolve apenas `error`.

## Invariantes impostas por testes

- construir `Within` não inicia transação;
- cada execução recebe um escopo novo;
- o commit ocorre só no sucesso; qualquer erro reverte;
- um pânico reverte e é então reerguido com seu valor original;
- o cancelamento chega ao adaptador e nunca vira erro de negócio;
- o escopo é inalcançável através do contexto.

## Obrigações do adaptador

`postgres.InTx` satisfaz `Boundary` depois de fechar sobre `db`. Ele inicia,
confirma em erro nulo e reverte a partir de um `defer` usando um contexto
desacoplado do cancelamento, para que o rollback ainda rode após desconexão do
cliente. Também instala a transação como valor ambiente de contexto — esse é o
mecanismo próprio do módulo `postgres` para participação aninhada, e é ortogonal
a este módulo, que nem escreve nem lê valores de contexto.

## Separação de conceitos

| Conceito | Responsabilidade |
|---|---|
| `context.Context` | cancelamento, prazo, ciclo de vida da requisição |
| `functional/state.Transaction` | agrupar mutações puras de um plano `State[T]` |
| escopo de transação do workflow | fronteira de composição e capacidades escopadas |
| `postgres.Tx` / `sql.Tx` | a implementação ACID concreta |

Somente o último realiza uma transação física. `functional/state.Transaction` é
não relacionado e nunca é reutilizado para este propósito.
