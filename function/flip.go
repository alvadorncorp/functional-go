package function

// Flip returns fn with its two arguments reversed.
func Flip[A, B, C any](fn func(A, B) C) func(B, A) C {
	return func(b B, a A) C { return fn(a, b) }
}
