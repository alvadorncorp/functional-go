package vector

import (
	"fmt"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

func TestPopEmptyPanics(t *testing.T) {
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
	_, _ = Empty[int]().Pop()
}

func TestPopSoleElement(t *testing.T) {
	t.Parallel()

	r := Empty[int]().Append(42)
	next, val := r.Pop()
	if val != 42 {
		t.Fatalf("val = %d, want 42", val)
	}
	if next.Len() != 0 {
		t.Fatalf("Len = %d, want 0", next.Len())
	}
	if r.Len() != 1 || r.At(0) != 42 {
		t.Fatal("prior sole-element version mutated")
	}
}

func TestPopTailRoomPreservesPrior(t *testing.T) {
	t.Parallel()

	// n=5: all in tail, O(1) shrink path.
	r := Empty[int]().Append(0, 1, 2, 3, 4)
	next, val := r.Pop()
	if val != 4 {
		t.Fatalf("val = %d, want 4", val)
	}
	if next.Len() != 4 {
		t.Fatalf("Len = %d, want 4", next.Len())
	}
	for i := 0; i < 4; i++ {
		if g := next.At(i); g != i {
			t.Fatalf("next At(%d)=%d", i, g)
		}
	}
	// Prior version intact with original Len.
	if r.Len() != 5 {
		t.Fatalf("prior Len = %d", r.Len())
	}
	for i := 0; i < 5; i++ {
		if g := r.At(i); g != i {
			t.Fatalf("prior At(%d)=%d", i, g)
		}
	}
	// Root pointer shared on pure tail shrink.
	if r.root != next.root {
		t.Fatal("tail-only Pop should keep the same root pointer")
	}
}

func TestPopPullsLeafFromTrie(t *testing.T) {
	t.Parallel()

	for _, n := range []int{33, 34, 65} {
		r := Empty[int]()
		for i := 0; i < n; i++ {
			r = r.Append(i)
		}
		prior := r
		next, val := r.Pop()
		if val != n-1 {
			t.Fatalf("n=%d: val = %d, want %d", n, val, n-1)
		}
		if next.Len() != n-1 {
			t.Fatalf("n=%d: Len = %d", n, next.Len())
		}
		for i := 0; i < n-1; i++ {
			if g := next.At(i); g != i {
				t.Fatalf("n=%d: next At(%d)=%d", n, i, g)
			}
		}
		if prior.Len() != n || prior.At(n-1) != n-1 {
			t.Fatalf("n=%d: prior version broken", n)
		}
	}
}

func TestPopRoundTripBeyondLevelOverflow(t *testing.T) {
	t.Parallel()

	// Root overflow (shift 5→10) happens when the 33rd leaf is pushed: n≥1057.
	// Sizes cover the overflow edge, a pull boundary (tail len==1), and larger
	// deep trees.
	for _, n := range []int{1057, 1089, 1100, 2048} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()

			r := Empty[int]()
			for i := 0; i < n; i++ {
				r = r.Append(i)
			}
			if r.Len() != n {
				t.Fatalf("build Len = %d", r.Len())
			}
			if r.shift < 2*bitsPerLevel {
				t.Fatalf("precondition: expected shift>=%d after overflow, got %d",
					2*bitsPerLevel, r.shift)
			}

			// Full Pop round-trip of values 0..n-1; also watch for shift
			// compression once the deep root collapses to a single child.
			cur := r
			sawCompress := false
			for want := n - 1; want >= 0; want-- {
				prevShift := cur.shift
				var val int
				cur, val = cur.Pop()
				if val != want {
					t.Fatalf("Pop val = %d, want %d", val, want)
				}
				if cur.Len() != want {
					t.Fatalf("Len after Pop = %d, want %d", cur.Len(), want)
				}
				if prevShift >= 2*bitsPerLevel && cur.shift < prevShift {
					sawCompress = true
					// Compression only when the remaining root has one child
					// (or the trie is emptied entirely at shift→0).
					if cur.root != nil && cur.shift > bitsPerLevel && len(cur.root.children) != 1 {
						t.Fatalf("shift compressed to %d but root children=%d",
							cur.shift, len(cur.root.children))
					}
				}
			}
			if cur.Len() != 0 {
				t.Fatalf("final Len = %d", cur.Len())
			}
			if !sawCompress {
				t.Fatal("expected shift compression while draining past overflow")
			}

			// Prior full version still readable at overflow depth.
			if r.Len() != n {
				t.Fatalf("prior Len = %d", r.Len())
			}
			if r.shift < 2*bitsPerLevel {
				t.Fatalf("prior shift mutated to %d", r.shift)
			}
			for _, i := range []int{0, 31, 32, 1023, 1056, n - 1} {
				if i >= n {
					continue
				}
				if g := r.At(i); g != i {
					t.Fatalf("prior At(%d)=%d", i, g)
				}
			}

			// Round-trip rebuild matches the preserved prior.
			rebuilt := Empty[int]()
			for i := 0; i < n; i++ {
				rebuilt = rebuilt.Append(i)
			}
			if rebuilt.shift != r.shift {
				t.Fatalf("rebuild shift = %d, want %d", rebuilt.shift, r.shift)
			}
			for i := 0; i < n; i++ {
				if rebuilt.At(i) != r.At(i) {
					t.Fatalf("rebuild mismatch at %d", i)
				}
			}
		})
	}
}

