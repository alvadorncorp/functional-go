package vector

import (
	"github.com/alvadorncorp/gopackages/functional/internal/core/cow"
	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

// cloneBranch returns a shallow copy of an internal node (child pointers
// shared) stamped with owner. Pure immutable callers pass owner == nil.
func cloneBranch[T any](n *node[T], owner *ownership.Owner) *node[T] {
	return cow.EnsureEditable(n.owner, owner, n, func() *node[T] {
		children := make([]*node[T], len(n.children))
		copy(children, n.children)
		return &node[T]{owner: owner, children: children}
	})
}

// cloneLeaf returns a shallow copy of a leaf (values copied) stamped with owner.
// Pure immutable callers pass owner == nil.
func cloneLeaf[T any](n *node[T], owner *ownership.Owner) *node[T] {
	return cow.EnsureEditable(n.owner, owner, n, func() *node[T] {
		values := make([]T, len(n.values))
		copy(values, n.values)
		return &node[T]{owner: owner, values: values}
	})
}
