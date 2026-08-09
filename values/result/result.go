package result

const panicMessage = "functional/values/result: Result is Err"

// Result is an explicit success-or-failure value whose zero value is Err(zero(E)).
type Result[T any, E any] struct {
	value T
	err   E
	ok    bool
}

// Ok returns a successful Result containing value.
func Ok[T, E any](value T) Result[T, E] { return Result[T, E]{value: value, ok: true} }

// Err returns a failed Result containing err.
func Err[T, E any](err E) Result[T, E] { return Result[T, E]{err: err} }

// IsOk reports whether r is successful.
func (r Result[T, E]) IsOk() bool { return r.ok }

// IsErr reports whether r is failed.
func (r Result[T, E]) IsErr() bool { return !r.ok }

// Get returns the successful value and whether r is successful.
func (r Result[T, E]) Get() (T, bool) { return r.value, r.ok }

// OrZero returns the successful value or the zero value of T.
func (r Result[T, E]) OrZero() T { return r.value }

// ErrValue returns the failure value and whether r is failed.
func (r Result[T, E]) ErrValue() (E, bool) { return r.err, !r.ok }

// OrElse returns the successful value or fallback when r is failed.
func (r Result[T, E]) OrElse(fallback T) T {
	if r.ok {
		return r.value
	}
	return fallback
}

// MustGet returns the successful value or panics when r is failed.
func (r Result[T, E]) MustGet() T {
	if !r.ok {
		panic(panicMessage)
	}
	return r.value
}

// Map transforms a successful Result or preserves a failed one.
func Map[T, U, E any](r Result[T, E], transform func(T) U) Result[U, E] {
	if !r.ok {
		return Err[U](r.err)
	}
	return Ok[U, E](transform(r.value))
}

// MapErr transforms a failed Result or preserves a successful one.
func MapErr[T, E, F any](r Result[T, E], transform func(E) F) Result[T, F] {
	if r.ok {
		return Ok[T, F](r.value)
	}
	return Err[T](transform(r.err))
}

// FlatMap transforms a successful Result or preserves a failed one.
func FlatMap[T, U, E any](r Result[T, E], transform func(T) Result[U, E]) Result[U, E] {
	if !r.ok {
		return Err[U](r.err)
	}
	return transform(r.value)
}

// Equal reports whether a and b have equal discriminators and values.
func Equal[T, E comparable](a, b Result[T, E]) bool { return a == b }