func TestPopCompressesShiftAtOverflowBoundary(t *testing.T) {
	t.Parallel()

	// n=1057: shift=10 with two root children; the first Pop pulls the lone
	// overflow leaf and must compress shift back to 5.
	const n = 1057
	r := Empty[int]()
	for i := 0; i < n; i++ {
		r = r.Append(i)
	}
	if r.shift < 2*bitsPerLevel {
		t.Fatalf("shift = %d, want >= %d", r.shift, 2*bitsPerLevel)
	}
	if r.root == nil || len(r.root.children) != 2 {
		t.Fatalf("root children = %v, want 2", func() int {
			if r.root == nil {
				return -1
			}
			return len(r.root.children)
		}())
	}

	next, val := r.Pop()
	if val != n-1 {
		t.Fatalf("val = %d, want %d", val, n-1)
	}
	if next.Len() != n-1 {
		t.Fatalf("Len = %d, want %d", next.Len(), n-1)
	}
	if next.shift != bitsPerLevel {
		t.Fatalf("shift after compress Pop = %d, want %d", next.shift, bitsPerLevel)
	}
	if next.root == nil || len(next.root.children) != branching {
		t.Fatalf("after compress: children=%v, want %d", func() int {
			if next.root == nil {
				return -1
			}
			return len(next.root.children)
		}(), branching)
	}
	// Prior deep version intact.
	if r.shift < 2*bitsPerLevel || r.Len() != n {
		t.Fatal("prior version mutated by compress Pop")
	}
	for i := 0; i < next.Len(); i++ {
		if g := next.At(i); g != i {
			t.Fatalf("next At(%d)=%d", i, g)
		}
	}
}

func TestAppendAfterPopPastOverflow(t *testing.T) {
	t.Parallel()

	// Cross the overflow boundary both ways: Pop below it, then Append back
	// past it, and confirm values + shift stay correct.
	const n = 1100
	r := Empty[int]()
	for i := 0; i < n; i++ {
		r = r.Append(i)
	}
	if r.shift < 2*bitsPerLevel {
		t.Fatalf("precondition shift = %d", r.shift)
	}

	// Pop past the overflow edge (1057→1056 compresses) down to a shallow size.
	cur := r
	for cur.Len() > 1000 {
		cur, _ = cur.Pop()
	}
	if cur.shift >= 2*bitsPerLevel {
		t.Fatalf("expected compressed shift after popping below overflow, got %d", cur.shift)
	}
	for i := 0; i < cur.Len(); i++ {
		if g := cur.At(i); g != i {
			t.Fatalf("after Pop At(%d)=%d", i, g)
		}
	}

	// Re-grow through and past the overflow boundary.
	target := 1200
	for i := cur.Len(); i < target; i++ {
		cur = cur.Append(i)
	}
	if cur.Len() != target {
		t.Fatalf("Len = %d, want %d", cur.Len(), target)
	}
	if cur.shift < 2*bitsPerLevel {
		t.Fatalf("expected re-overflow shift>=%d, got %d", 2*bitsPerLevel, cur.shift)
	}
	for i := 0; i < target; i++ {
		if g := cur.At(i); g != i {
			t.Fatalf("after re-Append At(%d)=%d", i, g)
		}
	}
	// Original deep version still independent.
	if r.Len() != n || r.At(n-1) != n-1 {
		t.Fatal("prior deep version mutated")
	}
}

