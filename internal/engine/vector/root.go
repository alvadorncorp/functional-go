package vector

import (
	"iter"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// Root is an immutable 32-way persistent vector.
//
// The zero value is a valid empty vector (same as Empty). Pure immutable
// Set/Append/Delete/Pop never mutate r; they return a new Root that
// structurally shares untouched trie nodes with prior versions.
//
// TransientAppend/TransientSet/TransientDelete/TransientPop edit under an
// active *ownership.Owner: owned nodes (and an exclusively owned tail) mutate
// in place; foreign or published structure is shallow-cloned and stamped.
// After the owner is invalidated (Freeze), residual stamps are treated as
// published — no tree walk is required to publish.
//
// At, Set, Delete, and Pop panic with "vector: index out of range" on invalid
// indices (including Pop on an empty vector). Public List wrappers may map
// those panics later.
type Root[T any] struct {
	count     int
	shift     uint // root level shift (5, 10, 15, ...); 0 when the trie is unused
	root      *node[T]
	tail      []T              // up to 32 trailing elements; never mutated after publish
	tailOwner *ownership.Owner // exclusive ownership of tail; nil = published/shared
}

// Empty returns an empty Root. Equivalent to the zero value.
func Empty[T any]() Root[T] {
	return Root[T]{}
}

// Len reports the number of elements in r.
func (r Root[T]) Len() int { return r.count }

// tailoff is the index of the first element stored in the tail.
func (r Root[T]) tailoff() int {
	if r.count < branching {
		return 0
	}
	return ((r.count - 1) >> bitsPerLevel) << bitsPerLevel
}

// arrayFor returns the leaf values slice that holds index i (caller bounds-checks).
func (r Root[T]) arrayFor(i int) []T {
	if i >= r.tailoff() {
		return r.tail
	}
	return r.leafNode(i).values
}

// leafNode returns the trie leaf that holds index i (caller guarantees i is in
// the trie, not the tail).
func (r Root[T]) leafNode(i int) *node[T] {
	n := r.root
	for level := r.shift; level > 0; level -= bitsPerLevel {
		n = n.children[(i>>level)&indexMask]
	}
	return n
}

// At returns the element at index i.
// It panics with "vector: index out of range" when i < 0 or i >= Len().
func (r Root[T]) At(i int) T {
	if i < 0 || i >= r.count {
		panic("vector: index out of range")
	}
	return r.arrayFor(i)[i&indexMask]
}

// Set returns a new Root with index i replaced by v. r is unchanged.
// It panics with "vector: index out of range" when i < 0 or i >= Len().
// New path nodes are published with owner == nil.
func (r Root[T]) Set(i int, v T) Root[T] {
	if i < 0 || i >= r.count {
		panic("vector: index out of range")
	}
	if i >= r.tailoff() {
		newTail := make([]T, len(r.tail))
		copy(newTail, r.tail)
		newTail[i&indexMask] = v
		return Root[T]{count: r.count, shift: r.shift, root: r.root, tail: newTail}
	}
	return Root[T]{
		count: r.count,
		shift: r.shift,
		root:  doAssoc(r.shift, r.root, i, v),
		tail:  r.tail,
	}
}

// Append returns a new Root with items added at the end, in order.
// r is unchanged. New path nodes are published with owner == nil.
func (r Root[T]) Append(items ...T) Root[T] {
	for _, item := range items {
		r = r.appendOne(item)
	}
	return r
}

func (r Root[T]) appendOne(v T) Root[T] {
	// Fast path: room remains in the tail (fewer than 32 trailing elements).
	if r.count-r.tailoff() < branching {
		newTail := make([]T, len(r.tail)+1)
		copy(newTail, r.tail)
		newTail[len(r.tail)] = v
		return Root[T]{count: r.count + 1, shift: r.shift, root: r.root, tail: newTail}
	}
	// Tail is full: push it into the trie and start a fresh tail with v.
	return r.pushFullTail(v)
}

// pushFullTail pushes the full published tail into the trie and starts a new
// one-element tail with v. Caller guarantees the tail is full.
func (r Root[T]) pushFullTail(v T) Root[T] {
	// The old tail slice is immutable and may be shared by the new leaf.
	tailLeaf := &node[T]{values: r.tail}

	var newRoot *node[T]
	newShift := r.shift

	switch {
	case r.root == nil:
		// First leaf pushed into the tree (count == 32 on a zero/empty root).
		newRoot = &node[T]{children: []*node[T]{tailLeaf}}
		newShift = bitsPerLevel
	case uint(r.count)>>bitsPerLevel > 1<<r.shift:
		// Root overflow: add a level.
		newRoot = &node[T]{
			children: []*node[T]{
				r.root,
				newPath[T](r.shift, tailLeaf),
			},
		}
		newShift = r.shift + bitsPerLevel
	default:
		newRoot = pushTail(r.shift, r.root, tailLeaf, r.count)
	}

	return Root[T]{
		count: r.count + 1,
		shift: newShift,
		root:  newRoot,
		tail:  []T{v},
	}
}

// Delete returns a new Root without the element at index i; subsequent
// elements shift left by one. r is unchanged.
//
// Delete is O(n): this version rebuilds via a sequential walk (All) rather
// than a path-aware shift. It panics with "vector: index out of range" when
// i < 0 or i >= Len().
func (r Root[T]) Delete(i int) Root[T] {
	if i < 0 || i >= r.count {
		panic("vector: index out of range")
	}
	if r.count == 1 {
		return Empty[T]()
	}
	// O(n) rebuild from sequential walk: correct shift-left semantics without
	// a specialized path, and without O(log n) At per index.
	vals := make([]T, 0, r.count-1)
	for j, v := range r.All() {
		if j == i {
			continue
		}
		vals = append(vals, v)
	}
	return Empty[T]().Append(vals...)
}

// Pop returns a new Root without the last element and the removed value.
// r is unchanged. Complexity: O(1) when the tail has more than one element;
// O(log32 n) when the last full leaf must be pulled from the trie into the tail.
// It panics with "vector: index out of range" when Len() == 0
// (same string as At/Set/Delete for empty/invalid index).
func (r Root[T]) Pop() (next Root[T], val T) {
	if r.count == 0 {
		panic("vector: index out of range")
	}
	if r.count == 1 {
		return Empty[T](), r.tail[0]
	}
	if len(r.tail) > 1 {
		val = r.tail[len(r.tail)-1]
		newTail := make([]T, len(r.tail)-1)
		copy(newTail, r.tail[:len(r.tail)-1])
		return Root[T]{
			count: r.count - 1,
			shift: r.shift,
			root:  r.root,
			tail:  newTail,
		}, val
	}
	// Single-element tail: pull the last full leaf from the trie.
	return r.popPullLeaf()
}

// popPullLeaf implements the O(log32 n) Pop path when len(tail) == 1.
// Caller guarantees count > 1 and a non-nil root.
func (r Root[T]) popPullLeaf() (next Root[T], val T) {
	val = r.tail[0]
	newTail := r.arrayFor(r.count - 2) // share published leaf values
	newRoot := popTail(r.shift, r.root, r.count)
	newShift := r.shift
	if newRoot == nil {
		newShift = 0
	} else if newShift > bitsPerLevel && len(newRoot.children) == 1 {
		newRoot = newRoot.children[0]
		newShift -= bitsPerLevel
	}
	return Root[T]{
		count: r.count - 1,
		shift: newShift,
		root:  newRoot,
		tail:  newTail,
	}, val
}

// TransientSet returns a Root with index i replaced by v, editing under owner.
// Owned path nodes and an exclusively owned tail mutate in place; others are
// shallow-cloned and stamped. r's node graph is not mutated when r's nodes are
// published or owned by a different/inactive token.
// It panics with "vector: index out of range" when i < 0 or i >= Len().
func (r Root[T]) TransientSet(owner *ownership.Owner, i int, v T) Root[T] {
	if i < 0 || i >= r.count {
		panic("vector: index out of range")
	}
	if i >= r.tailoff() {
		tail := r.editableTail(owner)
		tail[i&indexMask] = v
		return Root[T]{
			count:     r.count,
			shift:     r.shift,
			root:      r.root,
			tail:      tail,
			tailOwner: owner,
		}
	}
	return Root[T]{
		count:     r.count,
		shift:     r.shift,
		root:      tdoAssoc(r.shift, r.root, i, v, owner),
		tail:      r.tail,
		tailOwner: r.tailOwner,
	}
}

// TransientAppend returns a Root with items added at the end, editing under
// owner. Owned tails and trie spines mutate in place when the active token
// matches; published structure is copy-on-write and stamped.
func (r Root[T]) TransientAppend(owner *ownership.Owner, items ...T) Root[T] {
	for _, item := range items {
		r = r.transientAppendOne(owner, item)
	}
	return r
}

func (r Root[T]) transientAppendOne(owner *ownership.Owner, v T) Root[T] {
	// Fast path: room remains in the tail.
	if r.count-r.tailoff() < branching {
		return r.transientAppendToTail(owner, v)
	}
	// Tail is full: push it into the trie and start a fresh exclusive tail.
	return r.transientPushFullTail(owner, v)
}

// transientAppendToTail appends v into a non-full tail under owner.
func (r Root[T]) transientAppendToTail(owner *ownership.Owner, v T) Root[T] {
	var newTail []T
	if ownership.IsOwnedBy(r.tailOwner, owner) {
		// Owned: append may grow the exclusive backing store in place.
		newTail = r.tail
		newTail = append(newTail, v)
	} else {
		newTail = make([]T, len(r.tail)+1, branching)
		copy(newTail, r.tail)
		newTail[len(r.tail)] = v
	}
	return Root[T]{
		count:     r.count + 1,
		shift:     r.shift,
		root:      r.root,
		tail:      newTail,
		tailOwner: owner,
	}
}

// transientPushFullTail pushes the full tail into the trie under owner and
// starts a fresh exclusive one-element tail with v. Caller guarantees the
// tail is full.
func (r Root[T]) transientPushFullTail(owner *ownership.Owner, v T) Root[T] {
	// When the tail is exclusively owned, stamp the leaf; otherwise share the
	// published values with a nil-owner leaf so later edits still COW.
	var tailLeaf *node[T]
	if ownership.IsOwnedBy(r.tailOwner, owner) {
		tailLeaf = &node[T]{owner: owner, values: r.tail}
	} else {
		tailLeaf = &node[T]{values: r.tail}
	}

	var newRoot *node[T]
	newShift := r.shift

	switch {
	case r.root == nil:
		newRoot = &node[T]{owner: owner, children: []*node[T]{tailLeaf}}
		newShift = bitsPerLevel
	case uint(r.count)>>bitsPerLevel > 1<<r.shift:
		newRoot = &node[T]{
			owner: owner,
			children: []*node[T]{
				r.root,
				tnewPath[T](r.shift, tailLeaf, owner),
			},
		}
		newShift = r.shift + bitsPerLevel
	default:
		newRoot = tpushTail(r.shift, r.root, tailLeaf, r.count, owner)
	}

	newTail := make([]T, 1, branching)
	newTail[0] = v
	return Root[T]{
		count:     r.count + 1,
		shift:     newShift,
		root:      newRoot,
		tail:      newTail,
		tailOwner: owner,
	}
}

// TransientDelete returns a Root without the element at index i, editing
// under owner. Like Delete, this version rebuilds via sequential walk (O(n))
// but uses TransientAppend so the result is exclusively owned for further
// edits. It panics with "vector: index out of range" when i < 0 or i >= Len().
func (r Root[T]) TransientDelete(owner *ownership.Owner, i int) Root[T] {
	if i < 0 || i >= r.count {
		panic("vector: index out of range")
	}
	if r.count == 1 {
		return Empty[T]()
	}
	out := Empty[T]()
	for j, v := range r.All() {
		if j == i {
			continue
		}
		out = out.TransientAppend(owner, v)
	}
	return out
}

// TransientPop is the owner-scoped counterpart of Pop: exclusive tail/nodes may
// mutate in place; published structure is shallow-cloned and stamped.
// It panics with "vector: index out of range" when Len() == 0.
func (r Root[T]) TransientPop(owner *ownership.Owner) (next Root[T], val T) {
	if r.count == 0 {
		panic("vector: index out of range")
	}
	if r.count == 1 {
		return Empty[T](), r.tail[0]
	}
	if len(r.tail) > 1 {
		val = r.tail[len(r.tail)-1]
		var newTail []T
		if ownership.IsOwnedBy(r.tailOwner, owner) {
			newTail = r.tail[:len(r.tail)-1]
		} else {
			newTail = make([]T, len(r.tail)-1, branching)
			copy(newTail, r.tail[:len(r.tail)-1])
		}
		return Root[T]{
			count:     r.count - 1,
			shift:     r.shift,
			root:      r.root,
			tail:      newTail,
			tailOwner: owner,
		}, val
	}
	return r.transientPopPullLeaf(owner)
}

// transientPopPullLeaf implements the O(log32 n) TransientPop path when
// len(tail) == 1. Caller guarantees count > 1 and a non-nil root.
func (r Root[T]) transientPopPullLeaf(owner *ownership.Owner) (next Root[T], val T) {
	val = r.tail[0]
	leaf := r.leafNode(r.count - 2)
	newRoot := tpopTail(r.shift, r.root, r.count, owner)
	newShift := r.shift
	if newRoot == nil {
		newShift = 0
	} else if newShift > bitsPerLevel && len(newRoot.children) == 1 {
		newRoot = newRoot.children[0]
		newShift -= bitsPerLevel
	}
	var newTail []T
	if ownership.IsOwnedBy(leaf.owner, owner) {
		newTail = leaf.values
	} else {
		newTail = make([]T, len(leaf.values), branching)
		copy(newTail, leaf.values)
	}
	return Root[T]{
		count:     r.count - 1,
		shift:     newShift,
		root:      newRoot,
		tail:      newTail,
		tailOwner: owner,
	}, val
}

// editableTail returns a tail slice safe to mutate under owner. When the
// current tail is already exclusively owned, it is returned as-is; otherwise
// a copy with room to grow to a full branch is allocated.
func (r Root[T]) editableTail(owner *ownership.Owner) []T {
	if ownership.IsOwnedBy(r.tailOwner, owner) {
		return r.tail
	}
	capHint := max(len(r.tail), branching)
	out := make([]T, len(r.tail), capHint)
	copy(out, r.tail)
	return out
}

// All yields every (index, value) pair in ascending index order.
// The walk is sequential over the trie leaves then the tail (O(n)), not a
// per-index At (which would be O(n log n)).
func (r Root[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		idx := 0
		if r.root != nil {
			if !walkAll(r.root, r.shift, &idx, yield) {
				return
			}
		}
		for _, v := range r.tail {
			if !yield(idx, v) {
				return
			}
			idx++
		}
	}
}

// walkAll yields trie-resident values left-to-right, advancing *idx for each.
// Returns false when yield requests an early stop.
func walkAll[T any](n *node[T], level uint, idx *int, yield func(int, T) bool) bool {
	if level == 0 {
		for _, v := range n.values {
			if !yield(*idx, v) {
				return false
			}
			*idx++
		}
		return true
	}
	for _, ch := range n.children {
		if ch == nil {
			continue
		}
		if !walkAll(ch, level-bitsPerLevel, idx, yield) {
			return false
		}
	}
	return true
}
