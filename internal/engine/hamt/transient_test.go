package hamt

import (
	"math"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/cow"
	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

func TestTransientMultiSetThenPublish(t *testing.T) {
	t.Parallel()

	base := Empty[string, int]().Set("a", 1).Set("b", 2)
	oldNode := base.node

	owner := ownership.New()
	r := base
	r = r.TransientSet(owner, "c", 3)
	r = r.TransientSet(owner, "a", 10)
	r = r.TransientDelete(owner, "b")

	// Publish: invalidate owner token (O(1), no tree walk).
	owner.Invalidate()

	if r.Len() != 2 {
		t.Fatalf("Len = %d, want 2", r.Len())
	}
	assertGet(t, r, "a", 10)
	assertGet(t, r, "c", 3)
	if _, ok := r.Get("b"); ok {
		t.Fatal("b should be deleted")
	}

	// Old immutable root unchanged.
	if base.node != oldNode {
		t.Fatal("base root node pointer changed")
	}
	assertGet(t, base, "a", 1)
	assertGet(t, base, "b", 2)
	if _, ok := base.Get("c"); ok {
		t.Fatal("base must not contain c")
	}

	// Stale stamps on r's nodes fail IsOwnedBy after invalidate.
	if r.node != nil {
		assertNotOwnedByInactive(t, r.node, owner)
	}
}

func TestTransientSharingAfterNewOwner(t *testing.T) {
	t.Parallel()

	// Build a wide published root.
	pub := Empty[int, int]()
	for i := range 64 {
		pub = pub.Set(i, i)
	}

	ownerX := ownership.New()
	v2 := pub
	// Touch a key that sits on one path.
	v2 = v2.TransientSet(ownerX, 1000, 1000)
	ownerX.Invalidate()

	// Second transient with a new owner; edit a different key.
	ownerY := ownership.New()
	v3 := v2.TransientSet(ownerY, 2000, 2000)

	ob, ok1 := v2.node.(*bitmapNode[int, int])
	nb, ok2 := v3.node.(*bitmapNode[int, int])
	if !ok1 || !ok2 {
		t.Fatalf("expected bitmap roots, got %T and %T", v2.node, v3.node)
	}
	if ob == nb {
		t.Fatal("root must be path-copied under a new owner")
	}

	shared := 0
	for idx := range uint(32) {
		bit := bitpos(idx)
		if ob.bitmap&bit == 0 || nb.bitmap&bit == 0 {
			continue
		}
		oc := ob.children[childIndex(ob.bitmap, idx)]
		nc := nb.children[childIndex(nb.bitmap, idx)]
		if oc == nc {
			shared++
		}
	}
	if shared == 0 {
		t.Fatal("expected at least one shared child pointer between published v2 and v3")
	}

	assertGet(t, v2, 0, 0)
	assertGet(t, v3, 0, 0)
	assertGet(t, v2, 1000, 1000)
	assertGet(t, v3, 1000, 1000)
	assertGet(t, v3, 2000, 2000)
	if _, ok := v2.Get(2000); ok {
		t.Fatal("v2 must not contain 2000")
	}
}

func TestMutationAfterFreezeNeedsCopy(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[string, int]().TransientSet(owner, "k", 1)
	lf, ok := r.node.(*leaf[string, int])
	if !ok {
		t.Fatalf("expected leaf, got %T", r.node)
	}
	if cow.NeedsCopy(lf.owner, owner) {
		t.Fatal("active owner must not need copy")
	}

	owner.Invalidate()
	if !cow.NeedsCopy(lf.owner, owner) {
		t.Fatal("after invalidate, same stamp must need copy")
	}

	// ensureEditable must clone under a fresh owner.
	fresh := ownership.New()
	cloned := ensureEditableLeaf(lf, fresh)
	if cloned == lf {
		t.Fatal("expected shallow clone after freeze")
	}
	if cloned.owner != fresh {
		t.Fatal("clone must be stamped with fresh owner")
	}
	// Original leaf value unchanged by clone.
	if lf.val != 1 {
		t.Fatalf("original val = %d", lf.val)
	}
	cloned.val = 99
	if lf.val != 1 {
		t.Fatal("mutating clone must not affect published leaf")
	}
}

func TestMultiVersionBuilderLineage(t *testing.T) {
	t.Parallel()

	// V1 pure immutable.
	v1 := Empty[int, int]()
	for i := range 48 {
		v1 = v1.Set(i, i)
	}
	assertOwnersNil(t, v1.node)

	// Builder X → V2
	ox := ownership.New()
	v2 := v1
	v2 = v2.TransientSet(ox, 100, 100)
	v2 = v2.TransientSet(ox, 0, 999)
	v2 = v2.TransientDelete(ox, 1)
	ox.Invalidate()

	// Builder Y → V3
	oy := ownership.New()
	v3 := v2
	v3 = v3.TransientSet(oy, 200, 200)
	v3 = v3.TransientSet(oy, 2, 222)
	oy.Invalidate()

	// Integrity of all versions.
	if v1.Len() != 48 {
		t.Fatalf("v1 Len = %d", v1.Len())
	}
	assertGet(t, v1, 0, 0)
	assertGet(t, v1, 1, 1)
	if _, ok := v1.Get(100); ok {
		t.Fatal("v1 must not have 100")
	}

	if v2.Len() != 48 { // removed 1, added 100: 48-1+1=48
		t.Fatalf("v2 Len = %d, want 48", v2.Len())
	}
	assertGet(t, v2, 0, 999)
	if _, ok := v2.Get(1); ok {
		t.Fatal("v2 must not have 1")
	}
	assertGet(t, v2, 100, 100)
	if _, ok := v2.Get(200); ok {
		t.Fatal("v2 must not have 200")
	}

	if v3.Len() != 49 {
		t.Fatalf("v3 Len = %d, want 49", v3.Len())
	}
	assertGet(t, v3, 0, 999)
	assertGet(t, v3, 2, 222)
	assertGet(t, v3, 100, 100)
	assertGet(t, v3, 200, 200)

	// V1 still fully intact after later builders.
	for i := range 48 {
		assertGet(t, v1, i, i)
	}

	// Sharing: V2 and V3 should share some untouched child under root.
	b2, ok2 := v2.node.(*bitmapNode[int, int])
	b3, ok3 := v3.node.(*bitmapNode[int, int])
	if !ok2 || !ok3 {
		t.Fatalf("expected bitmap roots, got %T %T", v2.node, v3.node)
	}
	shared := 0
	for idx := range uint(32) {
		bit := bitpos(idx)
		if b2.bitmap&bit == 0 || b3.bitmap&bit == 0 {
			continue
		}
		if b2.children[childIndex(b2.bitmap, idx)] == b3.children[childIndex(b3.bitmap, idx)] {
			shared++
		}
	}
	if shared == 0 {
		t.Fatal("expected structural sharing between v2 and v3 on untouched branches")
	}
}

func TestTransientInPlaceSecondEditSameOwner(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[float64, int]().TransientSet(owner, 0, 1)
	n1 := r.node
	r = r.TransientSet(owner, math.Copysign(0, -1), 2)
	if r.node != n1 {
		t.Fatal("second edit under same active owner should reuse root leaf pointer")
	}
	assertGet(t, r, 0, 2)
	if r.Len() != 1 {
		t.Fatalf("Len = %d, want 1", r.Len())
	}
	for key, val := range r.All() {
		if !math.Signbit(key) || val != 2 {
			t.Fatalf("replacement = (%v, %d), want (-0, 2)", key, val)
		}
	}
	owner.Invalidate()

	fresh := ownership.New()
	defer fresh.Invalidate()
	next := r.TransientSet(fresh, 0, 3)
	if next.node == r.node {
		t.Fatal("new owner must copy the frozen leaf")
	}
	if next.Len() != 1 {
		t.Fatalf("next Len = %d, want 1", next.Len())
	}
	for key, val := range next.All() {
		if math.Signbit(key) || val != 3 {
			t.Fatalf("replacement = (%v, %d), want (+0, 3)", key, val)
		}
	}
	for key, val := range r.All() {
		if !math.Signbit(key) || val != 2 {
			t.Fatalf("snapshot = (%v, %d), want (-0, 2)", key, val)
		}
	}
}

func TestTransientDeleteMissingNoOp(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[int, int]().Set(1, 1)
	d := r.TransientDelete(owner, 99)
	if d.node != r.node || d.size != r.size {
		t.Fatal("missing TransientDelete must retain node and size")
	}
	owner.Invalidate()
}

func assertNotOwnedByInactive[K comparable, V any](t *testing.T, n node[K, V], inactive *ownership.Owner) {
	t.Helper()
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *leaf[K, V]:
		if ownership.IsOwnedBy(x.owner, inactive) {
			t.Fatal("leaf still IsOwnedBy inactive owner")
		}
	case *collision[K, V]:
		if ownership.IsOwnedBy(x.owner, inactive) {
			t.Fatal("collision still IsOwnedBy inactive owner")
		}
	case *bitmapNode[K, V]:
		if ownership.IsOwnedBy(x.owner, inactive) {
			t.Fatal("bitmap still IsOwnedBy inactive owner")
		}
		for _, ch := range x.children {
			assertNotOwnedByInactive(t, ch, inactive)
		}
	default:
		t.Fatalf("unknown node %T", n)
	}
}