func TestTransientPopDeepExclusiveAndCOW(t *testing.T) {
	t.Parallel()

	// n=1100 is past overflow (shift>=10) with a multi-element tail.
	// n=1089 is the deep pull boundary (tail len==1) used for COW-on-pull.
	const deepN = 1100
	const pullN = 1089

	base := Empty[int]()
	for i := 0; i < deepN; i++ {
		base = base.Append(i)
	}
	if base.shift < 2*bitsPerLevel {
		t.Fatalf("precondition shift = %d", base.shift)
	}

	// Tail-only TransientPop on published deep vector shares the root pointer
	// (no trie edit) but must leave the published base intact.
	ownerTail := ownership.New()
	tailNext, val := base.TransientPop(ownerTail)
	if val != deepN-1 {
		t.Fatalf("tail TransientPop val = %d, want %d", val, deepN-1)
	}
	if base.root != tailNext.root {
		t.Fatal("tail-only TransientPop should keep the published root pointer")
	}
	if base.Len() != deepN || base.At(deepN-1) != deepN-1 {
		t.Fatal("published deep base mutated by tail TransientPop")
	}
	ownerTail.Invalidate()

	// Published deep pull (tail len==1): must COW the root under a new owner.
	pullBase := Empty[int]()
	for i := 0; i < pullN; i++ {
		pullBase = pullBase.Append(i)
	}
	if pullBase.shift < 2*bitsPerLevel {
		t.Fatalf("pullBase shift = %d", pullBase.shift)
	}
	if len(pullBase.tail) != 1 {
		t.Fatalf("pullBase tail len = %d, want 1", len(pullBase.tail))
	}
	ownerPull := ownership.New()
	pulled, val := pullBase.TransientPop(ownerPull)
	if val != pullN-1 {
		t.Fatalf("pull TransientPop val = %d, want %d", val, pullN-1)
	}
	if pullBase.root == pulled.root {
		t.Fatal("new owner must COW published deep root on TransientPop pull")
	}
	if pullBase.Len() != pullN || pullBase.At(pullN-1) != pullN-1 {
		t.Fatal("published pullBase mutated by TransientPop pull")
	}
	if pullBase.shift < 2*bitsPerLevel {
		t.Fatal("published pullBase shift mutated")
	}
	// Further exclusive TransientPop under the same owner reuses owned structure.
	rootOwned := pulled.root
	pulled, val = pulled.TransientPop(ownerPull)
	if val != pullN-2 {
		t.Fatalf("exclusive TransientPop val = %d", val)
	}
	if pulled.root != rootOwned {
		t.Fatal("second TransientPop under same owner should reuse owned root on tail shrink")
	}
	for i := 0; i < pulled.Len(); i++ {
		if g := pulled.At(i); g != i {
			t.Fatalf("after exclusive TransientPop At(%d)=%d", i, g)
		}
	}
	ownerPull.Invalidate()

	// Exclusive build + TransientPop across the overflow boundary under one owner.
	ownerExcl := ownership.New()
	excl := Empty[int]()
	for i := 0; i < deepN; i++ {
		excl = excl.TransientAppend(ownerExcl, i)
	}
	if excl.shift < 2*bitsPerLevel {
		t.Fatalf("exclusive build shift = %d", excl.shift)
	}
	for remain := deepN; remain > 1050; remain-- {
		var got int
		excl, got = excl.TransientPop(ownerExcl)
		if got != remain-1 {
			t.Fatalf("exclusive deep TransientPop val = %d, want %d", got, remain-1)
		}
	}
	if excl.shift >= 2*bitsPerLevel {
		t.Fatalf("expected compressed exclusive shift after popping below overflow, got %d", excl.shift)
	}
	for i := 0; i < excl.Len(); i++ {
		if g := excl.At(i); g != i {
			t.Fatalf("exclusive deep At(%d)=%d", i, g)
		}
	}
	// Published bases still untouched after exclusive path.
	if base.Len() != deepN || base.At(0) != 0 || base.At(deepN-1) != deepN-1 {
		t.Fatal("published base corrupted after exclusive deep TransientPop")
	}
	if pullBase.Len() != pullN || pullBase.At(pullN-1) != pullN-1 {
		t.Fatal("published pullBase corrupted after exclusive deep TransientPop")
	}
	ownerExcl.Invalidate()
}

