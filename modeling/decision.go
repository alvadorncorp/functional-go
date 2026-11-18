package modeling

import (
	"iter"

	internaldecision "go.alvadorn.org/functional/modeling/internal/decision"
)

// Decision is the successful result of a domain decision: the next state and
// the ordered domain events emitted while reaching it. Its zero value is valid.
// Immutability is shallow, following ordinary Go value semantics for S and E.
type Decision[S, E any] struct {
	value internaldecision.Value[S, E]
}

// Decide creates a Decision for state with no emitted events. E intentionally
// comes before S so callers can write Decide[MyEvent](state) and infer S.
func Decide[E, S any](state S) Decision[S, E] {
	return Decision[S, E]{value: internaldecision.New[E](state)}
}

// State returns the next domain state.
func (d Decision[S, E]) State() S {
	return d.value.State()
}

// Emit returns a new Decision with events appended in argument order.
func (d Decision[S, E]) Emit(events ...E) Decision[S, E] {
	d.value = d.value.Append(events...)
	return d
}

// EventCount returns the number of emitted events.
func (d Decision[S, E]) EventCount() int {
	return d.value.EventCount()
}

// Events returns the emitted events in order without exposing their storage.
func (d Decision[S, E]) Events() iter.Seq[E] {
	return d.value.Events()
}
