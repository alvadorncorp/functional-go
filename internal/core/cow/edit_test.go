package cow_test

import (
	"testing"

	"go.alvadorn.org/functional/internal/core/cow"
	"go.alvadorn.org/functional/internal/core/ownership"
)

// fakeNode is a test double, not a collection engine.
type fakeNode struct {
	owner *ownership.Owner
	value int
}

func TestEnsureEditableReturnsSameWhenOwned(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	n := &fakeNode{owner: owner, value: 1}
	calls := 0

	got := cow.EnsureEditable(n.owner, owner, n, func() *fakeNode {
		calls++
		return &fakeNode{owner: owner, value: n.value}
	})
	if got != n {
		t.Fatal("expected same node pointer when owned")
	}
	if calls != 0 {
		t.Fatalf("clone called %d times, want 0", calls)
	}
}

func TestEnsureEditableReturnsSameWhenOwnedWithNilClone(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	n := &fakeNode{owner: owner, value: 1}

	got := cow.EnsureEditable[*fakeNode](n.owner, owner, n, nil)
	if got != n {
		t.Fatal("expected same node pointer when owned even if clone is nil")
	}
}

func TestEnsureEditableClonesWhenNotOwned(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	n := &fakeNode{owner: nil, value: 7}
	calls := 0

	got := cow.EnsureEditable(n.owner, owner, n, func() *fakeNode {
		calls++
		return &fakeNode{owner: owner, value: n.value}
	})
	if got == n {
		t.Fatal("expected a new node when not owned")
	}
	if calls != 1 {
		t.Fatalf("clone called %d times, want 1", calls)
	}
	if got.owner != owner || got.value != 7 {
		t.Fatalf("got %+v", got)
	}
}

func TestEnsureEditablePanicsOnNilCloneWhenCopyRequired(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	n := &fakeNode{owner: nil, value: 1}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if s, ok := r.(string); !ok || s != "cow: clone required but nil" {
			t.Fatalf("panic = %v, want cow: clone required but nil", r)
		}
	}()
	_ = cow.EnsureEditable[*fakeNode](n.owner, owner, n, nil)
}

func TestEnsureEditableClonesAfterInvalidate(t *testing.T) {
	t.Parallel()

	owner := ownership.New()
	n := &fakeNode{owner: owner, value: 3}
	owner.Invalidate()
	fresh := ownership.New()
	calls := 0

	got := cow.EnsureEditable(n.owner, fresh, n, func() *fakeNode {
		calls++
		return &fakeNode{owner: fresh, value: n.value}
	})
	if got == n {
		t.Fatal("expected clone when node stamp is an inactive owner")
	}
	if calls != 1 {
		t.Fatalf("clone called %d times, want 1", calls)
	}
	if got.owner != fresh || got.value != 3 {
		t.Fatalf("got %+v", got)
	}
}
