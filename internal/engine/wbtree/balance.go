package wbtree

import "github.com/alvadorncorp/gopackages/functional/internal/core/ownership"

// Straka's balance parameters for weight = size+1 ("Adams' Trees Revisited",
// 2011): the only integer pair for which one rotation per level restores
// balance after a single insert or delete.
const (
	delta = 3
	gamma = 2
)

// weight is Straka's balance weight, size+1, so empty subtrees weigh 1.
func weight[K, V any](n *node[K, V]) int {
	return size(n) + 1
}

// isBalanced reports whether a is heavy enough to sit beside b:
// delta*(size(a)+1) >= size(b)+1.
func isBalanced[K, V any](a, b *node[K, V]) bool {
	return delta*weight(a) >= weight(b)
}

// isSingle reports whether a single rotation suffices when a is the inner
// and b the outer grandchild of the heavy side: size(a)+1 < gamma*(size(b)+1).
func isSingle[K, V any](a, b *node[K, V]) bool {
	return weight(a) < gamma*weight(b)
}

// balance restores the weight invariant at m after one insert or delete in
// one of its children and recomputes sizes. m must already be editable under
// owner; its children need not be.
func balance[K, V any](owner *ownership.Owner, m *node[K, V]) *node[K, V] {
	switch {
	case !isBalanced(m.left, m.right):
		return rotateLeft(owner, m)
	case !isBalanced(m.right, m.left):
		return rotateRight(owner, m)
	}
	resize(m)
	return m
}

// rotateLeft lifts the heavy right child (single) or its left child (double).
func rotateLeft[K, V any](owner *ownership.Owner, m *node[K, V]) *node[K, V] {
	r := edit(m.right, owner)
	if isSingle(r.left, r.right) {
		m.right = r.left
		resize(m)
		r.left = m
		resize(r)
		return r
	}
	rl := edit(r.left, owner)
	m.right, r.left = rl.left, rl.right
	resize(m)
	resize(r)
	rl.left, rl.right = m, r
	resize(rl)
	return rl
}

// rotateRight lifts the heavy left child (single) or its right child (double).
func rotateRight[K, V any](owner *ownership.Owner, m *node[K, V]) *node[K, V] {
	l := edit(m.left, owner)
	if isSingle(l.right, l.left) {
		m.left = l.right
		resize(m)
		l.right = m
		resize(l)
		return l
	}
	lr := edit(l.right, owner)
	m.left, l.right = lr.right, lr.left
	resize(m)
	resize(l)
	lr.left, lr.right = l, m
	resize(lr)
	return lr
}

func resize[K, V any](n *node[K, V]) {
	n.size = size(n.left) + size(n.right) + 1
}
