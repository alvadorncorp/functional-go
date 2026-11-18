# Capabilities as structs of functions

Dependencies are functions and small structs assembled at the composition root
(`adr:workflow:functional-dependency-injection-through-capabilities`). Checkout
example: `Orders{ Load func(...) action.Action[Order] }`. Writes exist only in
`TxScope` inside `transaction.Within`.

Interfaces stay correct where there is genuine polymorphism; they are not the
mandatory injection mechanism. The names `Repository`, `Store`, `CommandBus`
are forbidden in the `workflow` core and must not be reintroduced as an
application kit.
