# Package `projection`

**Import path:** `github.com/alvadorncorp/gopackages/modeling/projection`

Projection derives application-owned read models, summaries, indexes, and
statistics from facts. It is a pure synchronous modeling capability, not a
persistence adapter or CQRS runtime.

## Public API

```go
type Projection[P, E any] func(P, E) P

func Project[P, E any](
    initial P,
    events immutable.List[E],
    project Projection[P, E],
) P
```

Ordinary `func(P,E) P` functions work directly. P may be a struct, primitive,
or persistent collection; E remains entirely application-owned.

## Complete example

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
    events,
    ProjectSales,
)
```

`ProjectSales` reads no database, broker, Event Store, clock, filesystem, or
global registry. The application supplies facts and decides what to do with the
returned value.

## Other representation shapes

A Projection need not return a view struct. A primitive statistic is valid:

```go
var CountPaidOrders projection.Projection[int, OrderEvent] =
    func(total int, event OrderEvent) int {
        if event.Kind == OrderPaid {
            return total + 1
        }
        return total
    }
```

An index may use a persistent collection:

```go
type ProductIndex = immutable.Map[ProductID, ProductView]
```

Customer views, dashboards, accumulated money values, and other derived
representations follow the same contract; modeling does not define what counts
as a read model.

## Incremental and batch use

The Projection function is already the incremental operation:

```go
next := ProjectSales(current, event)
```

Project is its ordered batch form:

```go
rebuilt := projection.Project(empty, history, ProjectSales)
```

Both paths share the same `P x E -> P` semantics. Project applies events exactly
once in immutable-list encounter order and threads each result into the next
call. Empty input returns the initial value.

## Projection versus Reducer

Both types have a `value x event -> value` shape, but they communicate different
ownership:

- `reducer.Reducer[S,E]` evolves domain state from facts;
- `projection.Projection[P,E]` evolves a derived representation from facts.

The same event history can be replayed into domain state and projected into a
read model independently. The packages do not import one another.

## Ignored and unknown events

A Projection ignores an event by returning its current representation:

```go
default:
    return current
```

The application owns heterogeneous event unions and unknown-kind policy. The
package adds no error, optional value, marker interface, registry, or reflection
dispatcher for this case.

## Persistent collections

P may contain or be an `immutable.Map`, List, or Set. Returning the new value
uses those collections' structural sharing naturally. Projection does not
convert persistent collections to mutable slices or maps.

## Persistence and CQRS boundary

A typical application flow is:

```text
load facts/current view -> Projection -> save returned view
```

Only the middle transformation belongs to this package. Loading, repositories,
database writes, subscriptions, offsets, checkpoints, retries, consumer groups,
background workers, and eventual consistency remain in application or
infrastructure layers.

## Laws and cost

Project satisfies empty identity and concatenation:

```text
Project(p, empty, f) = p
Project(p, A ++ B, f) = Project(Project(p, A, f), B, f)
```

For deterministic f, repeated calls with equal inputs return equal results.
Project runs in `O(n * cost(f))`, iterates `List.Values()` directly, and does not
materialize the event list. Benchmarks report behavior rather than promise a
machine-independent threshold.

## Deliberate exclusions

There is no FallibleProjection, ContramapEvent, Compose, Zip, Map,
ContramapState, Replay alias, parallel API, ReadModel/View/Event marker,
registry, dispatcher, handler, persistence, broker consumer, Event Store, or
CQRS runtime in this package.
