package policy

const panicNilMapper = "modeling/policy: nil mapper"

// Map transforms the output of p.
func Map[I, O, P any](p Policy[I, O], mapper func(O) P) Policy[I, P] {
	requirePolicy(p)
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input I) P { return mapper(p(input)) }
}

// Contramap adapts a larger or different input before invoking p.
func Contramap[A, I, O any](p Policy[I, O], mapper func(A) I) Policy[A, O] {
	requirePolicy(p)
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input A) O { return p(mapper(input)) }
}
