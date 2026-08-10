package modeling

import "errors"

const panicNilInvariant = "modeling: nil invariant"

// Invariant checks whether a state remains valid.
type Invariant[S any] func(S) error

// FailFast composes invariants in argument order and returns the first failure.
func FailFast[S any](invariants ...Invariant[S]) Invariant[S] {
	requireInvariants(invariants)
	return func(state S) error {
		for _, invariant := range invariants {
			if err := invariant(state); err != nil {
				return err
			}
		}
		return nil
	}
}

// Accumulate composes invariants in argument order and joins every failure.
func Accumulate[S any](invariants ...Invariant[S]) Invariant[S] {
	requireInvariants(invariants)
	return func(state S) error {
		failures := make([]error, 0, len(invariants))
		for _, invariant := range invariants {
			if err := invariant(state); err != nil {
				failures = append(failures, err)
			}
		}
		return errors.Join(failures...)
	}
}

func requireInvariants[S any](invariants []Invariant[S]) {
	for _, invariant := range invariants {
		if invariant == nil {
			panic(panicNilInvariant)
		}
	}
}
