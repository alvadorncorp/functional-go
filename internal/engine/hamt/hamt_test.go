package hamt

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/engine/keyhash"
)

func TestBasicSetGetDeleteLen(t *testing.T) {
	t.Parallel()

	r := Empty[string, int]()
	if r.Len() != 0 {
		t.Fatalf("empty Len = %d", r.Len())
	}
	if _, ok := r.Get("missing"); ok {
		t.Fatal("Get on empty must miss")
	}

	r1 := r.Set("a", 1)
	r2 := r1.Set("b", 2)
	r3 := r2.Set("c", 3)

	if r.Len() != 0 || r1.Len() != 1 || r2.Len() != 2 || r3.Len() != 3 {
		t.Fatalf("lens: empty=%d r1=%d r2=%d r3=%d", r.Len(), r1.Len(), r2.Len(), r3.Len())
	}
	assertGet(t, r3, "a", 1)
	assertGet(t, r3, "b", 2)
	assertGet(t, r3, "c", 3)

	// Original empty and intermediate versions stay intact.
	if _, ok := r.Get("a"); ok {
		t.Fatal("empty root mutated")
	}
	if _, ok := r1.Get("b"); ok {
		t.Fatal("r1 should not contain b")
	}

	d := r3.Delete("b")
	if d.Len() != 2 {
		t.Fatalf("after delete Len = %d", d.Len())
	}
	if _, ok := d.Get("b"); ok {
		t.Fatal("deleted key still present")
	}
	assertGet(t, d, "a", 1)
	assertGet(t, d, "c", 3)
	assertGet(t, r3, "b", 2) // prior version intact
}

func TestReplaceDoesNotGrow(t *testing.T) {
	t.Parallel()

	r := Empty[int, string]().Set(1, "one").Set(2, "two")
	if r.Len() != 2 {
		t.Fatalf("Len = %d", r.Len())
	}
	r2 := r.Set(1, "ONE")
	if r2.Len() != 2 {
		t.Fatalf("replace grew size: %d", r2.Len())
	}
	assertGet(t, r2, 1, "ONE")
	assertGet(t, r, 1, "one")
}

func TestDeleteMissingNoOp(t *testing.T) {
	t.Parallel()

	r := Empty[int, int]().Set(1, 1).Set(2, 2)
	d := r.Delete(99)
	if d.Len() != r.Len() {
		t.Fatalf("size changed on missing delete: %d vs %d", d.Len(), r.Len())
	}
	// Prefer structural identity for no-op Delete.
	if d.node != r.node {
		t.Fatal("no-op Delete should retain the same root node pointer")
	}
	if d.size != r.size {
		t.Fatal("no-op Delete should retain size")
	}
}

func TestSeedPreservedAcrossSetDelete(t *testing.T) {
	t.Parallel()

	r := Empty[string, int]()
	seed0 := r.seed

	r = r.Set("x", 1).Set("y", 2).Set("z", 3)
	r = r.Delete("y").Set("w", 4)
	// Empty-out then reinsert still keeps the lineage seed.
	r = r.Delete("x").Delete("z").Delete("w")
	if r.Len() != 0 || r.node != nil {
		t.Fatal("expected fully empty root after deleting all keys")
	}
	if r.seed != seed0 {
		t.Fatal("seed must be preserved across Set/Delete chain, including empty")
	}
	r = r.Set("again", 9)
	if r.seed != seed0 {
		t.Fatal("seed must remain after reinsert into emptied root")
	}
	// Hash under preserved seed is stable for the same key.
	h1 := keyhash.Hash(seed0, "probe")
	h2 := keyhash.Hash(r.seed, "probe")
	if h1 != h2 {
		t.Fatal("preserved seed must hash identically")
	}
}

func TestIndependentEmptyRootsHaveOwnSeeds(t *testing.T) {
	t.Parallel()

	// Two Empty() calls must not be forced to share a seed; otherwise
	// independent maps could not be hashed under separate seeds.
	a := Empty[int, int]()
	b := Empty[int, int]()
	// Extremely unlikely to collide if MakeSeed is working; retry a few times.
	differ := false
	for i := 0; i < 8 && !differ; i++ {
		if a.seed != b.seed {
			differ = true
			break
		}
		b = Empty[int, int]()
	}
	if !differ {
		t.Fatal("expected independent Empty roots to obtain distinct seeds")
	}
}

