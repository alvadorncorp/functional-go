package vector

import (
	"fmt"
	"testing"
)

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

func TestLargeAppendAtSet(t *testing.T) {
	t.Parallel()

	const n = 2000
	r := Empty[int]()
	for i := range n {
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
	for i := range 1100 {
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
