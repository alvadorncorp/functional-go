package vector

import "testing"

func TestStructuralSharingOnSet(t *testing.T) {
	t.Parallel()

	// 65 elements → trie holds two full leaves (0..31 and 32..63) and the
	// tail holds index 64. That gives the root two sibling children so Set
	// on the first leaf can prove the second child pointer is shared.
	r := Empty[int]()
	for i := range 65 {
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
	for i := range n {
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
