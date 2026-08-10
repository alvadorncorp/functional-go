package hamt

import (
	"maps"
	"testing"
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
	for i := range n {
		r = r.Set(i, i*i)
	}
	if r.Len() != n {
		t.Fatalf("Len = %d want %d", r.Len(), n)
	}
	for i := range n {
		assertGet(t, r, i, i*i)
	}
	// Delete half.
	for i := 0; i < n; i += 2 {
		r = r.Delete(i)
	}
	if r.Len() != n/2 {
		t.Fatalf("after half delete Len = %d", r.Len())
	}
	for i := range n {
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
	got := maps.Collect(r.All())
	if len(got) != 3 || got["a"] != 1 || got["b"] != 2 || got["c"] != 3 {
		t.Fatalf("All = %#v", got)
	}
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
