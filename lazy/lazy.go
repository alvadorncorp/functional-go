package lazy

import "sync"

const nilInitializerPanic = "functional/lazy: nil initializer"

// Lazy is a concurrent-safe deferred, memoized computation of T.
// It evaluates at most once and memoizes either its value or panic.
type Lazy[T any] struct {
	once        sync.Once
	initializer func() T
	value       T
	panicValue  any
	panicked    bool
}

// New returns a Lazy that will evaluate initializer at most once on Get.
// New panics if initializer is nil.
func New[T any](initializer func() T) *Lazy[T] {
	if initializer == nil {
		panic(nilInitializerPanic)
	}
	return &Lazy[T]{initializer: initializer}
}

// Of returns a Lazy that yields value without further computation.
func Of[T any](value T) *Lazy[T] {
	return New(func() T { return value })
}

// Get evaluates the initializer at most once and returns the memoized value.
// If the initializer panics, Get re-panics with the same payload on every call.
func (l *Lazy[T]) Get() T {
	l.once.Do(func() {
		initializer := l.initializer
		l.initializer = nil
		defer func() {
			if recovered := recover(); recovered != nil {
				l.panicValue = recovered
				l.panicked = true
			}
		}()
		l.value = initializer()
	})
	if l.panicked {
		panic(l.panicValue)
	}
	return l.value
}
