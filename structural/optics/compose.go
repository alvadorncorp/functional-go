package optics

// Compose builds a Lens from S to B by focusing through outer then inner.
// Get reaches the nested focus; Set rebuilds outer values through both setters.
func Compose[S, A, B any](outer Lens[S, A], inner Lens[A, B]) Lens[S, B] {
	return NewLens(
		func(source S) B { return inner.Get(outer.Get(source)) },
		func(source S, value B) S {
			focused := outer.Get(source)
			return outer.Set(source, inner.Set(focused, value))
		},
	)
}
