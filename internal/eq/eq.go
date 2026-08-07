package eq

import "reflect"

// Equal reports whether a and b are equal, resolved in three tiers:
//  1. if the dynamic value has a method Equal(T) bool, its result;
//  2. if the dynamic values are comparable, == semantics;
//  3. otherwise reflect.DeepEqual.
//
// Equal never panics, including when the dynamic type is a slice, map, or
// func, and treats two nil interfaces as equal (reflect.DeepEqual(nil, nil)
// alone would report false).
func Equal[T any](a, b T) bool {
	if eqA, ok := any(a).(interface{ Equal(T) bool }); ok {
		return eqA.Equal(b)
	}

	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch {
	case !av.IsValid() && !bv.IsValid():
		// Both are nil interfaces (only possible when T is itself an
		// interface type, e.g. Equal[any](nil, nil)).
		return true
	case av.IsValid() != bv.IsValid():
		return false
	case av.Type() == bv.Type() && av.Type().Comparable():
		return av.Equal(bv)
	default:
		return reflect.DeepEqual(a, b)
	}
}
