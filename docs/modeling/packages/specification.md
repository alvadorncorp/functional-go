# Specification

**Import path:** `go.alvadorn.org/functional/modeling/specification`

`Specification[T]` is a boolean classification with explicit meaning in the
domain vocabulary. It answers: does this value belong to this category?

```go
var IsActive specification.Specification[Customer] = func(customer Customer) bool {
    return customer.Active
}

if IsActive(customer) {
    // application behavior
}
```

## Neighboring concepts

| Concept | Question | Shape |
|---|---|---|
| Predicate | generic boolean condition | `T -> bool` |
| Specification | belongs to the domain category? | `T -> bool` |
| Rule | satisfies an explainable condition? | `T -> error` |
| Requirement | is the operation permitted? | `S x I -> error` |
| Invariant | is the resulting state valid? | `S -> error` |
| Policy | which strategy/result to use? | `I -> O` |

Predicate and Specification share a representation but not a responsibility.
Specification's nominal identity communicates domain intent. Specification
returns neither error nor message; when a failure needs explaining, use
Rule/Requirement/Invariant according to the question.

## Composition

`All` ANDs left-to-right and stops at the first false. `Any` ORs left-to-right
and stops at the first true. `All()` returns true and `Any()` returns false.
`Not` creates a derived classification; it does not turn classification into
validation.

```go
eligibleForCredit := specification.All(IsActive, HasVerifiedEmail, HasNoOutstandingDebt)
inactive := specification.Not(IsActive)
specialDiscount := specification.Any(IsVIP, IsEmployee, HasPromotionCode)
```

`Contramap` adapts the input while keeping the original classification:

```go
isHighValueOrder := specification.Contramap(isHighValueMoney, func(order Order) Money {
    return order.Total
})
```

For more than one value, declare a domain context and classify that value:

```go
type TransferContext struct {
    Account Account
    Amount  Money
}

var RequiresManualReview specification.Specification[TransferContext]
```

## Specification e Policy

The application can use a Specification as a Policy condition, keeping the
adaptation explicit:

```go
discount := policy.Choose(
    func(context PricingContext) bool { return IsVIP(context) },
    vipDiscount,
    regularDiscount,
)
```

There is no Policy -> Specification dependency and no Specification -> Policy
dependency.

## Surface and nil

The surface is exactly `Specification`, `All`, `Any`, `Not`, and `Contramap`.
`Of` was rejected because direct assignment/conversion is enough; `Xor` was
rejected for lack of a concrete case. Filter, Validate, Explain, Reason, Error,
and Message belong to other responsibilities.

The zero Specification is a nil function when called. Combinators reject a nil
Specification eagerly; Contramap also rejects a nil mapper. Nil never means
true, false, or skip. The package imports only the stdlib `slices` and has no
effects, runtime, reflection, registry, persistence, or infrastructure.
