# Package `projection`

**Import path:** `github.com/alvadorncorp/gopackages/modeling/projection`

Projection deriva read models, sumários, índices e estatísticas de propriedade
da aplicação a partir de fatos. É uma capability de modelagem síncrona e pura,
não um adaptador de persistência nem runtime CQRS.

## API pública

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events iter.Seq[E],
    project Projection[P, E],
) P
```

Funções ordinárias `func(P, E) P` funcionam diretamente. P pode ser struct,
primitivo ou coleção persistente da aplicação; E permanece inteiramente
application-owned. A semântica é de valor: o metamodelo não usa ponteiro,
`Ref`, `Box` nem `Readonly` no tipo `Projection`.

## Exemplo completo

```go
type SalesSummary struct {
    Orders  int
    Revenue Money
}

func ProjectSales(
    summary SalesSummary,
    event OrderEvent,
) SalesSummary {
    switch event.Kind {
    case OrderCreated:
        summary.Orders++
    case OrderPaid:
        summary.Revenue = summary.Revenue.Add(event.Amount)
    }
    return summary
}

summary := projection.Project(
    SalesSummary{},
    events, // iter.Seq[OrderEvent]
    ProjectSales,
)
```

`ProjectSales` não lê database, broker, Event Store, relógio, filesystem nem
registry global. A aplicação fornece os fatos e decide o que fazer com o valor
devolvido.

## Outras formas de representação

Uma Projection não precisa devolver um struct de view. Uma estatística
primitiva é válida:

```go
var CountPaidOrders projection.Projection[int, OrderEvent] =
    func(total int, event OrderEvent) int {
        if event.Kind == OrderPaid {
            return total + 1
        }
        return total
    }
```

Um índice pode usar coleção persistente da aplicação:

```go
type ProductIndex = immutable.Map[ProductID, ProductView]
```

Views de cliente, dashboards, valores monetários acumulados e outras
representações derivadas seguem o mesmo contrato; modeling não define o que
conta como read model.

## Uso incremental e em lote

A função Projection já é a operação incremental:

```go
next := ProjectSales(current, event)
```

Project é a forma em lote, de passagem única:

```go
rebuilt := projection.Project(empty, history, ProjectSales)
```

Ambos compartilham a semântica `P x E -> P`. Project aplica eventos exatamente
uma vez na ordem de encontro da sequência e encadeia cada resultado. Entrada
vazia devolve o valor inicial.

## Ponte com Decision

`Decision[S,E].Events()` já é `iter.Seq[E]`:

```go
view := projection.Project(empty, decision.Events(), ProjectSales)
```

Não é necessário `immutable.CollectList`. Multi-pass exige materialização
explícita antes de Project.

## Projection versus Reducer

Ambos têm forma `valor x evento -> valor`, com ownership distinto:

- `reducer.Reducer[S,E]` evolui estado de domínio a partir de fatos;
- `projection.Projection[P,E]` evolui uma representação derivada a partir de
  fatos.

A mesma história de eventos pode ser reaplicada em estado de domínio e
projetada em read model de forma independente. Os pacotes não se importam.

## Eventos ignorados e desconhecidos

Uma Projection ignora um evento devolvendo a representação corrente:

```go
default:
    return current
```

A aplicação é dona de uniões heterogêneas e da política de kind desconhecido.
O pacote não adiciona erro, optional, marker, registry nem reflection para
esse caso.

## Coleções persistentes

P pode conter ou ser um `immutable.Map`, List ou Set da aplicação. Devolver o
novo valor usa o compartilhamento estrutural dessas coleções. Projection não
converte coleções persistentes em slices ou maps mutáveis e **não** depende de
`functional/immutable` em produção — quem usa Map/List o faz no código da
aplicação.

## Dependência e custo

Produção importa somente a biblioteca padrão (`iter`). Project roda em
`O(n * cost(f))`, consome a sequência uma vez e não materializa a lista de
eventos.

Leis:

```text
Project(p, empty, f) = p
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

Para `f` determinístico, entradas iguais produzem resultados iguais.
Benchmarks de `Project` cobrem N ∈ {0, 1, 10, 100, 1000, 10000} com
`b.ReportAllocs()`. Os números abaixo são evidência local (não limiar de CI).

### Baseline local (darwin/arm64, Apple M2 Max, go1.26.4)

Comando:

```bash
go test -bench=BenchmarkProjectInt -benchtime=50ms -count=1 ./projection/
```

| N | ns/op | B/op | allocs/op |
|---|------:|-----:|----------:|
| 0 | 2.3 | 0 | 0 |
| 1 | 2.1 | 0 | 0 |
| 10 | 4.8 | 0 | 0 |
| 100 | 34 | 0 | 0 |
| 1000 | 331 | 0 | 0 |
| 10000 | 3172 | 0 | 0 |

`Project` sobre `int` não aloca no hot path; o custo cresce linearmente com N.
Valores medidos em 2026-08-12; reexecute o comando ao comparar outra máquina.

## Valores grandes e `hugeParam`

`hugeParam` é um guarda-corpo sobre o tamanho de valores concretos na
aplicação — não um motivo para ponteirizar `Projection` ou o metamodelo.
Representações grandes devem ser modeladas pela aplicação; o pacote não
autoriza `//nolint` nem `Projection[*P, E]` como atalho.

## Persistência e fronteira CQRS

Fluxo típico da aplicação:

```text
carregar fatos/view corrente -> Projection -> salvar view devolvida
```

Somente a transformação do meio pertence a este pacote. Carga, repositórios,
escritas em database, subscriptions, offsets, checkpoints, retries, consumer
groups, workers em background e consistência eventual permanecem em camadas de
aplicação ou infraestrutura.

## Exclusões deliberadas

Não há FallibleProjection, ContramapEvent, Compose, Zip, Map, ContramapState,
alias Replay, API paralela, marker ReadModel/View/Event, registry, dispatcher,
handler, persistência, broker consumer, Event Store nem runtime CQRS neste
pacote.
