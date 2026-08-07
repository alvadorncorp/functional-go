// Package ownership provides the exclusive-mutation token used by the
// immutability core.
//
// An Owner is a pointer-identity token with an Active state. Two values are
// the same owner only when they are the same pointer and that token is still
// Active. Invalidate permanently deactivates a token (Freeze); nodes may keep
// the stamp, but IsOwnedBy becomes false so later mutation windows copy.
//
// A nil stamp on a node means the node is published/shared (born published)
// and must not be mutated in place.
//
// This package answers only "who may mutate?". How a structure copies itself
// belongs to the concrete collection/engine, not to ownership.
package ownership
