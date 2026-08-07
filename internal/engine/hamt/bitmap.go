package hamt

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

// bitmapNode is an array-mapped trie node (branching factor 32).
// children is dense over the set bits of bitmap (popcount packing).
type bitmapNode[K comparable, V any] struct {
	owner    *ownership.Owner // nil for published path-copy nodes
	bitmap   uint32
	children []node[K, V]
}

func (b *bitmapNode[K, V]) get(hash uint64, shift uint, key K) (V, bool) {
	idx := hashIndex(hash, shift)
	bit := bitpos(idx)
	if b.bitmap&bit == 0 {
		var zero V
		return zero, false
	}
	return b.children[childIndex(b.bitmap, idx)].get(hash, shift+bitsPerLevel, key)
}

func (b *bitmapNode[K, V]) set(hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	idx := hashIndex(hash, shift)
	bit := bitpos(idx)
	if b.bitmap&bit == 0 {
		pos := childIndex(b.bitmap, idx)
		child := &leaf[K, V]{hash: hash, key: key, val: val}
		return &bitmapNode[K, V]{
			bitmap:   b.bitmap | bit,
			children: insertChild(b.children, pos, child),
		}, true
	}
	pos := childIndex(b.bitmap, idx)
	nextChild, added := b.children[pos].set(hash, shift+bitsPerLevel, key, val)
	return &bitmapNode[K, V]{
		bitmap:   b.bitmap,
		children: replaceChild(b.children, pos, nextChild),
	}, added
}

func (b *bitmapNode[K, V]) del(hash uint64, shift uint, key K) (node[K, V], bool) {
	idx := hashIndex(hash, shift)
	bit := bitpos(idx)
	if b.bitmap&bit == 0 {
		return b, false
	}
	pos := childIndex(b.bitmap, idx)
	nextChild, removed := b.children[pos].del(hash, shift+bitsPerLevel, key)
	if !removed {
		return b, false
	}
	if nextChild == nil {
		return b.removeSlot(bit, pos), true
	}
	return &bitmapNode[K, V]{
		bitmap:   b.bitmap,
		children: replaceChild(b.children, pos, nextChild),
	}, true
}

// removeSlot drops the child at pos. When two children remain, a leaf or
// collision sibling is returned directly (path compression): those nodes
// store full hashes so Root.Get may start at shift 0. An internal bitmap
// sibling must NOT be elevated — its bitmap is relative to a deeper shift.
// Pure path: always allocates a fresh bitmap when compression does not apply
// (owner nil).
func (b *bitmapNode[K, V]) removeSlot(bit uint32, pos int) node[K, V] {
	if len(b.children) == 1 {
		return nil
	}
	if len(b.children) == 2 {
		keep := b.children[1-pos]
		switch keep.(type) {
		case *leaf[K, V], *collision[K, V]:
			return keep
		}
	}
	return &bitmapNode[K, V]{
		bitmap:   b.bitmap &^ bit,
		children: removeChild(b.children, pos),
	}
}

func (b *bitmapNode[K, V]) tset(owner *ownership.Owner, hash uint64, shift uint, key K, val V) (node[K, V], bool) {
	b = ensureEditableBitmap(b, owner)
	idx := hashIndex(hash, shift)
	bit := bitpos(idx)
	if b.bitmap&bit == 0 {
		pos := childIndex(b.bitmap, idx)
		child := &leaf[K, V]{owner: owner, hash: hash, key: key, val: val}
		b.children = insertChild(b.children, pos, child)
		b.bitmap |= bit
		return b, true
	}
	pos := childIndex(b.bitmap, idx)
	nextChild, added := b.children[pos].tset(owner, hash, shift+bitsPerLevel, key, val)
	b.children[pos] = nextChild
	return b, added
}

func (b *bitmapNode[K, V]) tdel(owner *ownership.Owner, hash uint64, shift uint, key K) (node[K, V], bool) {
	idx := hashIndex(hash, shift)
	bit := bitpos(idx)
	if b.bitmap&bit == 0 {
		return b, false
	}
	pos := childIndex(b.bitmap, idx)
	// Peek: only clone when the child path will actually remove.
	// We still must clone before mutating children, so call tdel first on child
	// without cloning parent if child reports no removal — child may have been
	// cloned already on a true path; on false path child is unchanged.
	nextChild, removed := b.children[pos].tdel(owner, hash, shift+bitsPerLevel, key)
	if !removed {
		return b, false
	}
	b = ensureEditableBitmap(b, owner)
	// Recompute pos after clone: bitmap/layout identical so pos is stable.
	pos = childIndex(b.bitmap, idx)
	if nextChild == nil {
		return b.removeSlotOwned(bit, pos), true
	}
	b.children[pos] = nextChild
	return b, true
}

// removeSlotOwned is the transient counterpart of removeSlot: mutates an
// already-owned bitmap in place when compression does not apply.
func (b *bitmapNode[K, V]) removeSlotOwned(bit uint32, pos int) node[K, V] {
	if len(b.children) == 1 {
		return nil
	}
	if len(b.children) == 2 {
		keep := b.children[1-pos]
		switch keep.(type) {
		case *leaf[K, V], *collision[K, V]:
			return keep
		}
	}
	b.bitmap &^= bit
	b.children = removeChild(b.children, pos)
	return b
}

func (b *bitmapNode[K, V]) forEach(yield func(K, V) bool) bool {
	for _, ch := range b.children {
		if !ch.forEach(yield) {
			return false
		}
	}
	return true
}

func insertChild[K comparable, V any](children []node[K, V], pos int, child node[K, V]) []node[K, V] {
	next := make([]node[K, V], len(children)+1)
	copy(next[:pos], children[:pos])
	next[pos] = child
	copy(next[pos+1:], children[pos:])
	return next
}

func replaceChild[K comparable, V any](children []node[K, V], pos int, child node[K, V]) []node[K, V] {
	next := make([]node[K, V], len(children))
	copy(next, children)
	next[pos] = child
	return next
}

func removeChild[K comparable, V any](children []node[K, V], pos int) []node[K, V] {
	next := make([]node[K, V], len(children)-1)
	copy(next[:pos], children[:pos])
	copy(next[pos:], children[pos+1:])
	return next
}
