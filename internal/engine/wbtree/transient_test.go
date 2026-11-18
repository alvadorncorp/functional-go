package wbtree

import (
	"cmp"
	"maps"
	"testing"

	"go.alvadorn.org/functional/internal/core/ownership"
)

func TestTransientEditsInPlaceUnderSameOwner(t *testing.T) {
	t.Parallel()

	owner := newOwner(t)
	r := New[int, int](cmp.Compare[int])
	for i := range 32 {
		r = r.TransientSet(owner, i, i)
	}
	checkInvariants(t, r)

	before := nodeSet(r.node)
	replaced := r.TransientSet(owner, 7, 70)
	if replaced.node != r.node || !maps.Equal(nodeSet(replaced.node), before) {
		t.Fatal("replacing a key under the same owner must edit in place")
	}
	assertGet(t, replaced, 7, 70)

	before = nodeSet(replaced.node)
	grown := replaced.TransientSet(owner, 100, 100)
	checkInvariants(t, grown)
	fresh := 0
	for n := range nodeSet(grown.node) {
		if !before[n] {
			fresh++
		}
	}
	if fresh != 1 {
		t.Fatalf("insert under the same owner allocated %d nodes, want 1", fresh)
	}

	// grown is consumed by the in-place edit below, so read it first.
	before, wantLen, rootKey := nodeSet(grown.node), grown.Len()-1, grown.node.key
	shrunk := grown.TransientDelete(owner, rootKey)
	checkInvariants(t, shrunk)
	for n := range nodeSet(shrunk.node) {
		if !before[n] {
			t.Fatal("delete under the same owner must not allocate")
		}
	}
	if shrunk.Len() != wantLen {
		t.Fatalf("Len = %d, want %d", shrunk.Len(), wantLen)
	}
	if _, ok := shrunk.Get(rootKey); ok {
		t.Fatalf("deleted key %d still present", rootKey)
	}
}

func TestTransientLeavesPublishedBaseIntact(t *testing.T) {
	t.Parallel()

	base := New[int, int](cmp.Compare[int])
	for i := range 48 {
		base = base.Set(i, i)
	}
	want := maps.Collect(base.All())
	baseNodes := nodeSet(base.node)

	owner := newOwner(t)
	r := base
	r = r.TransientSet(owner, 100, 100)
	r = r.TransientSet(owner, 0, 999)
	r = r.TransientDelete(owner, 1)
	owner.Invalidate()

	checkInvariants(t, base)
	checkInvariants(t, r)
	if got := maps.Collect(base.All()); !maps.Equal(got, want) {
		t.Fatalf("base changed: %v", got)
	}
	if !maps.Equal(nodeSet(base.node), baseNodes) {
		t.Fatal("base node graph changed")
	}
	assertGet(t, r, 0, 999)
	assertGet(t, r, 100, 100)
	if _, ok := r.Get(1); ok {
		t.Fatal("1 should be deleted")
	}
	if r.Len() != base.Len() {
		t.Fatalf("Len = %d, want %d", r.Len(), base.Len())
	}
}

func TestTransientCopiesAfterFrozenOwner(t *testing.T) {
	t.Parallel()

	ownerX := ownership.New()
	v1 := New[int, int](cmp.Compare[int])
	for i := range 64 {
		v1 = v1.TransientSet(ownerX, i, i)
	}
	ownerX.Invalidate() // freeze: stamps stay but no longer grant edits
	want := maps.Collect(v1.All())
	v1Nodes := nodeSet(v1.node)

	ownerY := newOwner(t)
	v2 := v1.TransientSet(ownerY, 1000, 1000)
	v2 = v2.TransientSet(ownerY, 63, -63)

	if v2.node == v1.node {
		t.Fatal("root must be copied under a new owner after freeze")
	}
	if got := maps.Collect(v1.All()); !maps.Equal(got, want) {
		t.Fatalf("frozen version changed: %v", got)
	}
	if !maps.Equal(nodeSet(v1.node), v1Nodes) {
		t.Fatal("frozen version node graph changed")
	}
	if v2.node.left != v1.node.left {
		t.Fatal("untouched left subtree must be shared with the frozen version")
	}
	checkInvariants(t, v2)
	assertGet(t, v2, 1000, 1000)
	assertGet(t, v2, 63, -63)
}

func TestTransientDeleteMissingNoOp(t *testing.T) {
	t.Parallel()

	owner := newOwner(t)
	r := New[int, int](cmp.Compare[int]).Set(1, 1).TransientSet(owner, 2, 2)
	d := r.TransientDelete(owner, 99)
	if d.node != r.node || d.Len() != r.Len() {
		t.Fatal("missing TransientDelete must return the same node pointer")
	}
}

// TestComparatorPanicLeavesOwnedNodesUntouched panics at every comparison of
// a transient edit's path in turn. Owned nodes are edited in place, so any
// edit made before the panic would show up in a snapshot of every node.
func TestComparatorPanicLeavesOwnedNodesUntouched(t *testing.T) {
	t.Parallel()

	// The tree holds the even keys 0..62; mid-range keys turn both ways.
	ops := map[string]func(r Root[int, int], owner *ownership.Owner){
		"set new":      func(r Root[int, int], o *ownership.Owner) { r.TransientSet(o, 21, 21) },
		"set existing": func(r Root[int, int], o *ownership.Owner) { r.TransientSet(o, 42, -1) },
		"delete root":  func(r Root[int, int], o *ownership.Owner) { r.TransientDelete(o, r.node.key) },
		"delete mid":   func(r Root[int, int], o *ownership.Owner) { r.TransientDelete(o, 22) },
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			calls, panicAt := 0, 0
			counting := func(a, b int) int {
				calls++
				if calls == panicAt {
					panic("comparator failure")
				}
				return cmp.Compare(a, b)
			}
			owner := newOwner(t)
			r := New[int, int](counting)
			for i := range 32 {
				r = r.TransientSet(owner, 2*i, 2*i)
			}
			want := snapshot(r.node)

			for at := 1; ; at++ {
				calls, panicAt = 0, at
				if !panics(func() { op(r, owner) }) {
					break // the edit completed: every comparison was already tried
				}
				if !maps.Equal(snapshot(r.node), want) {
					t.Fatalf("panic at comparison %d edited an owned node", at)
				}
			}
		})
	}
}

// panics reports whether fn panicked.
func panics(fn func()) (panicked bool) {
	defer func() { panicked = recover() != nil }()
	fn()
	return false
}

// snapshot copies every node of the subtree n, keyed by its pointer.
func snapshot(n *node[int, int]) map[*node[int, int]]node[int, int] {
	nodes := map[*node[int, int]]node[int, int]{}
	for nd := range nodeSet(n) {
		nodes[nd] = *nd
	}
	return nodes
}

func newOwner(t *testing.T) *ownership.Owner {
	t.Helper()
	owner := ownership.New()
	t.Cleanup(owner.Invalidate)
	return owner
}

func nodeSet[K, V any](n *node[K, V]) map[*node[K, V]]bool {
	set := map[*node[K, V]]bool{}
	var walk func(*node[K, V])
	walk = func(n *node[K, V]) {
		if n == nil {
			return
		}
		set[n] = true
		walk(n.left)
		walk(n.right)
	}
	walk(n)
	return set
}
