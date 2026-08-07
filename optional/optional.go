package optional

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// panicMessage is what MustGet panics with when the Optional is absent.
const panicMessage = "functional/optional: Optional has no value"

// Optional holds a value of type T that may or may not be present.
//
// The zero Optional is a valid None, so a struct field of type Optional[T]
// does not need a constructor or a nil check. Some(zero) and None are
// distinct: presence is tracked apart from the value.
//
// Optional guards T inline, not behind a pointer. Consequences, all
// deliberate:
//   - Optional[T] is ==-comparable exactly when T is comparable;
//   - constructing a Some never allocates;
//   - the held value is a copy, it cannot be mutated through an external
//     alias;
//   - immutability is SHALLOW: Optional[[]byte] shares the slice's backing
//     array. For an element that must not be mutable, compose with an
//     immutable type (e.g. functional/persistent.List) as T.
//
// The internal representation is frozen at exactly {value T; present bool}.
// Adding any non-comparable field would silently remove ==-comparability
// from every instantiation — this is tested.
type Optional[T any] struct {
	value   T
	present bool
}

// Some returns an Optional holding value, marked present.
//
// Some(nil) for a pointer/map/slice/channel/func/interface T is present and
// holds nil. That is intentional and is NOT the same as None. Use
// FromPointer to convert a possibly-nil pointer into presence.
func Some[T any](value T) Optional[T] {
	return Optional[T]{value: value, present: true}
}

// None returns an absent Optional[T].
func None[T any]() Optional[T] {
	return Optional[T]{}
}

// FromPointer returns Some(*p) when p is not nil and None otherwise. The
// value is copied: later writes through p do not reach the Optional.
func FromPointer[T any](p *T) Optional[T] {
	if p == nil {
		return None[T]()
	}
	return Some(*p)
}

// FromOK adapts any comma-ok result from Go. Because Go expands a
// multi-valued call into a full argument list, it composes directly:
//
//	optional.FromOK(os.LookupEnv("HOME"))
func FromOK[T any](value T, ok bool) Optional[T] {
	if !ok {
		return None[T]()
	}
	return Some(value)
}

// IsSome reports whether o holds a value.
func (o Optional[T]) IsSome() bool { return o.present }

// IsNone reports whether o holds no value.
func (o Optional[T]) IsNone() bool { return !o.present }

// IsZero reports whether o holds no value. It is the same as !IsSome and
// enables the stdlib omitzero protocol for whatever composes Optional[T] as
// a field.
func (o Optional[T]) IsZero() bool { return !o.present }

// Get returns the held value and true, or the zero value of T and false.
func (o Optional[T]) Get() (T, bool) { return o.value, o.present }

// MustGet returns the held value, or panics with
// "functional/optional: Optional has no value" when absent. It exists for
// tests and package initialization, not for the normal production path.
func (o Optional[T]) MustGet() T {
	if !o.present {
		panic(panicMessage)
	}
	return o.value
}

// OrElse returns the held value, or fallback when absent.
func (o Optional[T]) OrElse(fallback T) T {
	if o.present {
		return o.value
	}
	return fallback
}

// OrElseGet returns the held value, calling fallback only when absent. Use
// it when producing the fallback is expensive.
func (o Optional[T]) OrElseGet(fallback func() T) T {
	if o.present {
		return o.value
	}
	return fallback()
}

// OrZero returns the held value, or the zero value of T when absent.
func (o Optional[T]) OrZero() T { return o.value }

// Or returns o when present, and other otherwise, for a left-to-right
// precedence chain: flagValue.Or(envValue).Or(defaultValue)
func (o Optional[T]) Or(other Optional[T]) Optional[T] {
	if o.present {
		return o
	}
	return other
}

// Filter returns o when present and keep reports true; None otherwise. keep
// is NEVER called on an absent Optional.
func (o Optional[T]) Filter(keep func(T) bool) Optional[T] {
	if !o.present || !keep(o.value) {
		return None[T]()
	}
	return o
}

// Pointer returns a pointer to a COPY of the value, or nil when absent. The
// caller owns the copy; writing through it does not reach the Optional.
func (o Optional[T]) Pointer() *T {
	if !o.present {
		return nil
	}
	v := o.value
	return &v
}

// String renders "Some(<value>)" or "None" via fmt.Sprintf("%v", ...) on the
// value. It exists for readable test failures and %v — it is NOT a
// serialization format.
func (o Optional[T]) String() string {
	if !o.present {
		return "None"
	}
	return fmt.Sprintf("Some(%v)", o.value)
}

// Equal is a METHOD, not a free function: go-cmp only consults a method
// named Equal, and without it cmp.Diff would PANIC on a struct with
// unexported fields. It delegates to functional/internal/eq.Equal.
func (o Optional[T]) Equal(other Optional[T]) bool {
	if o.present != other.present {
		return false
	}
	if !o.present {
		return true
	}
	return eq.Equal(o.value, other.value)
}
