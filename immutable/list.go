package immutable

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/vector"
	"github.com/alvadorncorp/gopackages/functional/internal/eq"
)

// List is an immutable, ordered collection of T. Every write method returns
// a new List; the receiver is never changed. See the package doc for the
// full contract, including why List is not ==-comparable and why documented
// complexity is an upper bound rather than today's constant.
//
// Representation: List is a small handle to an unexported *vector.Root.
// A nil handle is a valid empty list (same as the zero List): Len is 0, At
// panics, and Append starts a lineage from Empty. Constructed Lists
// (ListOf, CollectList, Builder.Build, and every non-empty Append/Set/Delete
// result) hold a pointer to a published root; empty results collapse back to
// the nil handle so zero/empty paths stay allocation-free.
//
// The blank [0]func() field keeps List deliberately not ==-comparable,
// matching Map/Set: pointer identity of the root must never be observable
// via ==. Use Equal instead.
type List[T any] struct {
	_    [0]func()
	root *vector.Root[T]
}

// rootValue returns the published vector root, treating a nil handle as Empty.
func (l List[T]) rootValue() vector.Root[T] {
	if l.root == nil {
		return vector.Empty[T]()
	}
	return *l.root
}

// listFromRoot returns a List handle for root. Empty roots collapse to the
// nil handle so the zero List and empty results stay representation-equal
// and allocation-free.
func listFromRoot[T any](root vector.Root[T]) List[T] {
	if root.Len() == 0 {
		return List[T]{}
	}
	return List[T]{root: &root}
}

// ListOf returns a List holding a copy of items, in order. The caller's
// slice is never retained: later mutation of items through the caller's own
// reference does not reach the returned List.
func ListOf[T any](items ...T) List[T] {
	return (List[T]{}).Append(items...)
}

// CollectList drains seq into a new List, in iteration order.
func CollectList[T any](seq iter.Seq[T]) List[T] {
	var list List[T]
	var b *ListBuilder[T]
	for v := range seq {
		switch {
		case b != nil:
			b.Append(v)
		case list.IsEmpty():
			list = list.Append(v)
		default:
			b = list.Builder().Append(v)
		}
	}
	if b == nil {
		return list
	}
	return b.Build()
}

// Len returns the number of elements in l.
func (l List[T]) Len() int { return l.rootValue().Len() }

// IsEmpty reports whether l holds no elements. The zero List is empty.
func (l List[T]) IsEmpty() bool { return l.Len() == 0 }

// At returns the element at index i. It panics with
// "immutable: index out of range" when i is outside [0, Len).
//
// Bounds are checked here so a vector engine panic never leaks a "vector:"
// message to public callers.
func (l List[T]) At(i int) T {
	root := l.rootValue()
	if i < 0 || i >= root.Len() {
		panic(panicIndexOutOfRange)
	}
	return root.At(i)
}

// Append returns a new List with items added to the end, in order. l is
// unchanged.
//
// Documented complexity is O(log n) amortized (structure-sharing vector).
// Calling Append n times in a loop still allocates n intermediate Lists;
// use Builder for bulk construction.
func (l List[T]) Append(items ...T) List[T] {
	if len(items) == 0 {
		return l
	}
	if len(items) == 1 {
		return listFromRoot(l.rootValue().Append(items[0]))
	}
	return l.Builder().Append(items...).Build()
}

// Set returns a new List with the element at index i replaced by v. l is
// unchanged. It panics with "immutable: index out of range" when i is
// outside [0, Len).
//
// Bounds are checked here so a vector engine panic never leaks a "vector:"
// message to public callers.
func (l List[T]) Set(i int, v T) List[T] {
	root := l.rootValue()
	if i < 0 || i >= root.Len() {
		panic(panicIndexOutOfRange)
	}
	return listFromRoot(root.Set(i, v))
}

// Delete returns a new List without the element at index i; every later
// element shifts down by one. l is unchanged. It panics with
// "immutable: index out of range" when i is outside [0, Len).
//
// Delete is O(n) in general: the vector engine rebuilds by iteration, except
// that deleting the last element uses its tail-pop path. Bounds are checked
// here so a vector engine panic never leaks a "vector:" message to callers.
func (l List[T]) Delete(i int) List[T] {
	root := l.rootValue()
	if i < 0 || i >= root.Len() {
		panic(panicIndexOutOfRange)
	}
	return listFromRoot(root.Delete(i))
}

// Values returns an iterator over l's elements in index order.
func (l List[T]) Values() iter.Seq[T] {
	root := l.rootValue()
	return func(yield func(T) bool) {
		for _, v := range root.All() {
			if !yield(v) {
				return
			}
		}
	}
}

// All returns an iterator over l's (index, element) pairs in index order.
func (l List[T]) All() iter.Seq2[int, T] {
	return l.rootValue().All()
}

