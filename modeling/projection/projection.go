// Package projection defines pure, synchronous derivation of application-owned
// read models and other representations from facts. It provides no persistence,
// subscription, broker, Event Store, registry, or CQRS runtime behavior.
package projection

// Projection derives the next representation from one event that has already
// occurred. Ordinary functions with the same signature are directly assignable.
// Projections should be pure and deterministic; the package does not enforce it.
type Projection[P, E any] func(P, E) P
