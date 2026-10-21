package validated

import "github.com/alvadorncorp/gopackages/functional/values/nonempty"

// Validated holds either a valid value or one or more validation errors.
// Unlike Result's fail-fast sequencing, Validated accumulates errors with Map2.
// It intentionally has no FlatMap, whose sequential dependency would make
// accumulation semantics ambiguous.
type Validated[T any, E any] struct {
	value  T
	errors nonempty.NonEmpty[E]
	valid  bool
}

// Valid constructs a Validated containing value.
func Valid[T, E any](value T) Validated[T, E] { return Validated[T, E]{value: value, valid: true} }

// Invalid constructs an invalid Validated containing head and tail errors.
func Invalid[T, E any](head E, tail ...E) Validated[T, E] {
	return InvalidNonEmpty[T](nonempty.Of(head, tail...))
}

// InvalidNonEmpty constructs an invalid Validated containing errors.
func InvalidNonEmpty[T, E any](errors nonempty.NonEmpty[E]) Validated[T, E] {
	return Validated[T, E]{errors: errors}
}

// IsValid reports whether v holds a value.
func (v Validated[T, E]) IsValid() bool { return v.valid }

// IsInvalid reports whether v holds validation errors.
func (v Validated[T, E]) IsInvalid() bool { return !v.valid }

// Get returns v's value and whether it is valid.
func (v Validated[T, E]) Get() (T, bool) { return v.value, v.valid }

// Errors returns v's validation errors and whether it is invalid.
func (v Validated[T, E]) Errors() (nonempty.NonEmpty[E], bool) {
	return v.errors, !v.valid
}

// Map transforms a valid value. It does not call transform for an invalid value.
func Map[T, U, E any](v Validated[T, E], transform func(T) U) Validated[U, E] {
	if !v.valid {
		return InvalidNonEmpty[U](v.errors)
	}
	return Valid[U, E](transform(v.value))
}

// MapErrors transforms validation errors. It does not call transform for a valid value.
func MapErrors[T, E, F any](v Validated[T, E], transform func(E) F) Validated[T, F] {
	if v.valid {
		return Valid[T, F](v.value)
	}
	return InvalidNonEmpty[T](nonempty.Map(v.errors, transform))
}

// Map2 combines two valid values. Unlike Result's fail-fast sequencing, it
// accumulates every left error followed by every right error when both are invalid.
func Map2[A, B, R, E any](a Validated[A, E], b Validated[B, E], combine func(A, B) R) Validated[R, E] {
	switch {
	case a.valid && b.valid:
		return Valid[R, E](combine(a.value, b.value))
	case !a.valid && !b.valid:
		return InvalidNonEmpty[R](a.errors.Append(b.errors.Items()...))
	case !a.valid:
		return InvalidNonEmpty[R](a.errors)
	default:
		return InvalidNonEmpty[R](b.errors)
	}
}

// Match calls invalidFn with the errors of an invalid value, including the zero
// Validated, or validFn with a valid value and returns its result.
func Match[T, E, R any](v Validated[T, E], invalidFn func(nonempty.NonEmpty[E]) R, validFn func(T) R) R {
	if v.valid {
		return validFn(v.value)
	}
	return invalidFn(v.errors)
}
