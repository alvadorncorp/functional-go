# Package `lazy`

**Import path:** `github.com/alvadorncorp/gopackages/functional/lazy`

`Lazy[T]` is a pointer handle for a deferred, memoized computation. New rejects
a nil initializer immediately. Get runs the initializer at most once; concurrent
callers wait for that evaluation and observe the same value. If initialization
panics, Get memorizes and re-panics the same payload on every call. The stored
initializer is cleared for capture collection after success or panic; the cached
value or panic remains for the Lazy lifetime.

The zero Lazy and nil *Lazy are invalid. Of wraps an existing value. Map and
FlatMap are free functions because they change type parameters; construction is
fully lazy and the returned Lazy memoizes the whole chain.

There is no Value alias, IsEvaluated, Reset, TTL, Thunk, Task, Future, Promise,
IO, cancellation, or retry. A blocked initializer blocks all Get callers. The
package is a stdlib-only leaf. See [module overview](../overview.md) and
[examples](../examples.md).
