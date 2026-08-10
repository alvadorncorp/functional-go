# Package `primitives/predicate`

**Import path:** `github.com/alvadorncorp/gopackages/functional/primitives/predicate`

`Predicate[T]` names an ordinary `func(T) bool`. `Not`, `And`, and `Or` compose
predicates with left-to-right short-circuiting. Empty `And` is true and empty
`Or` is false. `All`, `Any`, and `None` quantify over slices without allocating
intermediate collections; empty All/None are true and empty Any is false.

`EqualTo`, `NotEqualTo`, and `OneOf` are limited to comparable values. `OneOf`
uses a linear scan and introduces no map/set or immutable dependency.

The package is a standard-library-only leaf. It does not import `function` just
to express negation or composition.

See [module overview](../overview.md) and [examples](../examples.md).
