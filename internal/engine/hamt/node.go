package hamt

import (
	"math/bits"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

const (
	bitsPerLevel = 5
	levelMask    = 1<<bitsPerLevel - 1 // 0x1f
	// maxShift is the last shift still inside a 64-bit hash (0,5,...,60).
	maxShift = 60
)

// node is the unexported HAMT spine.
// Pure immutable set/del always path-copy and publish new nodes with owner=nil.
// Transient tset/tdel mutate nodes exclusively owned by the active token and
// shallow-clone foreign/published nodes, stamping the clone with owner.
type node[K comparable, V any] interface {
	get(hash uint64, shift uint, key K) (V, bool)
	// set returns the replacement node and whether key was newly inserted.
	set(hash uint64, shift uint, key K, val V) (node[K, V], bool)
	// del returns the replacement node (nil if empty) and whether key was removed.
	del(hash uint64, shift uint, key K) (node[K, V], bool)
	// tset is the owner-scoped edit path used by builders.
	tset(owner *ownership.Owner, hash uint64, shift uint, key K, val V) (node[K, V], bool)
	// tdel is the owner-scoped delete path used by builders.
	tdel(owner *ownership.Owner, hash uint64, shift uint, key K) (node[K, V], bool)
	forEach(yield func(K, V) bool) bool
}

func hashIndex(hash uint64, shift uint) uint {
	return uint((hash >> shift) & levelMask)
}

func bitpos(idx uint) uint32 {
	return uint32(1) << idx
}

func childIndex(bitmap uint32, idx uint) int {
	return bits.OnesCount32(bitmap & (bitpos(idx) - 1))
}

// merge combines nodes with unequal full hashes, branching at or below shift.
// Callers handle equal hashes before merging; unequal hashes must diverge by maxShift.
// owner is stamped on every newly allocated node (nil for pure immutable path).
func merge[K comparable, V any](
	shift uint,
	h1 uint64, n1 node[K, V],
	h2 uint64, n2 node[K, V],
	owner *ownership.Owner,
) node[K, V] {
	if h1 == h2 || shift > maxShift {
		panic("hamt: unreachable merge")
	}

	i1 := hashIndex(h1, shift)
	i2 := hashIndex(h2, shift)
	if i1 != i2 {
		var bm uint32
		children := make([]node[K, V], 2)
		if i1 < i2 {
			bm = bitpos(i1) | bitpos(i2)
			children[0], children[1] = n1, n2
		} else {
			bm = bitpos(i1) | bitpos(i2)
			children[0], children[1] = n2, n1
		}
		return &bitmapNode[K, V]{owner: owner, bitmap: bm, children: children}
	}
	child := merge(shift+bitsPerLevel, h1, n1, h2, n2, owner)
	return &bitmapNode[K, V]{
		owner:    owner,
		bitmap:   bitpos(i1),
		children: []node[K, V]{child},
	}
}
