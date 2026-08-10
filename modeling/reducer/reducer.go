// Package reducer defines pure, synchronous state evolution from ordered
// application-owned events. It provides no Event Sourcing runtime,
// persistence, registry, dispatch, serialization, or transport behavior.
package reducer

// Reducer evolves state from one event that has already occurred.
// Ordinary functions with the same signature are directly assignable.
// Reducers should be pure and deterministic; the package does not enforce it.
type Reducer[S, E any] func(S, E) S
