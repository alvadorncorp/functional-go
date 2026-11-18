# Package `primitives/function`

**Import path:** `go.alvadorn.org/functional/primitives/function`

`function` provides small, stateless generic utilities for ordinary Go function
values, including a named homogeneous transform type. It is a standard-library-only
leaf and intentionally contains no collection operations, arbitrary-arity
currying, memoization, retry policy, or higher-kinded abstraction.

| API | Contract |
|---|---|
| `Identity(value)` | returns `value` unchanged |
| `Constant(value)` | ignores its later argument and returns `value` |
| `Compose(outer, inner)` | `outer(inner(x))` |
| `Pipe(first, second)` | `second(first(x))` |
| `Flip(fn)` | reverses a binary function's arguments |
| `Tap(effect)` | executes `effect(x)` once and returns `x` unchanged |
| `Transform[T]` | named `func(T) T` (endomorphism); call as `transform(value)` |
| `PipeTransforms(ts...)` | left-to-right composition of `Transform[T]`; empty is identity |
| `ComposeTransforms(ts...)` | right-to-left composition of `Transform[T]`; empty is identity |

`Compose`/`Pipe` remain fixed arity-2 and may change intermediate types.
`PipeTransforms`/`ComposeTransforms` are specialized for homogeneous `T → T`
pipelines of arbitrary length.

Compose and Pipe obey left/right identity and associativity. Transform pipelines
obey the same laws under their respective directions. Flip is an involution. Tap
preserves its input independently of the effect. Nil callbacks are not
special-cased; invoking a returned function naturally panics if its required
callback is nil, like a direct Go function call.

See [module overview](../overview.md) and [examples](../examples.md).
