package specification

const panicNilMapper = "modeling/specification: nil mapper"

// Contramap adapts a larger or different input before invoking spec.
func Contramap[A, B any](spec Specification[B], mapper func(A) B) Specification[A] {
	requireSpecifications([]Specification[B]{spec})
	if mapper == nil {
		panic(panicNilMapper)
	}
	return func(input A) bool { return spec(mapper(input)) }
}
