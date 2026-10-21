package wbtree

import (
	"cmp"
	"testing"
)

// TestPureEditsShareUntouchedSubtrees edits one edge of a 1024-entry tree and
// checks that the opposite subtree of the root is reused by pointer while the
// prior version keeps every node it had.
func TestPureEditsShareUntouchedSubtrees(t *testing.T) {
	t.Parallel()

	const n = 1024
	base := New[int, int](cmp.Compare[int])
	for i := range n {
		base = base.Set(i, i)
	}
	baseNodes := nodeSet(base.node)

	tests := map[string]struct {
		edit       func(Root[int, int]) Root[int, int]
		sharedSide func(*node[int, int]) *node[int, int]
	}{
		"set past max":  {func(r Root[int, int]) Root[int, int] { return r.Set(n, n) }, leftChild},
		"delete max":    {func(r Root[int, int]) Root[int, int] { return r.Delete(n - 1) }, leftChild},
		"set below min": {func(r Root[int, int]) Root[int, int] { return r.Set(-1, -1) }, rightChild},
		"delete min":    {func(r Root[int, int]) Root[int, int] { return r.Delete(0) }, rightChild},
		"replace mid":   {func(r Root[int, int]) Root[int, int] { return r.Set(n/2, -1) }, nil},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			next := tc.edit(base)
			checkInvariants(t, next)
			if next.node == base.node {
				t.Fatal("root must be path-copied")
			}
			if tc.sharedSide != nil && tc.sharedSide(next.node) != tc.sharedSide(base.node) {
				t.Fatal("untouched root subtree must be shared by pointer")
			}
			nextNodes := nodeSet(next.node)
			fresh := 0
			for nd := range nextNodes {
				if !baseNodes[nd] {
					fresh++
				}
			}
			if limit := height(base.node) + 2; fresh > limit {
				t.Fatalf("edit allocated %d nodes, want at most %d (path plus one rotation)", fresh, limit)
			}
			assertOwnersNil(t, next.node)
		})
	}
}

func leftChild(n *node[int, int]) *node[int, int]  { return n.left }
func rightChild(n *node[int, int]) *node[int, int] { return n.right }

func height[K, V any](n *node[K, V]) int {
	if n == nil {
		return 0
	}
	return 1 + max(height(n.left), height(n.right))
}

// TestRotationsNeverMutatePublishedNodes snapshots every node of a published
// tree, applies each possible pure Set and Delete, and requires the snapshot
// to be unchanged. Over every size and key, inserts exercise rotations
// of the just-copied (fresh) child, and light-side deletes exercise rotations
// of a shared published sibling.
func TestRotationsNeverMutatePublishedNodes(t *testing.T) {
	t.Parallel()

	for n := 1; n <= 150; n++ {
		base := New[int, int](cmp.Compare[int])
		for i := range n {
			base = base.Set(2*i, i)
		}
		before := shapeOf(base.node)
		for k := -1; k <= 2*n; k++ {
			next := base.Set(k, -1)
			checkInvariants(t, next)
			assertUnchanged(t, before, base.node)
			next = base.Delete(k)
			checkInvariants(t, next)
			assertUnchanged(t, before, base.node)
		}
	}
}

func shapeOf(root *node[int, int]) map[*node[int, int]]node[int, int] {
	s := map[*node[int, int]]node[int, int]{}
	for nd := range nodeSet(root) {
		s[nd] = *nd
	}
	return s
}

func assertUnchanged(t *testing.T, before map[*node[int, int]]node[int, int], root *node[int, int]) {
	t.Helper()
	got := nodeSet(root)
	if len(got) != len(before) {
		t.Fatalf("published tree has %d nodes, want %d", len(got), len(before))
	}
	for nd, want := range before {
		if !got[nd] || *nd != want {
			t.Fatalf("published node %d mutated: %+v, want %+v", want.key, *nd, want)
		}
	}
}
