package vector

import (
	"fmt"
	"testing"
)

func TestEmptyZeroValueAndLen(t *testing.T) {
	t.Parallel()

	var z Root[int]
	if z.Len() != 0 {
		t.Fatalf("zero Len = %d", z.Len())
	}
	e := Empty[int]()
	if e.Len() != 0 {
		t.Fatalf("Empty Len = %d", e.Len())
	}
	if e.root != nil || e.tail != nil || e.count != 0 {
		t.Fatal("Empty must match zero-value shape")
	}
}

func TestAppendAtSetLen(t *testing.T) {
	t.Parallel()

	r := Empty[string]()
	r1 := r.Append("a")
	r2 := r1.Append("b", "c")

	if r.Len() != 0 || r1.Len() != 1 || r2.Len() != 3 {
		t.Fatalf("lens: empty=%d r1=%d r2=%d", r.Len(), r1.Len(), r2.Len())
	}
	if g := r2.At(0); g != "a" {
		t.Fatalf("At(0)=%q", g)
	}
	if g := r2.At(1); g != "b" {
		t.Fatalf("At(1)=%q", g)
	}
	if g := r2.At(2); g != "c" {
		t.Fatalf("At(2)=%q", g)
	}

	// Prior versions intact.
	if r1.Len() != 1 || r1.At(0) != "a" {
		t.Fatal("r1 mutated by later Append")
	}

	set := r2.Set(1, "B")
	if set.At(1) != "B" {
		t.Fatalf("Set At(1)=%q", set.At(1))
	}
	if r2.At(1) != "b" {
		t.Fatal("Set mutated prior version")
	}
	if set.Len() != 3 {
		t.Fatalf("Set changed Len: %d", set.Len())
	}
}

func TestDeleteMidPreservesOrder(t *testing.T) {
	t.Parallel()

	r := Empty[int]().Append(0, 1, 2, 3, 4)
	d := r.Delete(2)
	if d.Len() != 4 {
		t.Fatalf("Len after Delete = %d", d.Len())
	}
	want := []int{0, 1, 3, 4}
	for i, w := range want {
		if g := d.At(i); g != w {
			t.Fatalf("At(%d)=%d want %d", i, g, w)
		}
	}
	// Prior version intact.
	if r.Len() != 5 || r.At(2) != 2 {
		t.Fatal("Delete mutated prior version")
	}

	// Delete head and tail.
	dh := r.Delete(0)
	if dh.Len() != 4 || dh.At(0) != 1 || dh.At(3) != 4 {
		t.Fatalf("delete head: len=%d first=%d last=%d", dh.Len(), dh.At(0), dh.At(3))
	}
	dt := r.Delete(4)
	if dt.Len() != 4 || dt.At(3) != 3 {
		t.Fatalf("delete tail: len=%d last=%d", dt.Len(), dt.At(3))
	}

	// Delete sole element yields empty.
	one := Empty[int]().Append(42)
	if got := one.Delete(0); got.Len() != 0 {
		t.Fatalf("delete sole: Len=%d", got.Len())
	}
}

func TestBoundsPanics(t *testing.T) {
	t.Parallel()

	r := Empty[int]().Append(1, 2, 3)
	for _, tc := range []struct {
		name string
		fn   func()
	}{
		{"At(-1)", func() { _ = r.At(-1) }},
		{"At(len)", func() { _ = r.At(3) }},
		{"Set(-1)", func() { _ = r.Set(-1, 0) }},
		{"Set(len)", func() { _ = r.Set(3, 0) }},
		{"Delete(-1)", func() { _ = r.Delete(-1) }},
		{"Delete(len)", func() { _ = r.Delete(3) }},
		{"At empty", func() { _ = Empty[int]().At(0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				rec := recover()
				if rec == nil {
					t.Fatal("expected panic")
				}
				msg, ok := rec.(string)
				if !ok || msg != "vector: index out of range" {
					t.Fatalf("panic = %#v", rec)
				}
			}()
			tc.fn()
		})
	}
}

func TestStructuralSharingOnSet(t *testing.T) {
	t.Parallel()

	// 65 elements → trie holds two full leaves (0..31 and 32..63) and the
	// tail holds index 64. That gives the root two sibling children so Set
	// on the first leaf can prove the second child pointer is shared.
	r := Empty[int]()
	for i := 0; i < 65; i++ {
		r = r.Append(i)
	}
	old := r

	// Mutate an element in the first leaf (index 0).
	next := r.Set(0, 999)

	if old.root == next.root {
		t.Fatal("root must be path-copied on Set into the trie")
	}
	if old.root == nil || next.root == nil {
		t.Fatal("expected non-nil roots at n=65")
	}
	if len(old.root.children) < 2 || len(next.root.children) < 2 {
		t.Fatalf("expected >=2 root children, got old=%d next=%d",
			len(old.root.children), len(next.root.children))
	}

	// Untouched sibling leaf (second child) must be the same pointer.
	if old.root.children[1] != next.root.children[1] {
		t.Fatal("untouched sibling leaf must be shared by pointer")
	}
	// Edited path child must differ.
	if old.root.children[0] == next.root.children[0] {
		t.Fatal("edited leaf path must be path-copied")
	}

	if old.At(0) != 0 {
		t.Fatal("old version mutated")
	}
	if next.At(0) != 999 {
		t.Fatalf("next At(0)=%d", next.At(0))
	}
	if next.At(33) != 33 || old.At(33) != 33 {
		t.Fatal("sibling region values broken")
	}
}

