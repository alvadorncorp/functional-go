# Package `structural/lazy`

**Import path:** `github.com/alvadorncorp/gopackages/functional/structural/lazy`

`Lazy[T]` is a pointer handle for a deferred, memoized computation built on
`sync.OnceValue`. New rejects a nil initializer immediately. Get runs the
initializer at most once; concurrent callers wait for that evaluation and
observe the same value. If initialization panics, Get memorizes and re-panics
the same payload on every call. The stored initializer is released after the
evaluation; the cached value or panic remains for the Lazy lifetime.

`runtime.Goexit` inside the initializer is not a successful evaluation. The
Lazy never caches the zero value for it: the Get that ran the initializer and
every later Get panic. Under Go's default panic(nil) conversion (main module
`go` >= 1.21 with `GODEBUG` `panicnil` unset) the payload is
`*runtime.PanicNilError`, which the first caller can recover. Under the legacy
`panicnil=1` override the native `sync.OnceValue` follows the runtime: it still
never returns the zero value as an initialized success, but the payload type is
not guaranteed. (The earlier hand-written implementation returned the zero
value silently after Goexit.) This guide does not change the default handling
of `panic(nil)`; that is plain Go behavior.

The zero Lazy and nil *Lazy are invalid. Of wraps an existing value. Map and
FlatMap are free functions because they change type parameters; construction is
fully lazy and the returned Lazy memoizes the whole chain.

There is no Value alias, IsEvaluated, Reset, TTL, Thunk, Task, Future, Promise,
IO, cancellation, or retry. A blocked initializer blocks all Get callers. The
package is a stdlib-only leaf. See [module overview](../overview.md) and
[examples](../examples.md).
