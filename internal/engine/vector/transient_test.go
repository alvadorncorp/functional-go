package vector

import (
	"testing"

	"go.alvadorn.org/functional/internal/core/ownership"
)

func TestTransientMultiAppendThenPublish(t *testing.T) {
	t.Parallel()

	base := Empty[int]().Append(0, 1, 2)
	oldRoot := base.root
	oldTail := base.tail

	owner := ownership.New()
	r := base
	r = r.TransientAppend(owner, 3, 4)
	r = r.TransientSet(owner, 0, 100)
	r = r.TransientDelete(owner, 1)

	owner.Invalidate()

	if r.Len() != 4 {
		t.Fatalf("Len = %d, want 4", r.Len())
	}
	want := []int{100, 2, 3, 4}
	for i, w := range want {
		if g := r.At(i); g != w {
			t.Fatalf("At(%d)=%d want %d", i, g, w)
		}
	}

	// Old immutable root unchanged (shared structure + values).
	if base.root != oldRoot {
		t.Fatal("base root pointer changed")
	}
	if len(base.tail) != len(oldTail) {
		t.Fatal("base tail length changed")
	}
	for i := 0; i < base.Len(); i++ {
		if base.At(i) != i {
			t.Fatalf("base At(%d)=%d", i, base.At(i))
		}
	}

	if r.root != nil {
		assertNotOwnedByInactive(t, r.root, owner)
	}
	if ownership.IsOwnedBy(r.tailOwner, owner) {
		t.Fatal("tail still IsOwnedBy inactive owner")
	}
}

func TestTransientSharingAfterNewOwner(t *testing.T) {
	t.Parallel()

	pub := Empty[int]()
	for i := range 65 {
		pub = pub.Append(i)
	}

	ownerX := ownership.New()
	v2 := pub.TransientSet(ownerX, 0, 1000)
	ownerX.Invalidate()

	ownerY := ownership.New()
	v3 := v2.TransientSet(ownerY, 40, 2000)

	if v2.root == v3.root {
		t.Fatal("root must be path-copied under a new owner")
	}
	if v2.root == nil || v3.root == nil {
		t.Fatal("expected non-nil roots")
	}
	// Untouched sibling under root should remain shared.
	if len(v2.root.children) < 2 || len(v3.root.children) < 2 {
		t.Fatalf("expected >=2 children, got %d and %d",
			len(v2.root.children), len(v3.root.children))
	}
	// Index 0 lives in child 0; index 40 lives in child 1. Editing child 1
	// under ownerY should leave child 0 shared with published v2.
	if v2.root.children[0] != v3.root.children[0] {
		t.Fatal("expected shared untouched child[0] between v2 and v3")
	}
	if v2.root.children[1] == v3.root.children[1] {
		t.Fatal("edited child[1] must differ under new owner")
	}

	if v2.At(0) != 1000 || v3.At(0) != 1000 {
		t.Fatal("shared path value broken")
	}
	if v2.At(40) != 40 {
		t.Fatalf("v2 At(40)=%d", v2.At(40))
	}
	if v3.At(40) != 2000 {
		t.Fatalf("v3 At(40)=%d", v3.At(40))
	}
}

func TestMutationAfterFreezeRequiresFreshOwner(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[int]()
	for i := range 40 {
		r = r.TransientAppend(owner, i)
	}
	if r.root == nil {
		t.Fatal("expected trie root")
	}
	if !ownership.IsOwnedBy(r.root.owner, owner) {
		t.Fatal("active owner must not need copy on root")
	}

	owner.Invalidate()
	if ownership.IsOwnedBy(r.root.owner, owner) {
		t.Fatal("after invalidate, same stamp must need copy")
	}

	fresh := ownership.New()
	cloned := cloneBranch(r.root, fresh)
	if cloned == r.root {
		t.Fatal("expected shallow clone after freeze")
	}
	if cloned.owner != fresh {
		t.Fatal("clone must be stamped with fresh owner")
	}
	// Mutating clone children must not affect published root.
	origChild := r.root.children[0]
	cloned.children[0] = nil
	if r.root.children[0] != origChild {
		t.Fatal("mutating clone must not affect published branch")
	}
}

