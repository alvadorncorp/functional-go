# Package `order`

**Import path:** `github.com/alvadorncorp/gopackages/functional/order`

`Order[T]` returns an integer whose sign means less, equivalent, or greater;
implementations need not normalize to -1/0/1. `Natural` delegates to
`cmp.Compare`. `Reverse` swaps arguments, avoiding integer-negation overflow.
`By` projects values to ordered keys, and `Then` composes orders
lexicographically without evaluating secondary after a primary decision.

`Less`, `Greater`, and `Equal` adapt an Order to ordinary binary predicates.
Order equality is equivalence under that Order, not necessarily Go `==`.
`Min` and `Max` return their first argument on an equivalent tie.

The package is a standard-library-only leaf and imports none of the other
primitive packages.

See [module overview](../overview.md) and [examples](../examples.md).
