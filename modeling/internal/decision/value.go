// Package decision implements the persistent value mechanics behind the
// public modeling Decision facade.
package decision

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// Value stores a Decision state and its ordered persistent events.
// Its zero value is valid.
type Value[S, E any] struct {
	state  S
	events immutable.List[E]
}

// New creates a Value for state with no events. E comes first so callers can
// specify the event type while S is inferred.
func New[E, S any](state S) Value[S, E] {
	return Value[S, E]{state: state}
}

// State returns the stored state.
func (v Value[S, E]) State() S {
	return v.state
}

// Append returns a new Value with events appended in argument order.
func (v Value[S, E]) Append(events ...E) Value[S, E] {
	v.events = v.events.Append(events...)
	return v
}

// EventCount returns the number of stored events.
func (v Value[S, E]) EventCount() int {
	return v.events.Len()
}

// Events returns the stored events in order.
func (v Value[S, E]) Events() iter.Seq[E] {
	return v.events.Values()
}
