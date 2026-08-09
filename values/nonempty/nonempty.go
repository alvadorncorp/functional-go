package nonempty

const indexPanicMessage = "functional/values/nonempty: index out of range"

// NonEmpty is a collection with at least one element.
// Its zero value contains exactly zero(T).
type NonEmpty[T any] struct {
	head T
	tail []T
}

// Of constructs a NonEmpty with head followed by a defensive copy of tail.
func Of[T any](head T, tail ...T) NonEmpty[T] {
	return NonEmpty[T]{head: head, tail: append([]T(nil), tail...)}
}

// FromSlice constructs a NonEmpty from a defensive copy of values.
// It returns false when values is empty.
func FromSlice[T any](values []T) (NonEmpty[T], bool) {
	if len(values) == 0 {
		return NonEmpty[T]{}, false
	}
	return Of(values[0], values[1:]...), true
}

// Head returns the first element.
func (n NonEmpty[T]) Head() T { return n.head }

// Len returns the number of elements, which is always at least one.
func (n NonEmpty[T]) Len() int { return 1 + len(n.tail) }

// At returns the element at index.
// It panics with "functional/values/nonempty: index out of range" when index is outside bounds.
func (n NonEmpty[T]) At(index int) T {
	if index < 0 || index >= n.Len() {
		panic(indexPanicMessage)
	}
	if index == 0 {
		return n.head
	}
	return n.tail[index-1]
}

// Last returns the final element.
func (n NonEmpty[T]) Last() T {
	if len(n.tail) == 0 {
		return n.head
	}
	return n.tail[len(n.tail)-1]
}

// Tail returns a defensive copy of every element after Head.
func (n NonEmpty[T]) Tail() []T { return append([]T(nil), n.tail...) }

// Items returns a defensive copy of all elements, in order.
func (n NonEmpty[T]) Items() []T {
	items := make([]T, n.Len())
	items[0] = n.head
	copy(items[1:], n.tail)
	return items
}

// Map applies transform to each element while preserving cardinality.
func Map[T, U any](n NonEmpty[T], transform func(T) U) NonEmpty[U] {
	tail := make([]U, len(n.tail))
	for index, value := range n.tail {
		tail[index] = transform(value)
	}
	return NonEmpty[U]{head: transform(n.head), tail: tail}
}

// Append returns a new NonEmpty with values after its existing elements.
func (n NonEmpty[T]) Append(values ...T) NonEmpty[T] {
	if len(values) == 0 {
		return n
	}
	tail := make([]T, len(n.tail)+len(values))
	copy(tail, n.tail)
	copy(tail[len(n.tail):], values)
	return NonEmpty[T]{head: n.head, tail: tail}
}
