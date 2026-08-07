package hamt

import "testing"

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