func TestStructuralSharingOnSetDeep(t *testing.T) {
	t.Parallel()

	// Force a deeper tree (shift > 5): more than 32 leaves under the root.
	// 32 leaves * 32 values = 1024 in the trie; use 1100 to be past overflow.
	const n = 1100
	r := Empty[int]()
	for i := 0; i < n; i++ {
		r = r.Append(i)
	}
	if r.shift < 2*bitsPerLevel {
		t.Fatalf("expected deeper tree, shift=%d", r.shift)
	}

	old := r
	// Edit near the end of the trie region (not in the tail).
	idx := r.tailoff() - 1
	next := r.Set(idx, -1)

	if old.root == next.root {
		t.Fatal("deep Set must path-copy root")
	}
	// At least one top-level child should remain shared.
	shared := 0
	for i := 0; i < len(old.root.children) && i < len(next.root.children); i++ {
		if old.root.children[i] == next.root.children[i] {
			shared++
		}
	}
	if shared == 0 {
		t.Fatal("expected at least one shared top-level child after deep Set")
	}
	if old.At(idx) == -1 {
		t.Fatal("old version mutated")
	}
	if next.At(idx) != -1 {
		t.Fatalf("next At(%d)=%d", idx, next.At(idx))
	}
}

func TestLargeAppendAtSet(t *testing.T) {
	t.Parallel()

	const n = 2000
	r := Empty[int]()
	for i := 0; i < n; i++ {
		r = r.Append(i * 10)
	}
	if r.Len() != n {
		t.Fatalf("Len=%d", r.Len())
	}

	// Mid access.
	mid := n / 2
	if g := r.At(mid); g != mid*10 {
		t.Fatalf("At(%d)=%d want %d", mid, g, mid*10)
	}
	if g := r.At(0); g != 0 {
		t.Fatalf("At(0)=%d", g)
	}
	if g := r.At(n - 1); g != (n-1)*10 {
		t.Fatalf("At(last)=%d", g)
	}

	// Set mid and verify.
	s := r.Set(mid, 7)
	if s.At(mid) != 7 {
		t.Fatalf("Set mid = %d", s.At(mid))
	}
	if r.At(mid) != mid*10 {
		t.Fatal("Set mutated prior large vector")
	}
	// Spot-check neighbors.
	if s.At(mid-1) != (mid-1)*10 || s.At(mid+1) != (mid+1)*10 {
		t.Fatal("neighbors corrupted by Set")
	}

	// Delete mid preserves order for a sample window.
	d := r.Delete(mid)
	if d.Len() != n-1 {
		t.Fatalf("Delete Len=%d", d.Len())
	}
	if d.At(mid-1) != (mid-1)*10 {
		t.Fatalf("before hole: %d", d.At(mid-1))
	}
	if d.At(mid) != (mid+1)*10 {
		t.Fatalf("after shift: got %d want %d", d.At(mid), (mid+1)*10)
	}
}

func TestAllIteration(t *testing.T) {
	t.Parallel()

	r := Empty[string]().Append("x", "y", "z")
	var got []string
	var idxs []int
	for i, v := range r.All() {
		idxs = append(idxs, i)
		got = append(got, v)
	}
	if fmt.Sprint(idxs) != "[0 1 2]" {
		t.Fatalf("idxs=%v", idxs)
	}
	if fmt.Sprint(got) != "[x y z]" {
		t.Fatalf("got=%v", got)
	}

	// Early stop.
	count := 0
	for range r.All() {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("early stop count=%d", count)
	}
}

func TestAppendAcrossTailPushAndLevelOverflow(t *testing.T) {
	t.Parallel()

	// Cross the first tail→trie push (32) and root overflow (1024+32).
	r := Empty[int]()
	for i := 0; i < 1100; i++ {
		r = r.Append(i)
	}
	for _, i := range []int{0, 31, 32, 33, 1023, 1024, 1099} {
		if g := r.At(i); g != i {
			t.Fatalf("At(%d)=%d", i, g)
		}
	}
	// Set in tail region and trie region.
	tailIdx := r.tailoff() + 1
	if tailIdx >= r.Len() {
		tailIdx = r.Len() - 1
	}
	s1 := r.Set(tailIdx, -tailIdx)
	s2 := r.Set(100, -100)
	if s1.At(tailIdx) != -tailIdx || r.At(tailIdx) != tailIdx {
		t.Fatal("tail Set broken")
	}
	if s2.At(100) != -100 || r.At(100) != 100 {
		t.Fatal("trie Set broken")
	}
}

func TestPublishedNodesHaveNilOwner(t *testing.T) {
	t.Parallel()

	r := Empty[int]()
	for i := 0; i < 40; i++ {
		r = r.Append(i)
	}
	if r.root == nil {
		t.Fatal("expected trie root after 40 appends")
	}
	if r.root.owner != nil {
		t.Fatal("published root owner must be nil")
	}
	for _, ch := range r.root.children {
		if ch == nil {
			continue
		}
		if ch.owner != nil {
			t.Fatal("published child owner must be nil")
		}
	}

	s := r.Set(0, 1)
	if s.root.owner != nil || s.root.children[0].owner != nil {
		t.Fatal("path-copied Set nodes must publish with nil owner")
	}
}
