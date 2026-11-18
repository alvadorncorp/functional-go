package wbtree

import (
	"go.alvadorn.org/functional/internal/core/cow"
	"go.alvadorn.org/functional/internal/core/ownership"
)

// node is one entry of the tree. size counts the entries of the subtree
// rooted here. A nil *node is the empty subtree.
type node[K, V any] struct {
	owner       *ownership.Owner // nil for published path-copy nodes
	key         K
	val         V
	left, right *node[K, V]
	size        int
}

func size[K, V any](n *node[K, V]) int {
	if n == nil {
		return 0
	}
	return n.size
}

// edit returns n when exclusively owned by owner; otherwise a shallow clone
// (children shared) stamped with owner. A nil owner always clones.
func edit[K, V any](n *node[K, V], owner *ownership.Owner) *node[K, V] {
	return cow.EnsureEditable(n.owner, owner, n, func() *node[K, V] {
		c := *n
		c.owner = owner
		return &c
	})
}

// freshOf returns got, the child a pure insert just rebuilt, as the node
// balance may rotate without cloning again; nil under an owner. A pure insert
// always returns a new allocation (a path copy, a rebalanced root or a new
// leaf), never a published node. A nil owner never marks nodes, hence the
// explicit check; transient edits already skip owned nodes.
func freshOf[K, V any](owner *ownership.Owner, got *node[K, V]) *node[K, V] {
	if owner != nil {
		return nil
	}
	return got
}

// insert returns the subtree n with key mapped to val. Comparisons run on the
// way down; nodes are edited only on the way back up.
func insert[K, V any](owner *ownership.Owner, cmp func(a, b K) int, n *node[K, V], key K, val V) *node[K, V] {
	if n == nil {
		return &node[K, V]{owner: owner, key: key, val: val, size: 1}
	}
	c := cmp(key, n.key)
	if c == 0 {
		m := edit(n, owner)
		m.key, m.val = key, val
		return m
	}
	if c < 0 {
		left := insert(owner, cmp, n.left, key, val)
		m := edit(n, owner)
		m.left = left
		return balance(owner, m, freshOf(owner, left))
	}
	right := insert(owner, cmp, n.right, key, val)
	m := edit(n, owner)
	m.right = right
	return balance(owner, m, freshOf(owner, right))
}

// remove returns the subtree n without key and whether key was present. When
// key is absent, n itself is returned and nothing is edited.
func remove[K, V any](owner *ownership.Owner, cmp func(a, b K) int, n *node[K, V], key K) (*node[K, V], bool) {
	if n == nil {
		return nil, false
	}
	c := cmp(key, n.key)
	switch {
	case c < 0:
		left, removed := remove(owner, cmp, n.left, key)
		if !removed {
			return n, false
		}
		m := edit(n, owner)
		m.left = left
		// A delete only lightens the changed side, so balance rotates the
		// published sibling, never the rebuilt child: nothing is fresh.
		return balance(owner, m, nil), true
	case c > 0:
		right, removed := remove(owner, cmp, n.right, key)
		if !removed {
			return n, false
		}
		m := edit(n, owner)
		m.right = right
		return balance(owner, m, nil), true
	}
	if n.left == nil {
		return n.right, true
	}
	if n.right == nil {
		return n.left, true
	}
	// Two children: the right subtree's minimum takes n's place.
	right, successor := removeMin(owner, n.right)
	m := edit(n, owner)
	m.key, m.val, m.right = successor.key, successor.val, right
	return balance(owner, m, nil), true
}

// removeMin returns the non-empty subtree n without its minimum and that
// minimum node. It needs no comparisons.
func removeMin[K, V any](owner *ownership.Owner, n *node[K, V]) (rest, minimum *node[K, V]) {
	if n.left == nil {
		return n.right, n
	}
	left, minimum := removeMin(owner, n.left)
	m := edit(n, owner)
	m.left = left
	return balance(owner, m, nil), minimum
}

// forEach yields the subtree in ascending order and reports whether the
// caller should continue.
func (n *node[K, V]) forEach(yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	return n.left.forEach(yield) && yield(n.key, n.val) && n.right.forEach(yield)
}
