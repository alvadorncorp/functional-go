# Policy

**Import path:** `go.alvadorn.org/functional/modeling/policy`

`Policy[I,O]` is a pure strategy that derives a result or policy choice from
explicit context. It answers which result or strategy to use; it does not
validate, run effects, fetch data, or drive workflow.

## Choosing the operation

| Need | Operation |
|---|---|
| constant result | `Constant` |
| transform output | `Map` |
| adapt input | `Contramap` |
| chain `A -> B -> C` | `Compose` |
| compute two independent decisions | `Zip` |
| choose between two total strategies | `Choose` |
| make a policy conditional | `When` |
| try a binary alternative | `OrElse` |
| first among N alternatives | `FirstApplicable` |
| supply a default | `Fallback` |

## Totality and inapplicability

Use `Choose` when exactly one branch must produce `O`. The predicate is
evaluated once and only the chosen branch runs. Use `When` when the policy may
not apply; it produces an `Optional[O]`. `OrElse` combines two conditional
forms, `FirstApplicable` combines N, and `Fallback` turns the selection into a
total Policy.

```go
vip := policy.When(isVIP, vipDiscount)
employee := policy.When(isEmployee, employeeDiscount)
volume := policy.When(isHighValue, volumeDiscount)

binary := policy.OrElse(vip, employee)
many := policy.FirstApplicable(vip, employee, volume)
discount := policy.Fallback(many, defaultDiscount)
_, _ = binary, discount
```

## Composition

`Compose(first, second)` runs first and passes its output to second.

```go
creditPolicy := policy.Compose(customerTierPolicy, creditLimitPolicy)
```

`Zip` evaluates left, then right, then combine over the same input and produces
the domain type the application chooses, with no public tuple.

```go
pricing := policy.Zip(discountPolicy, shippingPolicy, func(discount Discount, shipping ShippingMethod) PricingDecision {
    return PricingDecision{Discount: discount, Shipping: shipping}
})
```

## Policy versus the other concepts

| Concept | Question | Shape |
|---|---|---|
| Policy | which strategy/result to use? | `I -> O` |
| Specification | belongs to a domain category? | `T -> bool` |
| Rule | satisfies an explainable condition? | `T -> error` |
| Requirement | is the operation permitted? | `S x I -> error` |
| Invariant | is the resulting state valid? | `S -> error` |
| Change | how to transform the same state? | `S -> S` |

A Specification can feed Choose, but the application does the adaptation
explicitly. Policy and Specification do not import each other.

## Purity, nil, and zero values

Policies receive all data through their input; database, HTTP, broker, clock,
filesystem, and context belong to workflow/adapters. A zero output is a valid
result. The zero Policy is a nil function when invoked directly. Combinators
reject a nil Policy, mapper, predicate, conditional, or combiner at
construction with the panics documented in the TDD.

Policy has no FlatMap, engine, registry, dispatcher, methods, marker types,
services, effects, or automatic integration. When the current vocabulary does
not express a case, the application can declare its own Policy over an explicit
input struct.
