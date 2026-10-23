// Package eq resolves value equality for type parameters that carry no
// comparability constraint. It is shared by every functional subpackage that
// needs an Equal method on a generic type, so all of them agree on the answer
// for the same pair of values.
//
// Equal never uses reflect, unsafe, or reflect.DeepEqual. It prefers an
// Equal(T) bool method when present, then interface == for comparable dynamic
// values, and returns false for non-comparable types that lack Equal (no
// deep-compare fallback). EqualComparable[T comparable] exposes plain == for
// callers that already constrain T.
//
// For[T] selects a reusable comparator once for concrete T. Interface type
// parameters keep Equal's per-value dispatch, including nil and mixed dynamic
// types. Concrete Equal methods may still require boxing their receiver.
package eq
