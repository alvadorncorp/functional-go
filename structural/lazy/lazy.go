package lazy

import "sync"

const nilInitializerPanic = "functional/lazy: nil initializer"

// Lazy is a concurrent-safe deferred, memoized computation of T.
// It evaluates at most once and memoizes either its value or panic.
// Construct a Lazy with New or Of; a zero Lazy or nil *Lazy is invalid.
type Lazy[T any] struct {
	get func() T
}

// New returns a Lazy that will evaluate initializer at most once on Get.
// New panics if initializer is nil.
func New[T any](initializer func() T) *Lazy[T] {
	if initializer == nil {
		panic(nilInitializerPanic)
	}
	return &Lazy[T]{get: sync.OnceValue(initializer)}
}

// Of returns a Lazy that yields value without further computation.
func Of[T any](value T) *Lazy[T] {
	return New(func() T { return value })
}

// Get evaluates the initializer at most once and returns the memoized value.
// If the initializer panics, Get re-panics with the same payload on every call.
// If the initializer calls runtime.Goexit, then under Go's default panic(nil)
// conversion (main module go >= 1.21, GODEBUG panicnil unset) Get panics with
// *runtime.PanicNilError and subsequent calls also panic with that type. Under
// the legacy panicnil=1 override the payload follows the runtime and is not
// guaranteed to be that type, but Get never returns the zero value as success.
func (l *Lazy[T]) Get() T {
	return l.get()
}
