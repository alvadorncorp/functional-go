package wbtree

import (
	"cmp"
	"testing"
)

// checkInvariants fails t unless every node of r has a correct size, both
// children satisfy Straka's balance predicate on weight size+1, and keys are
// strictly ascending under r's comparator.
func checkInvariants[K, V any](t *testing.T, r Root[K, V]) {
	t.Helper()
	if r.node != nil && r.cmp == nil {
		t.Fatal("non-empty root without comparator")
	}
	var prev *node[K, V]
	var walk func(n *node[K, V]) int
	walk = func(n *node[K, V]) int {
		if n == nil {
			return 0
		}
		sl := walk(n.left)
		if prev != nil && r.cmp(prev.key, n.key) >= 0 {
			t.Fatalf("keys out of order: %v then %v", prev.key, n.key)
		}
		prev = n
		sr := walk(n.right)
		if n.size != sl+sr+1 {
			t.Fatalf("node %v: size %d, want %d", n.key, n.size, sl+sr+1)
		}
		if delta*(sl+1) < sr+1 || delta*(sr+1) < sl+1 {
			t.Fatalf("node %v unbalanced: left %d, right %d", n.key, sl, sr)
		}
		return n.size
	}
	walk(r.node)
}

// TestSequentialInsertDelete drives the rotation paths hardest: ascending and
// descending runs always insert or delete at one edge of the tree.
func TestSequentialInsertDelete(t *testing.T) {
	t.Parallel()

	const n = 10_000
	ascending := make([]int, n)
	descending := make([]int, n)
	for i := range n {
		ascending[i] = i
		descending[i] = n - 1 - i
	}
	tests := map[string]struct{ insert, delete []int }{
		"ascending":            {ascending, ascending},
		"descending":           {descending, descending},
		"ascending-descending": {ascending, descending},
		"descending-ascending": {descending, ascending},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := New[int, int](cmp.Compare[int])
			for i, k := range tc.insert {
				r = r.Set(k, k)
				checkInvariants(t, r)
				if r.Len() != i+1 {
					t.Fatalf("Len = %d after %d inserts", r.Len(), i+1)
				}
			}
			for i, k := range tc.delete {
				r = r.Delete(k)
				checkInvariants(t, r)
				if r.Len() != n-i-1 {
					t.Fatalf("Len = %d after %d deletes", r.Len(), i+1)
				}
			}
			if r.node != nil || !r.HasComparator() {
				t.Fatal("emptied root must have no nodes and keep its comparator")
			}
		})
	}
}

func TestDeleteMissingNoOp(t *testing.T) {
	t.Parallel()

	r := New[int, int](cmp.Compare[int]).Set(1, 1).Set(2, 2)
	d := r.Delete(99)
	if d.node != r.node || d.Len() != r.Len() {
		t.Fatal("missing Delete must return the same node pointer")
	}
}

func TestDeleteLastKeyKeepsComparator(t *testing.T) {
	t.Parallel()

	r := New[string, int](cmp.Compare[string]).Set("k", 1).Delete("k")
	if r.Len() != 0 || r.node != nil {
		t.Fatalf("Len = %d, want empty", r.Len())
	}
	if !r.HasComparator() {
		t.Fatal("emptied root lost its comparator")
	}
	assertGet(t, r.Set("again", 2), "again", 2)
}

func TestOwnerNilOnPublishedNodes(t *testing.T) {
	t.Parallel()

	r := New[int, int](cmp.Compare[int])
	for i := range 64 {
		r = r.Set(i, i)
	}
	for i := 0; i < 64; i += 3 {
		r = r.Delete(i)
	}
	assertOwnersNil(t, r.node)
}

func assertOwnersNil[K, V any](t *testing.T, n *node[K, V]) {
	t.Helper()
	if n == nil {
		return
	}
	if n.owner != nil {
		t.Fatalf("node %v owner must be nil on the pure path", n.key)
	}
	assertOwnersNil(t, n.left)
	assertOwnersNil(t, n.right)
}
