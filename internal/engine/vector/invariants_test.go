package vector

import "testing"

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

func TestAllMatchesAtAcrossTrieAndTail(t *testing.T) {
	t.Parallel()

	// Sizes cover empty, tail-only, first trie push (32), multi-leaf, and a
	// deeper root (past 1024) so All's sequential walk is checked against At.
	for _, n := range []int{0, 1, 31, 32, 33, 64, 100, 1100} {
		r := Empty[int]()
		for i := 0; i < n; i++ {
			r = r.Append(i * 3)
		}
		seen := 0
		for i, v := range r.All() {
			if i != seen {
				t.Fatalf("n=%d: All index = %d, want sequential %d", n, i, seen)
			}
			if want := r.At(i); v != want {
				t.Fatalf("n=%d: All(%d)=%d, At(%d)=%d", n, i, v, i, want)
			}
			if want := i * 3; v != want {
				t.Fatalf("n=%d: All(%d)=%d, want %d", n, i, v, want)
			}
			seen++
		}
		if seen != n {
			t.Fatalf("n=%d: All yielded %d pairs, want %d", n, seen, n)
		}
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
