package hamt

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
	"github.com/alvadorncorp/gopackages/functional/internal/engine/keyhash"
)

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

func TestOwnerNilOnPublishedNodes(t *testing.T) {
	t.Parallel()

	r := Empty[int, int]()
	for i := range 40 {
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
	for range 64 {
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
				for s := range keyLimit {
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

func TestDeletePromotesTerminalThroughSharedPrefix(t *testing.T) {
	t.Parallel()

	for _, terminal := range []string{"leaf", "collision"} {
		for _, mode := range []string{"pure", "transient-published", "transient-owned", "transient-frozen"} {
			t.Run(terminal+"/"+mode, func(t *testing.T) {
				// Diverge only in the last hash fragment, leaving twelve unary
				// bitmaps above the branch that disappears on deletion.
				const otherHash = uint64(1) << maxShift
				var base node[int, int] = &leaf[int, int]{hash: 0, key: 0, val: 10}
				want := map[int]int{0: 10, 1: 11}
				if terminal == "collision" {
					base, _ = base.set(0, 0, 4, 40)
					want[4] = 40
				}
				base, _ = base.set(otherHash, 0, 1, 11)
				current := base
				owner := ownership.New()
				defer owner.Invalidate()
				frozen := owner
				if mode == "transient-owned" || mode == "transient-frozen" {
					current, _ = current.tset(owner, 0, 0, 0, 10)
				}
				if mode == "transient-frozen" {
					owner.Invalidate()
					owner = ownership.New()
					defer owner.Invalidate()
				}
				keep := current
				for {
					b, ok := keep.(*bitmapNode[int, int])
					if !ok {
						break
					}
					keep = b.children[0]
				}
				var removed bool
				if mode == "pure" {
					current, removed = current.del(otherHash, 0, 1)
				} else {
					current, removed = current.tdel(owner, otherHash, 0, 1)
				}
				if !removed || current != keep {
					t.Fatalf("delete = (%T, %v), want original terminal %T", current, removed, keep)
				}
				switch n := current.(type) {
				case *leaf[int, int]:
					if terminal != "leaf" {
						t.Fatal("collision collapsed unexpectedly")
					}
					if mode == "transient-owned" && !ownership.IsOwnedBy(n.owner, owner) {
						t.Fatal("promoted leaf lost active ownership")
					}
				case *collision[int, int]:
					if terminal != "collision" {
						t.Fatal("leaf became a collision")
					}
					if mode == "transient-owned" && !ownership.IsOwnedBy(n.owner, owner) {
						t.Fatal("promoted collision lost active ownership")
					}
				default:
					t.Fatalf("terminal still wrapped in %T", current)
				}
				switch mode {
				case "pure":
					assertOwnersNil(t, current)
				case "transient-frozen":
					assertNotOwnedByInactive(t, current, frozen)
				}
				assertTerminalCompressionContents(t, base, want)
				delete(want, 1)
				assertTerminalCompressionContents(t, current, want)

				// Publishing and editing the promoted terminal under a new owner
				// must preserve both the original and compressed snapshots.
				owner.Invalidate()
				fresh := ownership.New()
				defer fresh.Invalidate()
				next, added := current.tset(fresh, 0, 0, 0, 99)
				if added {
					t.Fatal("replacement changed entry count")
				}
				assertTerminalCompressionContents(t, current, want)
				want[0] = 99
				assertTerminalCompressionContents(t, next, want)
				want[0], want[1] = 10, 11
				assertTerminalCompressionContents(t, base, want)
			})
		}
	}
}

// TestDeleteKeepsBitmapChildBelowUnaryParent guards the del/tdel branch where a
// unary parent's child is replaced by a surviving *bitmap* (not a terminal): it
// must stay wrapped because that bitmap's indexes are relative to shift 5.
func TestDeleteKeepsBitmapChildBelowUnaryParent(t *testing.T) {
	t.Parallel()

	// Hashes 0, 32 and 64 share the shift-0 index, so the root is unary; at
	// shift 5 they take indexes 0, 1 and 2 in one three-leaf bitmap.
	hashes := map[int]uint64{0: 0, 1: 32, 2: 64}
	for _, mode := range []string{"pure", "transient"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			var base node[int, int] = &leaf[int, int]{hash: hashes[0], key: 0, val: 10}
			base, _ = base.set(hashes[1], 0, 1, 11)
			base, _ = base.set(hashes[2], 0, 2, 12)
			rootBM, ok := base.(*bitmapNode[int, int])
			if !ok || len(rootBM.children) != 1 {
				t.Fatalf("setup root = %T children=%d, want unary bitmap", base, childrenLen(base))
			}
			setupInner, ok := rootBM.children[0].(*bitmapNode[int, int])
			if !ok || setupInner.bitmap != 0b111 || len(setupInner.children) != 3 {
				t.Fatalf("setup child = %T, want three-leaf bitmap 0b111", rootBM.children[0])
			}

			var got node[int, int]
			var removed bool
			switch mode {
			case "pure":
				got, removed = base.del(hashes[2], 0, 2)
			case "transient":
				owner := ownership.New()
				defer owner.Invalidate()
				got, removed = base.tdel(owner, hashes[2], 0, 2)
			}
			if !removed {
				t.Fatal("delete reported no removal")
			}

			// Consumer-visible behavior first: a promoted shift-5 bitmap would
			// lose lookups starting at shift 0.
			assertUnaryBitmapContents(t, got, hashes, map[int]int{0: 10, 1: 11})
			assertUnaryBitmapContents(t, base, hashes, map[int]int{0: 10, 1: 11, 2: 12})

			root, ok := got.(*bitmapNode[int, int])
			if !ok || root.bitmap != bitpos(0) || len(root.children) != 1 {
				t.Fatalf("root = %T, want unary bitmap at index 0", got)
			}
			inner, ok := root.children[0].(*bitmapNode[int, int])
			if !ok || inner.bitmap != 0b11 || len(inner.children) != 2 {
				t.Fatalf("root child = %T, want two-leaf bitmap 0b11 (not promoted)", root.children[0])
			}
		})
	}
}

func assertUnaryBitmapContents(t *testing.T, n node[int, int], hashes map[int]uint64, want map[int]int) {
	t.Helper()
	for key, hash := range hashes {
		got, found := n.get(hash, 0, key)
		value, present := want[key]
		if got != value || found != present {
			t.Fatalf("get(%d) = (%d, %v), want (%d, %v)", key, got, found, value, present)
		}
	}
	count := 0
	n.forEach(func(key, value int) bool {
		if expected, ok := want[key]; !ok || value != expected {
			t.Fatalf("unexpected entry (%d, %d)", key, value)
		}
		count++
		return true
	})
	if count != len(want) {
		t.Fatalf("entry count = %d, want %d", count, len(want))
	}
}

func assertTerminalCompressionContents(t *testing.T, n node[int, int], want map[int]int) {
	t.Helper()
	for _, key := range []int{0, 1, 4, 8} {
		var hash uint64
		if key == 1 {
			hash = uint64(1) << maxShift
		}
		got, found := n.get(hash, 0, key)
		value, present := want[key]
		if got != value || found != present {
			t.Fatalf("get(%d) = (%d, %v), want (%d, %v)", key, got, found, value, present)
		}
	}
	count := 0
	seen := make(map[int]bool)
	n.forEach(func(key, value int) bool {
		if expected, ok := want[key]; !ok || value != expected || seen[key] {
			t.Fatalf("unexpected or duplicate entry (%d, %d)", key, value)
		}
		seen[key] = true
		count++
		return true
	})
	if count != len(want) {
		t.Fatalf("entry count = %d, want %d", count, len(want))
	}
}

// Non-goal: V any is not comparable; Set of an existing key always may
// allocate a new path even when the new value equals the old one.
// Do not require root pointer identity on value-equal replace.
func TestSetEqualValueReplaceDoesNotRequireRootIdentity(t *testing.T) {
	t.Parallel()

	r := Empty[string, int]().Set("k", 42)
	next := r.Set("k", 42)

	if next.Len() != 1 {
		t.Fatalf("Len = %d, want 1", next.Len())
	}
	assertGet(t, next, "k", 42)
	assertGet(t, r, "k", 42)

	// Root pointer identity is intentionally NOT asserted. A path-copying
	// implementation may allocate a new path even when the value is equal,
	// because V is not constrained to be comparable and no equality check
	// is required on the Set hot path. Structural identity on equal-value
	// replace is not a requirement of this engine.
}

func childrenLen[K comparable, V any](n node[K, V]) int {
	if bm, ok := n.(*bitmapNode[K, V]); ok {
		return len(bm.children)
	}
	return -1
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
