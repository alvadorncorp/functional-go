package hamt

import (
	"github.com/alvadorncorp/gopackages/functional/internal/core/cow"
	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// ensureEditableLeaf returns l when exclusively owned by owner; otherwise a
// shallow clone stamped with owner.
func ensureEditableLeaf[K comparable, V any](l *leaf[K, V], owner *ownership.Owner) *leaf[K, V] {
	return cow.EnsureEditable(l.owner, owner, l, func() *leaf[K, V] {
		c := *l
		c.owner = owner
		return &c
	})
}

// ensureEditableBitmap returns b when exclusively owned by owner; otherwise a
// shallow clone (children slice copied, child pointers shared) stamped with owner.
func ensureEditableBitmap[K comparable, V any](b *bitmapNode[K, V], owner *ownership.Owner) *bitmapNode[K, V] {
	return cow.EnsureEditable(b.owner, owner, b, func() *bitmapNode[K, V] {
		children := make([]node[K, V], len(b.children))
		copy(children, b.children)
		return &bitmapNode[K, V]{
			owner:    owner,
			bitmap:   b.bitmap,
			children: children,
		}
	})
}

// ensureEditableCollision returns c when exclusively owned by owner; otherwise a
// shallow clone (entries slice copied) stamped with owner.
func ensureEditableCollision[K comparable, V any](c *collision[K, V], owner *ownership.Owner) *collision[K, V] {
	return cow.EnsureEditable(c.owner, owner, c, func() *collision[K, V] {
		entries := make([]entry[K, V], len(c.entries))
		copy(entries, c.entries)
		return &collision[K, V]{
			owner:   owner,
			hash:    c.hash,
			entries: entries,
		}
	})
}
