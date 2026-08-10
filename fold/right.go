package fold

// FoldRight accumulates values from right to left, beginning with initial.
// It uses an iterative reverse loop and does not grow the call stack.
func FoldRight[A, B any](values []A, initial B, fn func(A, B) B) B {
	state := initial
	for index := len(values) - 1; index >= 0; index-- {
		state = fn(values[index], state)
	}
	return state
}