func TestTransientPopEmptyPanics(t *testing.T) {
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
	owner := ownership.New()
	defer owner.Invalidate()
	_, _ = Empty[int]().TransientPop(owner)
}

func TestTransientPopSoleElement(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	defer owner.Invalidate()
	r := Empty[int]().TransientAppend(owner, 42)
	next, val := r.TransientPop(owner)
	if val != 42 {
		t.Fatalf("val = %d, want 42", val)
	}
	if next.Len() != 0 {
		t.Fatalf("Len = %d, want 0", next.Len())
	}

	// TransientPop on the resulting empty root panics with the same string.
	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("expected panic on empty TransientPop")
		}
		msg, ok := rec.(string)
		if !ok || msg != "vector: index out of range" {
			t.Fatalf("panic = %#v", rec)
		}
	}()
	_, _ = next.TransientPop(owner)
}

func TestPopStructuralSharing(t *testing.T) {
	t.Parallel()

	// n=65 → two trie leaves (0..31, 32..63) + one-element tail.
	// Pop pulls the second leaf into the tail; the first leaf must stay shared.
	r := Empty[int]()
	for i := 0; i < 65; i++ {
		r = r.Append(i)
	}
	old := r
	next, val := r.Pop()
	if val != 64 {
		t.Fatalf("val = %d", val)
	}
	if old.root == nil || next.root == nil {
		t.Fatal("expected non-nil roots before and after Pop at n=65→64")
	}
	if len(old.root.children) < 2 {
		t.Fatalf("old root children = %d, want >=2", len(old.root.children))
	}
	if len(next.root.children) != 1 {
		t.Fatalf("next root children = %d, want 1", len(next.root.children))
	}
	if old.root.children[0] != next.root.children[0] {
		t.Fatal("untouched sibling leaf must be shared by pointer after Pop")
	}
	if old.root == next.root {
		t.Fatal("root must be path-copied when a trie leaf is removed")
	}
	if old.At(0) != 0 || next.At(33) != 33 {
		t.Fatal("values broken after structural-sharing Pop")
	}
}

