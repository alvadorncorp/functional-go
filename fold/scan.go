package fold

// ScanLeft returns every left-fold state, including initial at index zero.
// The result length is always len(values)+1.
func ScanLeft[A, B any](values []A, initial B, fn func(B, A) B) []B {
	states := make([]B, len(values)+1)
	states[0] = initial
	for index, value := range values {
		states[index+1] = fn(states[index], value)
	}
	return states
}
