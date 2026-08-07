package vector

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

const (
	bitsPerLevel = 5
	branching    = 1 << bitsPerLevel // 32
	indexMask    = branching - 1     // 0x1f
)

// node is a 32-way trie node. Internal nodes use children; leaf nodes use
// values. Pure path-copy nodes are born with owner == nil (published).
type node[T any] struct {
	owner    *ownership.Owner
	children []*node[T] // internal levels; dense left-to-right, len 0..32
	values   []T        // leaf level; len 0..32
}

// newPath wraps leaf in a chain of single-child branches of the given level.
// level is a multiple of bitsPerLevel; level == 0 returns leaf as-is.
// New nodes are published with owner == nil.
func newPath[T any](level uint, leaf *node[T]) *node[T] {
	if level == 0 {
		return leaf
	}
	return &node[T]{children: []*node[T]{newPath(level-bitsPerLevel, leaf)}}
}

// tnewPath is newPath stamped with owner for exclusive transient construction.
func tnewPath[T any](level uint, leaf *node[T], owner *ownership.Owner) *node[T] {
	if level == 0 {
		return leaf
	}
	return &node[T]{
		owner:    owner,
		children: []*node[T]{tnewPath(level-bitsPerLevel, leaf, owner)},
	}
}

// pushTail inserts a full 32-element leaf into the trie, path-copying the
// spine. count is the vector length before the append that overflows the tail
// (so count is a positive multiple of 32).
//
// New nodes are published with owner == nil.
func pushTail[T any](level uint, parent, tailLeaf *node[T], count int) *node[T] {
	subidx := ((count - 1) >> level) & indexMask

	// Path-copy parent; grow by one slot when appending a new child.
	newLen := len(parent.children)
	if subidx >= newLen {
		newLen = subidx + 1
	}
	children := make([]*node[T], newLen)
	copy(children, parent.children)

	var insert *node[T]
	switch {
	case level == bitsPerLevel:
		insert = tailLeaf
	case subidx < len(parent.children) && parent.children[subidx] != nil:
		insert = pushTail(level-bitsPerLevel, parent.children[subidx], tailLeaf, count)
	default:
		insert = newPath[T](level-bitsPerLevel, tailLeaf)
	}
	children[subidx] = insert
	return &node[T]{children: children}
}

// tpushTail inserts a full leaf under owner: ensureEditable on the spine,
// mutate in place when exclusively owned, stamp new nodes with owner.
func tpushTail[T any](level uint, parent, tailLeaf *node[T], count int, owner *ownership.Owner) *node[T] {
	subidx := ((count - 1) >> level) & indexMask

	parent = cloneBranch(parent, owner)
	if subidx >= len(parent.children) {
		children := make([]*node[T], subidx+1)
		copy(children, parent.children)
		parent.children = children
	}

	var insert *node[T]
	switch {
	case level == bitsPerLevel:
		insert = tailLeaf
	case parent.children[subidx] != nil:
		insert = tpushTail(level-bitsPerLevel, parent.children[subidx], tailLeaf, count, owner)
	default:
		insert = tnewPath[T](level-bitsPerLevel, tailLeaf, owner)
	}
	parent.children[subidx] = insert
	return parent
}

// doAssoc path-copies the trie and replaces the value at index i.
// New nodes are published with owner == nil.
func doAssoc[T any](level uint, n *node[T], i int, v T) *node[T] {
	if level == 0 {
		out := cloneLeaf(n, nil)
		out.values[i&indexMask] = v
		return out
	}
	subidx := (i >> level) & indexMask
	out := cloneBranch(n, nil)
	out.children[subidx] = doAssoc(level-bitsPerLevel, n.children[subidx], i, v)
	return out
}

// tdoAssoc is doAssoc under owner: ensureEditable, mutate when owned.
func tdoAssoc[T any](level uint, n *node[T], i int, v T, owner *ownership.Owner) *node[T] {
	if level == 0 {
		out := cloneLeaf(n, owner)
		out.values[i&indexMask] = v
		return out
	}
	subidx := (i >> level) & indexMask
	out := cloneBranch(n, owner)
	out.children[subidx] = tdoAssoc(level-bitsPerLevel, out.children[subidx], i, v, owner)
	return out
}

// popTail removes the rightmost leaf from the trie rooted at parent.
// count is the vector length while the one-element tail is still present
// (the last trie-resident index is count-2). Returns nil when the subtree
// becomes empty after the removal. New nodes are published with owner == nil.
func popTail[T any](level uint, parent *node[T], count int) *node[T] {
	subidx := ((count - 2) >> level) & indexMask

	if level > bitsPerLevel {
		newChild := popTail(level-bitsPerLevel, parent.children[subidx], count)
		if newChild == nil && subidx == 0 {
			return nil
		}
		newLen := subidx + 1
		if newChild == nil {
			newLen = subidx
		}
		children := make([]*node[T], newLen)
		copy(children, parent.children[:newLen])
		if newChild != nil {
			children[subidx] = newChild
		}
		return &node[T]{children: children}
	}

	// Parent of leaves (level == bitsPerLevel).
	if subidx == 0 {
		return nil
	}
	children := make([]*node[T], subidx)
	copy(children, parent.children[:subidx])
	return &node[T]{children: children}
}

// tpopTail is popTail under owner: ensureEditable on the spine, shrink or
// replace the rightmost child in place when exclusively owned.
func tpopTail[T any](level uint, parent *node[T], count int, owner *ownership.Owner) *node[T] {
	subidx := ((count - 2) >> level) & indexMask

	if level > bitsPerLevel {
		newChild := tpopTail(level-bitsPerLevel, parent.children[subidx], count, owner)
		if newChild == nil && subidx == 0 {
			return nil
		}
		parent = cloneBranch(parent, owner)
		if newChild == nil {
			parent.children = parent.children[:subidx]
		} else {
			parent.children[subidx] = newChild
		}
		return parent
	}

	if subidx == 0 {
		return nil
	}
	parent = cloneBranch(parent, owner)
	parent.children = parent.children[:subidx]
	return parent
}
