package fold

// FoldLeft accumulates values from left to right, beginning with initial.
func FoldLeft[A, B any](values []A, initial B, fn func(B, A) B) B {
	state := initial
	for _, value := range values {
		state = fn(state, value)
	}
	return state
}
