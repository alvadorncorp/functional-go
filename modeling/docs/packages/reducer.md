# Package `reducer`

**Import path:** `github.com/alvadorncorp/gopackages/modeling/reducer`

Reducer evolui estado de domínio a partir de fatos já ocorridos. É uma
capability de modelagem síncrona e pura, não um runtime de Event Sourcing.

## API pública

```go
type Reducer[S, E any] func(S, E) S

func Replay[S, E any](
    initial S,
    events iter.Seq[E],
    reduce Reducer[S, E],
) S
```

Funções ordinárias `func(S, E) S` funcionam diretamente. S e E são de
propriedade da aplicação. A semântica é de valor: o metamodelo não usa
ponteiro, `Ref`, `Box` nem `Readonly` no tipo `Reducer`.

## Exemplo completo

```go
type Account struct {
    Balance int
    Closed  bool
}

type AccountEvent struct {
    Kind   AccountEventKind
    Amount int
}

func ReduceAccount(state Account, event AccountEvent) Account {
    switch event.Kind {
    case AccountOpened:
        state.Balance = event.Amount
    case MoneyDeposited:
        state.Balance += event.Amount
    case AccountClosed:
        state.Closed = true
    }
    return state
}

state := reducer.Replay(
    Account{},
    events, // iter.Seq[AccountEvent]
    ReduceAccount,
)
```

`ReduceAccount` não lê database, broker, Event Store, relógio, filesystem nem
registry global. A aplicação fornece os fatos e decide o que fazer com o
estado devolvido.

## Uso incremental e em lote

A função Reducer já é a operação incremental:

```go
next := ReduceAccount(current, event)
```

Replay é a forma em lote, de passagem única:

```go
rebuilt := reducer.Replay(empty, history, ReduceAccount)
```

Ambos compartilham a semântica `S x E -> S`. Replay aplica cada evento
exatamente uma vez na ordem de encontro da sequência e encadeia o resultado.
Sequência vazia devolve o estado inicial.

## Ponte com Decision

`Decision[S,E].Events()` já é `iter.Seq[E]`. Quando esses eventos reconstróem
estado:

```go
state := reducer.Replay(previous, decision.Events(), ReduceAccount)
```

Não é necessário `immutable.CollectList`. Quem precisar varrer a sequência
mais de uma vez materializa explicitamente antes de chamar Replay.

A consistência entre `decision.State()` e esse Replay é lei da aplicação, não
do pacote: Replay não compara os dois estados e não há `VerifyDecision` nem
`ReplayDecision`. As receitas de teste (tabela e fuzz) estão em
[Lei decide/evolve](../patterns/decide-evolve-law.md).

## Reducer versus Projection

Ambos têm forma `valor x evento -> valor`, com ownership distinto:

- `reducer.Reducer[S,E]` evolui estado de domínio a partir de fatos;
- `projection.Projection[P,E]` evolui uma representação derivada a partir de
  fatos.

A mesma história pode ser reaplicada em estado e projetada em read model de
forma independente. Os pacotes não se importam.

## Eventos ignorados e desconhecidos

Um Reducer ignora um evento devolvendo o estado corrente:

```go
default:
    return state
```

A aplicação é dona de uniões heterogêneas e da política de kind desconhecido.
O pacote não adiciona erro, optional, marker, registry nem reflection para
esse caso.

## Dependência e custo

Produção importa somente a biblioteca padrão (`iter`). Replay roda em
`O(n * cost(reduce))`, consome a sequência uma vez e não materializa `[]E`
nem List intermediária.

Leis:

```text
Replay(s, empty, r) = s
Replay(s, A ++ B, r) = Replay(Replay(s, A, r), B, r)
```

Para `r` determinístico, entradas iguais produzem estados iguais.

Benchmarks de `Replay` cobrem N ∈ {0, 1, 10, 100, 1000, 10000} com
`b.ReportAllocs()`. Os números abaixo são evidência local (não limiar de CI).

### Baseline local (darwin/arm64, Apple M2 Max, go1.26.4)

Comando:

```bash
go test -bench=BenchmarkReplayInt -benchtime=50ms -count=1 ./reducer/
```

| N | ns/op | B/op | allocs/op |
|---|------:|-----:|----------:|
| 0 | 2.1 | 0 | 0 |
| 1 | 2.1 | 0 | 0 |
| 10 | 5.0 | 0 | 0 |
| 100 | 44 | 0 | 0 |
| 1000 | 348 | 0 | 0 |
| 10000 | 3182 | 0 | 0 |

`Replay` sobre `int` não aloca no hot path; o custo cresce linearmente com N.
Valores medidos em 2026-08-12; reexecute o comando ao comparar outra máquina.

## Valores grandes e `hugeParam`

`hugeParam` (e regras similares de lint) é um guarda-corpo arquitetural sobre
o **tamanho de valores concretos** na aplicação — não um motivo para
ponteirizar `Reducer` ou o metamodelo. Estados grandes devem ser modelados
pela aplicação (por exemplo, estrutura interna adequada); o pacote não
autoriza `//nolint` nem `Reducer[*S, E]` como “solução”.

## Fronteira de Event Sourcing

Replay é um fold puro. A aplicação carrega fatos, invoca Replay e persiste o
resultado se desejar. Loader de stream, Event Store, snapshot, serialização,
versão, publicação, handler e transporte ficam fora deste pacote.

## Exclusões deliberadas

Não há Apply, FallibleReducer, sobrecarga de slice/List, ReplayDecision,
VerifyDecision, marker Event, registry, dispatcher, projection embutida,
persistência, broker consumer nem runtime de Event Sourcing.
