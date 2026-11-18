# Package `primitives/predicate`

**Import path:** `go.alvadorn.org/functional/primitives/predicate`

`Predicate[T]` names an ordinary `func(T) bool`. `Not`, `And`, and `Or` compose
predicates with left-to-right short-circuiting. Empty `And` is true and empty
`Or` is false.

`EqualTo`, `NotEqualTo`, and `OneOf` are limited to comparable values. `OneOf`
uses a linear scan and introduces no map/set or immutable dependency.

## Deprecated slice quantifiers (kept in v1)

`All`, `Any`, and `None` quantify over slices without allocating intermediate
collections (empty All/None are true, empty Any is false). They are marked
`Deprecated:` because each is a one-line spelling of `slices.ContainsFunc`;
behavior is unchanged and the names stay exported:

| Deprecated | Use |
|---|---|
| `predicate.Any(values, p)` | `slices.ContainsFunc(values, p)` |
| `predicate.None(values, p)` | `!slices.ContainsFunc(values, p)` |
| `predicate.All(values, p)` | `!slices.ContainsFunc(values, func(v T) bool { return !p(v) })` |

When the data is already an `iter.Seq`, `structural/iterator.Any` and
`iterator.All` quantify it directly. `predicate` stays a standard-library-only
leaf and does not import `iterator`. See
[ADR-014](../adr/functional/adr-014-v1-surface-hygiene.md).

The package is a standard-library-only leaf. It does not import `function` just
to express negation or composition.

See [module overview](../overview.md) and [examples](../examples.md).
