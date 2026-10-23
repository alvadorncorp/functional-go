package eq

// Equal reports whether a and b are equal without using reflect.
//
// Resolution order:
//  1. If a implements Equal(T) bool, return a.Equal(b).
//  2. Otherwise compare via interface ==. Two nil interfaces (possible only
//     when T is an interface type) are equal. When both dynamic values share
//     a comparable type, == decides. When the dynamic type is not comparable
//     (slice, map, func, …) and no Equal method is present, Equal returns
//     false — it never panics and never deep-compares.
//
// Callers that need structural equality for non-comparable types should
// implement Equal(T) bool on the type or use an EqualFunc at their package
// (for example a package-level EqualFunc).
func Equal[T any](a, b T) bool {
	type equaler interface{ Equal(T) bool }
	if ea, ok := any(a).(equaler); ok {
		return ea.Equal(b)
	}
	return equalAny(any(a), any(b))
}

// For returns a comparator with the same semantics as Equal, selecting the
// strategy once for concrete T. Interface T retains per-value Equal dispatch.
func For[T any]() func(T, T) bool {
	var zero T
	if any(zero) == nil {
		return Equal[T]
	}
	type equaler interface{ Equal(T) bool }
	if _, ok := any(zero).(equaler); ok {
		return func(a, b T) bool {
			return any(a).(equaler).Equal(b)
		}
	}
	return func(a, b T) bool {
		return equalAny(any(a), any(b))
	}
}

// EqualComparable reports whether a and b are equal under ==.
// Use this when T is known comparable and no Equal method should be consulted.
func EqualComparable[T comparable](a, b T) bool {
	return a == b
}

func equalAny(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return equalInterface(a, b)
}

// equalInterface compares two non-nil interface values with ==.
// Comparing interfaces whose dynamic type is not comparable panics in Go;
// recover that case and report false so Equal never panics.
func equalInterface(a, b any) (result bool) {
	defer func() {
		if recover() != nil {
			result = false
		}
	}()
	return a == b
}
