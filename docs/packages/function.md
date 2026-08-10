# Package `function`

**Import path:** `github.com/alvadorncorp/gopackages/functional/function`

`function` provides six stateless generic combinators for ordinary Go function
values. It is a standard-library-only leaf and intentionally contains no
collection operations, arbitrary-arity currying, memoization, retry policy, or
higher-kinded abstraction.

| API | Contract |
|---|---|
| `Identity(value)` | returns `value` unchanged |
| `Constant(value)` | ignores its later argument and returns `value` |
| `Compose(outer, inner)` | `outer(inner(x))` |
| `Pipe(first, second)` | `second(first(x))` |
| `Flip(fn)` | reverses a binary function's arguments |
| `Tap(effect)` | executes `effect(x)` once and returns `x` unchanged |

Compose and Pipe obey left/right identity and associativity. Flip is an
involution. Tap preserves its input independently of the effect. Nil callbacks
are not special-cased; invoking a returned function naturally panics if its
required callback is nil, like a direct Go function call.

See [module overview](../overview.md) and [examples](../examples.md).