func TestStructuralSharingUntouchedSubtree(t *testing.T) {
	t.Parallel()

	// Populate enough keys that the root is a multi-child bitmap node.
	r := Empty[int, int]()
	for i := 0; i < 64; i++ {
		r = r.Set(i, i)
	}
	old := r

	// Mutate one key that is almost certainly on a single path.
	next := r.Set(1000, 1000)

	ob, ok1 := old.node.(*bitmapNode[int, int])
	nb, ok2 := next.node.(*bitmapNode[int, int])
	if !ok1 || !ok2 {
		t.Fatalf("expected bitmap roots, got %T and %T", old.node, next.node)
	}
	if ob == nb {
		t.Fatal("root bitmap node must be path-copied on Set")
	}

	shared := 0
	for idx := uint(0); idx < 32; idx++ {
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
		t.Fatal("expected at least one shared child pointer between old and new root")
	}

	// Values from the untouched region remain readable on both versions.
	assertGet(t, old, 0, 0)
	assertGet(t, next, 0, 0)
	assertGet(t, next, 1000, 1000)
	if _, ok := old.Get(1000); ok {
		t.Fatal("old root must not contain the new key")
	}
}

func TestHashCollisionsSynthetic(t *testing.T) {
	t.Parallel()

	// Force a collision bucket by constructing leaves with equal hashes and
	// driving them through set. Real maphash collisions are rare; the engine
	// path is what we need to prove.
	const h uint64 = 0xdeadbeefcafebabe
	l := &leaf[string, int]{hash: h, key: "alpha", val: 1}
	n, added := l.set(h, 0, "beta", 2)
	if !added {
		t.Fatal("second key must count as insert")
	}
	c, ok := n.(*collision[string, int])
	if !ok {
		t.Fatalf("expected collision node, got %T", n)
	}
	if len(c.entries) != 2 {
		t.Fatalf("entries = %d", len(c.entries))
	}

	if v, found := n.get(h, 0, "alpha"); !found || v != 1 {
		t.Fatalf("alpha: got %v,%v", v, found)
	}
	if v, found := n.get(h, 0, "beta"); !found || v != 2 {
		t.Fatalf("beta: got %v,%v", v, found)
	}

	n2, added := n.set(h, 0, "gamma", 3)
	if !added {
		t.Fatal("gamma must be insert")
	}
	n3, added := n2.set(h, 0, "beta", 20)
	if added {
		t.Fatal("replace must not add")
	}
	if v, found := n3.get(h, 0, "beta"); !found || v != 20 {
		t.Fatalf("beta replace: got %v,%v", v, found)
	}

	// Delete down to a leaf.
	n4, removed := n3.del(h, 0, "alpha")
	if !removed {
		t.Fatal("alpha delete")
	}
	n5, removed := n4.del(h, 0, "gamma")
	if !removed {
		t.Fatal("gamma delete")
	}
	lf, ok := n5.(*leaf[string, int])
	if !ok || lf.key != "beta" || lf.val != 20 {
		t.Fatalf("expected collapsed leaf beta=20, got %T %#v", n5, n5)
	}

	// Real maphash collisions are rare; Root.Get/Set exercise the same node
	// methods via keyhash.Hash. The synthetic path above is the reliable
	// coverage for multi-entry collision buckets.
}

func TestManyKeysRoundTrip(t *testing.T) {
	t.Parallel()

	const n = 200
	r := Empty[int, int]()
	for i := 0; i < n; i++ {
		r = r.Set(i, i*i)
	}
	if r.Len() != n {
		t.Fatalf("Len = %d want %d", r.Len(), n)
	}
	for i := 0; i < n; i++ {
		assertGet(t, r, i, i*i)
	}
	// Delete half.
	for i := 0; i < n; i += 2 {
		r = r.Delete(i)
	}
	if r.Len() != n/2 {
		t.Fatalf("after half delete Len = %d", r.Len())
	}
	for i := 0; i < n; i++ {
		v, ok := r.Get(i)
		if i%2 == 0 {
			if ok {
				t.Fatalf("key %d should be deleted", i)
			}
			continue
		}
		if !ok || v != i*i {
			t.Fatalf("key %d: got %v,%v", i, v, ok)
		}
	}
}

func TestAllIterator(t *testing.T) {
	t.Parallel()

	r := Empty[string, int]().Set("a", 1).Set("b", 2).Set("c", 3)
	got := map[string]int{}
	for k, v := range r.All() {
		got[k] = v
	}
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("All = %#v", got)
	}
}

func TestOwnerNilOnPublishedNodes(t *testing.T) {
	t.Parallel()

	r := Empty[int, int]()
	for i := 0; i < 40; i++ {
		r = r.Set(i, i)
	}
	assertOwnersNil(t, r.node)
}

