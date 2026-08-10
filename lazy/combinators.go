package lazy

func Map[A, B any](source *Lazy[A], fn func(A) B) *Lazy[B] {
	return New(func() B { return fn(source.Get()) })
}

func FlatMap[A, B any](source *Lazy[A], fn func(A) *Lazy[B]) *Lazy[B] {
	return New(func() B { return fn(source.Get()).Get() })
}
