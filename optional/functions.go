package optional

// Map applies f to the held value and returns Some(f(value)); returns None
// when o is absent. f is never called for an absent Optional.
func Map[T, U any](o Optional[T], f func(T) U) Optional[U] {
	if !o.present {
		return None[U]()
	}
	return Some(f(o.value))
}

// FlatMap applies f to the held value and returns its result; returns None
// when o is absent. f is never called for an absent Optional.
func FlatMap[T, U any](o Optional[T], f func(T) Optional[U]) Optional[U] {
	if !o.present {
		return None[U]()
	}
	return f(o.value)
}

// MapErr applies a fallible conversion. An absent Optional returns (None,
// nil) without calling f; f failing returns (None, err).
//
//	port, err := optional.MapErr(rawPort, strconv.Atoi)
func MapErr[T, U any](o Optional[T], f func(T) (U, error)) (Optional[U], error) {
	if !o.present {
		return None[U](), nil
	}
	v, err := f(o.value)
	if err != nil {
		return None[U](), err
	}
	return Some(v), nil
}

// Zip combines two Optionals with combine, returning None if either is
// absent. combine is NEVER called unless both are present. It takes combine
// so callers don't need an exported Pair[A, B] type.
func Zip[A, B, R any](a Optional[A], b Optional[B], combine func(A, B) R) Optional[R] {
	if !a.present || !b.present {
		return None[R]()
	}
	return Some(combine(a.value, b.value))
}

// EqualFunc reports whether a and b hold values considered equal by eq, or
// are both absent. eq is never called unless both are present.
func EqualFunc[T, U any](a Optional[T], b Optional[U], eq func(T, U) bool) bool {
	if a.present != b.present {
		return false
	}
	if !a.present {
		return true
	}
	return eq(a.value, b.value)
}
