// Package specification provides pure domain classifications.
//
// A Specification names a Boolean condition in the application's domain
// vocabulary. It classifies a value; it does not validate, explain failure,
// select a strategy, execute effects, or provide a runtime.
package specification

const panicNilSpecification = "modeling/specification: nil specification"

// Specification reports whether a value belongs to a domain category.
type Specification[T any] func(T) bool

func requireSpecifications[T any](specifications []Specification[T]) {
	for _, spec := range specifications {
		if spec == nil {
			panic(panicNilSpecification)
		}
	}
}
