package change

// Apply applies one transformation to state using value-first argument order.
func Apply[S any](state S, transformation Change[S]) S {
	return transformation(state)
}

// ApplyAll applies changes to state from left to right.
//
// With no changes, ApplyAll returns state unchanged.
func ApplyAll[S any](state S, changes ...Change[S]) S {
	return Pipe(changes...)(state)
}
