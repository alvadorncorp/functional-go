# Transações

Quatro coisas diferentes compartilham a palavra. Elas não são intercambiáveis:

| Conceito | Responsabilidade |
|---|---|
| `context.Context` | cancelamento, prazo, ciclo de vida da requisição |
| `functional/state.Transaction` | agrupar mutações puras de um plano `State[T]` |
| escopo de transação do workflow | fronteira de composição e capacidades escopadas |
| `postgres.Tx` / `sql.Tx` | a implementação ACID concreta |

Somente o último realiza uma transação física. O escopo é uma abstração para
usá-la sem acoplar-se a ela. `functional/state.Transaction` é um lote puramente
em memória de mutações e não tem nada a ver com um banco.

## A forma

```go
type Boundary[S any] func(context.Context, func(context.Context, S) error) error

func Within[S, T any](boundary Boundary[S], fn func(S) action.Action[T]) action.Action[T]
```

O adaptador fornece `Boundary[S]`; a aplicação define `S`. O handle concreto
nunca aparece em uma assinatura de workflow e nunca é armazenado em um contexto.

Em execução bem-sucedida, a `Boundary` chama o callback exatamente uma vez, e
essa chamada termina de forma síncrona antes de a `Boundary` retornar `nil`. Ela
pode retornar um erro sem chamar o callback, por exemplo quando o `BEGIN` falha.
`Within` detecta a omissão no sucesso, mas não conta nem impede chamadas
repetidas: isso continua sendo contrato do adaptador.

## Estreitamento de capacidades

Ofereça apenas o que é legal onde é legal:

```text
Outside the transaction:  Accounts { Load }
Inside the transaction:   TxAccounts { Save }, TxOutbox { Append }
```

Uma escrita fora do lugar falha ao compilar em vez de escapar da fronteira em
runtime.

## O que uma fronteira não promete

`Boundary` descreve uma região com semântica transacional definida pelo seu
provedor. Uma fronteira PostgreSQL dá uma transação ACID local; um fake em
memória dá uma simulação. Compor duas fronteiras compõe efeitos — não cria uma
transação distribuída, e nenhuma API aqui aceita mais de uma fronteira.

## Leituras dentro ou fora

Ambas as estratégias são suportadas e nenhuma é imposta. Otimista: carregar
fora, decidir, depois `UPDATE ... WHERE version = expected` dentro. Pessimista:
iniciar, `SELECT ... FOR UPDATE`, decidir, atualizar. Um escopo pode expor uma
capacidade de leitura quando a forma pessimista for desejada.

## Fronteiras são curtas

`Within` não é mecanismo para trabalho de longa duração. Manter uma transação
aberta através de uma chamada remota, um sleep ou uma aprovação humana é um
erro; isso é coordenação e precisa de outra estratégia. Da mesma forma, uma
sequência `BEGIN / UPDATE / chamar API externa / publicar no broker / COMMIT` é
erro: uma transação de banco não torna HTTP ou Kafka transacionais. Persista
estado e uma linha de outbox, e deixe a entrega acontecer depois.

Chamadas aninhadas de `Within` não carregam semântica automática de savepoint ou
transação aninhada.

## `Bracket` não é `Within`

`Bracket` descreve acquire/use/release de qualquer recurso e não promete
atomicidade. `Within` executa uma Action dentro de `Boundary[S]`, cuja
implementação define begin, commit e rollback. Um adapter pode usar lifecycle
internamente para implementar sua Boundary, mas os contratos públicos não são
fundidos.
