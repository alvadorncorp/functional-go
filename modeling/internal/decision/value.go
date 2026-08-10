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

// MapState transforms the state and reuses the persistent event list.
func MapState[S, T, E any](value Value[S, E], mapper func(S) T) Value[T, E] {
	return Value[T, E]{state: mapper(value.state), events: value.events}
}

// MapEvents transforms events in source order into new persistent storage.
func MapEvents[S, E, F any](value Value[S, E], mapper func(E) F) Value[S, F] {
	if value.events.IsEmpty() {
		return Value[S, F]{state: value.state}
	}

	var mapped immutable.List[F]
	builder := mapped.Builder()
	for event := range value.events.Values() {
		builder.Append(mapper(event))
	}
	return Value[S, F]{state: value.state, events: builder.Build()}
}

// Combine returns the right state with left events followed by right events.
func Combine[S, E any](left, right Value[S, E]) Value[S, E] {
	if left.events.IsEmpty() {
		return right
	}
	if right.events.IsEmpty() {
		return Value[S, E]{state: right.state, events: left.events}
	}

	builder := left.events.Builder()
	for event := range right.events.Values() {
		builder.Append(event)
	}
	return Value[S, E]{state: right.state, events: builder.Build()}
}
