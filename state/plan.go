package state

// planNode is one reverse-linked immutable plan step. head points at the
// newest mutation; older walks toward the base. Apply allocates a single
// node and never runs mut.
type planNode[T any] struct {
	older *planNode[T]
	mut   Mutation[T]
}
