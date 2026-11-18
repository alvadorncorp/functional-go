// Package change models pure state evolution as first-class values.
//
// A Change[S] is an endomorphism S -> S: it accepts and returns the same state
// type. Changes are conceptually immutable, but Go cannot enforce deep
// immutability for slices, maps, pointers, or other shared references. Change
// implementations remain responsible for avoiding unintended shared mutation.
//
// Identity represents no change. A nil Change is not a no-change sentinel and
// retains ordinary Go nil-function behavior when invoked. The package adapts
// functional composition internally without exposing functional types.
package change

import "go.alvadorn.org/functional/primitives/function"

// Change transforms state while preserving its type.
type Change[S any] func(S) S

// Identity returns a Change that leaves state unchanged.
func Identity[S any]() Change[S] {
	return Change[S](function.Identity[S])
}