// Items returns a new slice holding a copy of l's elements, in order. The
// caller owns the result: mutating it never reaches l, and later writes to l
// never reach a slice returned earlier.
func (l List[T]) Items() []T {
	root := l.rootValue()
	n := root.Len()
	if n == 0 {
		return nil
	}
	items := make([]T, 0, n)
	for _, v := range root.All() {
		items = append(items, v)
	}
	return items
}

// Equal reports whether l and other hold the same elements in the same
// order, using functional/internal/eq's value equality. A shared root is
// always equal to itself, even with NaN or non-reflexive Equal methods.
func (l List[T]) Equal(other List[T]) bool {
	if l.root == other.root {
		return true
	}
	root := l.rootValue()
	otherRoot := other.rootValue()
	if root.Len() != otherRoot.Len() {
		return false
	}
	equal := eq.For[T]()
	for i, v := range root.All() {
		if !equal(v, otherRoot.At(i)) {
			return false
		}
	}
	return true
}

// EqualFunc reports whether l and other hold the same number of elements and
// equal returns true for every pair of elements at the same index, in index
// order. Comparison stops at the first false result from equal.
func (l List[T]) EqualFunc(other List[T], equal func(T, T) bool) bool {
	root := l.rootValue()
	otherRoot := other.rootValue()
	if root.Len() != otherRoot.Len() {
		return false
	}
	for i, v := range root.All() {
		if !equal(v, otherRoot.At(i)) {
			return false
		}
	}
	return true
}

// Builder returns a ListBuilder seeded from l's root (same node graph; later
// exclusive edits copy-on-write). l is not affected by any later operation
// on the builder.
func (l List[T]) Builder() *ListBuilder[T] {
	return &ListBuilder[T]{
		root: l.rootValue(),
		win:  transient.Open(),
	}
}

// ListBuilder accumulates writes for List in amortized linear time via the
// vector transient path. It is MUTABLE and is NOT safe for concurrent use.
//
// The builder holds a vector.Root and an open transient.Window.
// Append/Set/Delete call engine transient ops under win.Owner(). Build
// freezes the window (O(1) owner invalidation — no tree walk), returns a
// List with the current root, and permanently disables the builder.
//
// After Build runs, every method on the builder — including Build itself —
// panics with "immutable: builder used after Build".
type ListBuilder[T any] struct {
	root vector.Root[T]
	win  *transient.Window
}

func (b *ListBuilder[T]) checkNotBuilt() {
	checkBuilderWindow(b.win)
}

// Append adds items to the end under the builder's exclusive owner, in
// order, and returns b for chaining.
func (b *ListBuilder[T]) Append(items ...T) *ListBuilder[T] {
	b.checkNotBuilt()
	b.root = b.root.TransientAppend(b.win.Owner(), items...)
	return b
}

// Set replaces the element at index i with v under the builder's exclusive
// owner and returns b for chaining. It panics with
// "immutable: index out of range" when i is outside [0, Len).
func (b *ListBuilder[T]) Set(i int, v T) *ListBuilder[T] {
	b.checkNotBuilt()
	if i < 0 || i >= b.root.Len() {
		panic(panicIndexOutOfRange)
	}
	b.root = b.root.TransientSet(b.win.Owner(), i, v)
	return b
}

// Delete removes the element at index i under the builder's exclusive owner,
// shifting every later element down by one, and returns b for chaining. It
// panics with "immutable: index out of range" when i is outside [0, Len).
func (b *ListBuilder[T]) Delete(i int) *ListBuilder[T] {
	b.checkNotBuilt()
	if i < 0 || i >= b.root.Len() {
		panic(panicIndexOutOfRange)
	}
	b.root = b.root.TransientDelete(b.win.Owner(), i)
	return b
}

// Len returns the number of elements currently held by the builder.
func (b *ListBuilder[T]) Len() int {
	b.checkNotBuilt()
	return b.root.Len()
}

// At returns the element at index i currently held by the builder. It
// panics with "immutable: index out of range" when i is outside [0, Len).
// After Build, At panics with "immutable: builder used after Build".
func (b *ListBuilder[T]) At(i int) T {
	b.checkNotBuilt()
	if i < 0 || i >= b.root.Len() {
		panic(panicIndexOutOfRange)
	}
	return b.root.At(i)
}

// Build returns an immutable List holding the builder's current root and
// permanently disables the builder: every method, including a repeated
// Build, panics with "immutable: builder used after Build" from this point
// on. Freeze invalidates the owner token in O(1); nodes may retain stale
// inactive stamps.
func (b *ListBuilder[T]) Build() List[T] {
	freezeBuilderWindow(&b.win)
	root := b.root
	b.root = vector.Root[T]{}
	return listFromRoot(root)
}