func TestTransientPopMultiUnderSameOwner(t *testing.T) {
	t.Parallel()

	// Build to n=33 so the first TransientPop pulls a trie leaf (len(tail)==1).
	owner := ownership.New()
	r := Empty[int]()
	for i := 0; i < 33; i++ {
		r = r.TransientAppend(owner, i)
	}
	if r.root == nil || len(r.tail) != 1 {
		t.Fatalf("precondition: root nil=%v tail len=%d", r.root == nil, len(r.tail))
	}
	rootBeforePull := r.root
	var val int
	r, val = r.TransientPop(owner)
	if val != 32 {
		t.Fatalf("first TransientPop val = %d", val)
	}
	// Pull empties the trie (single leaf removed) → root becomes nil.
	if r.root != nil {
		t.Fatal("expected nil root after popping the only trie leaf")
	}
	if r.Len() != 32 {
		t.Fatalf("Len after pull = %d", r.Len())
	}
	// Multi-pop on exclusive tail reuses backing; values stay correct.
	tailAfterPull := r.tail
	r, val = r.TransientPop(owner)
	if val != 31 {
		t.Fatalf("second TransientPop val = %d", val)
	}
	if len(r.tail) > 0 && len(tailAfterPull) > 0 && &r.tail[0] != &tailAfterPull[0] {
		t.Fatal("exclusive tail TransientPop should reuse backing array")
	}
	_ = rootBeforePull
	owner.Invalidate()

	// Published n=65: two trie leaves + one-element tail. TransientPop under a
	// fresh owner must COW the root and share the untouched sibling leaf.
	base := Empty[int]()
	for i := 0; i < 65; i++ {
		base = base.Append(i)
	}
	owner2 := ownership.New()
	v, val := base.TransientPop(owner2)
	if val != 64 {
		t.Fatalf("pull val = %d", val)
	}
	if base.root == v.root {
		t.Fatal("new owner must COW published root on TransientPop pull")
	}
	if base.root == nil || v.root == nil {
		t.Fatal("expected non-nil roots at n=65→64")
	}
	if len(base.root.children) < 2 || len(v.root.children) != 1 {
		t.Fatalf("children old=%d new=%d", len(base.root.children), len(v.root.children))
	}
	if base.root.children[0] != v.root.children[0] {
		t.Fatal("untouched sibling leaf must remain shared under TransientPop")
	}
	// Second pop under same owner is a tail shrink; root pointer reused.
	rootOwned := v.root
	v, val = v.TransientPop(owner2)
	if val != 63 {
		t.Fatalf("tail TransientPop val = %d", val)
	}
	if v.root != rootOwned {
		t.Fatal("second TransientPop under same owner should reuse root")
	}
	// Freeze then a third owner must COW again when editing the spine.
	owner2.Invalidate()
	owner3 := ownership.New()
	// Grow tail back to force a later pull? Simpler: TransientSet into trie.
	v3 := v.TransientSet(owner3, 0, 999)
	if v.root == v3.root {
		t.Fatal("after freeze, new owner must COW root on TransientSet")
	}
	if v.At(0) != 0 {
		t.Fatal("COW edit mutated frozen version")
	}
	if v3.At(0) != 999 {
		t.Fatalf("v3 At(0)=%d", v3.At(0))
	}
	if base.Len() != 65 || base.At(64) != 64 {
		t.Fatal("published base mutated")
	}
	owner3.Invalidate()
}

func TestTransientPopDoesNotMutatePublished(t *testing.T) {
	t.Parallel()

	pub := Empty[int]()
	for i := 0; i < 40; i++ {
		pub = pub.Append(i)
	}
	owner := ownership.New()
	next, val := pub.TransientPop(owner)
	if val != 39 {
		t.Fatalf("val = %d", val)
	}
	if next.Len() != 39 {
		t.Fatalf("next Len = %d", next.Len())
	}
	if pub.Len() != 40 {
		t.Fatalf("published Len = %d", pub.Len())
	}
	for i := 0; i < 40; i++ {
		if g := pub.At(i); g != i {
			t.Fatalf("published At(%d)=%d", i, g)
		}
	}
	// Further exclusive tail pops must still leave pub intact.
	_, _ = next.TransientPop(owner)
	if pub.Len() != 40 || pub.At(39) != 39 {
		t.Fatal("second TransientPop mutated published vector")
	}
	owner.Invalidate()
}

func TestTransientPopTailInPlace(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[int]().TransientAppend(owner, 1, 2, 3, 4)
	tail0 := r.tail
	r, val := r.TransientPop(owner)
	if val != 4 {
		t.Fatalf("val = %d", val)
	}
	if len(r.tail) != 3 {
		t.Fatalf("tail len = %d", len(r.tail))
	}
	if len(tail0) > 0 && len(r.tail) > 0 && &r.tail[0] != &tail0[0] {
		t.Fatal("exclusive TransientPop should reuse tail backing when cap allows")
	}
	if r.At(0) != 1 || r.At(2) != 3 {
		t.Fatal("values after in-place TransientPop")
	}
	owner.Invalidate()
}

func TestPopAfterAppendLoopOrder(t *testing.T) {
	t.Parallel()

	// Exercise several pull boundaries: 33, 64→63 (tail shrink after rebuild),
	// and a deeper size.
	for _, n := range []int{1, 2, 32, 33, 64, 65, 100, 1025, 1057, 1100} {
		r := Empty[int]()
		for i := 0; i < n; i++ {
			r = r.Append(i * 7)
		}
		for remain := n; remain > 0; remain-- {
			if r.Len() != remain {
				t.Fatalf("n=%d remain=%d: Len=%d", n, remain, r.Len())
			}
			var val int
			r, val = r.Pop()
			if val != (remain-1)*7 {
				t.Fatalf("n=%d remain=%d: val=%d", n, remain, val)
			}
		}
	}
}
