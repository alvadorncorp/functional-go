package optional

const panicMessage = "functional/values/optional: Optional is None"

// Optional is an explicit Some-or-None value type whose zero value is None.
type Optional[T any] struct {
	value T
	ok    bool
}

// Some returns an Optional containing value.
func Some[T any](value T) Optional[T] { return Optional[T]{value: value, ok: true} }

// None returns an Optional without a value.
func None[T any]() Optional[T] { return Optional[T]{} }

// From returns Some(value) when ok is true; otherwise it returns None.
func From[T any](value T, ok bool) Optional[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// IsSome reports whether o contains a value.
func (o Optional[T]) IsSome() bool { return o.ok }

// IsNone reports whether o does not contain a value.
func (o Optional[T]) IsNone() bool { return !o.ok }

// Get returns the value and whether o contains one.
func (o Optional[T]) Get() (T, bool) { return o.value, o.ok }

// OrZero returns the contained value or the zero value of T.
func (o Optional[T]) OrZero() T { return o.value }

// OrElse returns the contained value or fallback when o is None.
func (o Optional[T]) OrElse(fallback T) T {
	if o.ok {
		return o.value
	}
	return fallback
}

// MustGet returns the contained value or panics when o is None.
func (o Optional[T]) MustGet() T {
	if !o.ok {
		panic(panicMessage)
	}
	return o.value
}

// Filter returns o when its value satisfies predicate; otherwise it returns None.
func (o Optional[T]) Filter(predicate func(T) bool) Optional[T] {
	if !o.ok || !predicate(o.value) {
		return None[T]()
	}
	return o
}

// Inspect calls inspect with the contained value and returns o unchanged.
func (o Optional[T]) Inspect(inspect func(T)) Optional[T] {
	if o.ok {
		inspect(o.value)
	}
	return o
}

// Map transforms the contained value, or returns None when o is None.
func Map[T, U any](o Optional[T], transform func(T) U) Optional[U] {
	if !o.ok {
		return None[U]()
	}
	return Some(transform(o.value))
}

// FlatMap transforms the contained value into an Optional, or returns None when o is None.
func FlatMap[T, U any](o Optional[T], transform func(T) Optional[U]) Optional[U] {
	if !o.ok {
		return None[U]()
	}
	return transform(o.value)
}

// Match calls noneFn for None, including the zero Optional, or someFn for Some and returns its result.
func Match[T, R any](o Optional[T], noneFn func() R, someFn func(T) R) R {
	if o.ok {
		return someFn(o.value)
	}
	return noneFn()
}

// Equal reports whether a and b have equal discriminators and values.
func Equal[T comparable](a, b Optional[T]) bool { return a == b }
