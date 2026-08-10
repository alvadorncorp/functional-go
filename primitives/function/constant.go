package function

// Constant returns a function that ignores its argument and returns value.
func Constant[A, B any](value A) func(B) A {
	return func(B) A { return value }
}
