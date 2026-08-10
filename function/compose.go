package function

// Compose returns a function that applies inner and then outer.
func Compose[A, B, C any](outer func(B) C, inner func(A) B) func(A) C {
	return func(value A) C { return outer(inner(value)) }
}

// Pipe returns a function that applies first and then second.
func Pipe[A, B, C any](first func(A) B, second func(B) C) func(A) C {
	return func(value A) C { return second(first(value)) }
}
