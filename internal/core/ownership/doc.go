// Package ownership provides the exclusive-mutation token used by the
// immutability core.
//
// An Owner is a pure identity: two values are the same owner only when they
// are the same pointer. Nodes in persistent structures stamp *Owner; a nil
// stamp means the node is published/shared and must not be mutated in place.
//
// This package answers only "who may mutate?". How a structure copies itself
// belongs to the concrete collection/engine, not to ownership.
package ownership