// TestPathCompressionDoesNotElevateBitmap guards the shift-relative bitmap
// invariant: when a two-child node loses one child and the survivor is an
// internal bitmap, it must stay wrapped so Root.Get can still start at shift 0.
func TestPathCompressionDoesNotElevateBitmap(t *testing.T) {
	t.Parallel()

	// Force a nested two-child shape: nestA/nestB share the root 5-bit index
	// (so they live under an internal bitmap), while sibling sits in a different
	// root slot as a leaf. Deleting sibling must NOT elevate the nested bitmap.
	const keyLimit = 2000
	for trial := 0; trial < 64; trial++ {
		r0 := Empty[int, int]()
		seed := r0.seed

		var nestA, nestB, sibling int
		var found bool
		for a := 0; a < keyLimit && !found; a++ {
			ha := keyhash.Hash(seed, a)
			ia := ha & levelMask
			for b := a + 1; b < keyLimit && !found; b++ {
				hb := keyhash.Hash(seed, b)
				if hb&levelMask != ia || ha == hb {
					continue
				}
				for s := 0; s < keyLimit; s++ {
					if s == a || s == b {
						continue
					}
					if keyhash.Hash(seed, s)&levelMask != ia {
						nestA, nestB, sibling = a, b, s
						found = true
						break
					}
				}
			}
		}
		if !found {
			continue
		}

		r := r0.Set(nestA, 1).Set(nestB, 2).Set(sibling, 3)

		// Pre-delete shape: root bitmap with exactly two children — one nested
		// bitmap (the nestA/nestB branch) and one leaf (sibling).
		rootBM, ok := r.node.(*bitmapNode[int, int])
		if !ok || len(rootBM.children) != 2 {
			t.Fatalf("pre-delete root: want 2-child bitmap, got %T children=%v",
				r.node, childrenLen(r.node))
		}
		var nestedCount, leafCount int
		for _, ch := range rootBM.children {
			switch ch.(type) {
			case *bitmapNode[int, int]:
				nestedCount++
			case *leaf[int, int]:
				leafCount++
			default:
				t.Fatalf("pre-delete unexpected child type %T", ch)
			}
		}
		if nestedCount != 1 || leafCount != 1 {
			t.Fatalf("pre-delete shape: nested=%d leaf=%d want 1 each", nestedCount, leafCount)
		}

		d := r.Delete(sibling)
		if d.Len() != 2 {
			t.Fatalf("Len after delete sibling = %d", d.Len())
		}
		assertGet(t, d, nestA, 1)
		assertGet(t, d, nestB, 2)
		if _, present := d.Get(sibling); present {
			t.Fatal("sibling still present after Delete")
		}

		// Remaining root must still be a single-child *bitmapNode wrapping the
		// nested *bitmapNode. Elevating the nested bitmap to root would keep
		// Get working only by accident for some keys; the shape is the real
		// regression signal.
		bm, ok := d.node.(*bitmapNode[int, int])
		if !ok {
			t.Fatalf("post-delete root elevated away from bitmap: got %T", d.node)
		}
		if len(bm.children) != 1 {
			t.Fatalf("post-delete root children = %d, want 1", len(bm.children))
		}
		if _, isBM := bm.children[0].(*bitmapNode[int, int]); !isBM {
			t.Fatalf("post-delete nested child elevated: got %T, want *bitmapNode", bm.children[0])
		}
		return
	}
	t.Fatal("could not find nestA/nestB/sibling key triple under any seed")
}

func childrenLen[K comparable, V any](n node[K, V]) int {
	if bm, ok := n.(*bitmapNode[K, V]); ok {
		return len(bm.children)
	}
	return -1
}

func assertGet[K, V comparable](t *testing.T, r Root[K, V], key K, want V) {
	t.Helper()
	got, ok := r.Get(key)
	if !ok {
		t.Fatalf("Get(%v) missing", key)
	}
	if got != want {
		t.Fatalf("Get(%v) = %v, want %v", key, got, want)
	}
}

func assertOwnersNil[K comparable, V any](t *testing.T, n node[K, V]) {
	t.Helper()
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *leaf[K, V]:
		if x.owner != nil {
			t.Fatal("leaf owner must be nil on pure immutable path")
		}
	case *collision[K, V]:
		if x.owner != nil {
			t.Fatal("collision owner must be nil on pure immutable path")
		}
	case *bitmapNode[K, V]:
		if x.owner != nil {
			t.Fatal("bitmap owner must be nil on pure immutable path")
		}
		for _, ch := range x.children {
			assertOwnersNil(t, ch)
		}
	default:
		t.Fatalf("unknown node %T", n)
	}
}