func TestTransientInPlaceSecondEditSameOwner(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[int]().TransientAppend(owner, 1)
	// Fill past first leaf so a trie root exists, then edit the leaf twice.
	for i := 2; i <= 40; i++ {
		r = r.TransientAppend(owner, i)
	}
	n1 := r.root
	r = r.TransientSet(owner, 0, 99)
	if r.root != n1 {
		t.Fatal("second edit under same active owner should reuse root pointer")
	}
	if r.At(0) != 99 {
		t.Fatalf("At(0)=%d", r.At(0))
	}
	// Leaf under child 0 should also be reused after first ensureEditable.
	leaf1 := r.root.children[0]
	r = r.TransientSet(owner, 1, 98)
	if r.root.children[0] != leaf1 {
		t.Fatal("second leaf edit under same owner should reuse leaf pointer")
	}
	if r.At(1) != 98 {
		t.Fatalf("At(1)=%d", r.At(1))
	}
	owner.Invalidate()
}

func TestTransientTailInPlaceUnderSameOwner(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	r := Empty[int]().TransientAppend(owner, 1, 2, 3)
	tail1 := r.tail
	r = r.TransientAppend(owner, 4)
	// Same backing array when capacity allows (cap was branching).
	if &r.tail[0] != &tail1[0] {
		t.Fatal("exclusive tail append should reuse backing array when cap allows")
	}
	r = r.TransientSet(owner, 1, 20)
	if r.At(1) != 20 || &r.tail[0] != &tail1[0] {
		t.Fatal("exclusive tail Set should mutate in place")
	}
	owner.Invalidate()
}

func TestTransientDoesNotMutatePublishedTail(t *testing.T) {
	t.Parallel()

	pub := Empty[int]().Append(1, 2, 3)
	owner := ownership.New()
	r := pub.TransientSet(owner, 1, 20)
	if pub.At(1) != 2 {
		t.Fatal("TransientSet mutated published tail")
	}
	if r.At(1) != 20 {
		t.Fatalf("transient At(1)=%d", r.At(1))
	}
	// Further append under owner must not touch published tail.
	r = r.TransientAppend(owner, 4)
	if r.Len() != 4 || r.At(3) != 4 {
		t.Fatalf("transient after append: Len=%d At(3)=%v", r.Len(), r.At(3))
	}
	if pub.Len() != 3 || pub.At(2) != 3 {
		t.Fatal("TransientAppend mutated published vector")
	}
	owner.Invalidate()
}

func TestTransientLargeAppend(t *testing.T) {
	t.Parallel()

	const n = 1100
	owner := ownership.New()
	r := Empty[int]()
	for i := range n {
		r = r.TransientAppend(owner, i)
	}
	owner.Invalidate()

	if r.Len() != n {
		t.Fatalf("Len=%d", r.Len())
	}
	for _, i := range []int{0, 31, 32, 33, 1023, 1024, n - 1} {
		if g := r.At(i); g != i {
			t.Fatalf("At(%d)=%d", i, g)
		}
	}
}

func TestTransientBoundsPanics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		fn   func(r Root[int], owner *ownership.Owner)
	}{
		{"TransientSet(-1)", func(r Root[int], owner *ownership.Owner) { _ = r.TransientSet(owner, -1, 0) }},
		{"TransientSet(len)", func(r Root[int], owner *ownership.Owner) { _ = r.TransientSet(owner, 3, 0) }},
		{"TransientDelete(-1)", func(r Root[int], owner *ownership.Owner) { _ = r.TransientDelete(owner, -1) }},
		{"TransientDelete(len)", func(r Root[int], owner *ownership.Owner) { _ = r.TransientDelete(owner, 3) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			owner := ownership.New()
			r := Empty[int]().TransientAppend(owner, 1, 2, 3)
			defer owner.Invalidate()
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
			tc.fn(r, owner)
		})
	}
}

func assertNotOwnedByInactive[T any](t *testing.T, n *node[T], inactive *ownership.Owner) {
	t.Helper()
	if n == nil {
		return
	}
	if ownership.IsOwnedBy(n.owner, inactive) {
		t.Fatal("node still IsOwnedBy inactive owner")
	}
	for _, ch := range n.children {
		assertNotOwnedByInactive(t, ch, inactive)
	}
}
