package modeling

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// Decision is the successful result of a domain decision: the next state and
// the ordered domain events emitted while reaching it. Its zero value is valid.
// Immutability is shallow, following ordinary Go value semantics for S and E.
type Decision[S, E any] struct {
	state  S
	events immutable.List[E]
}

// Decide creates a Decision for state with no emitted events. E intentionally
// comes before S so callers can write Decide[MyEvent](state) and infer S.
func Decide[E, S any](state S) Decision[S, E] {
	return Decision[S, E]{state: state}
}

// State returns the next domain state.
func (d Decision[S, E]) State() S {
	return d.state
}

// Emit returns a new Decision with events appended in argument order.
func (d Decision[S, E]) Emit(events ...E) Decision[S, E] {
	d.events = d.events.Append(events...)
	return d
}

// EventCount returns the number of emitted events.
func (d Decision[S, E]) EventCount() int {
	return d.events.Len()
}

// Events returns the emitted events in order without exposing their storage.
func (d Decision[S, E]) Events() iter.Seq[E] {
	return d.events.Values()
}
