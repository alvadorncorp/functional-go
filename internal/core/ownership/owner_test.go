package ownership_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

func TestNewReturnsNonNilDistinctActive(t *testing.T) {
	t.Parallel()

	a := ownership.New()
	b := ownership.New()
	if a == nil || b == nil {
		t.Fatal("New returned nil")
	}
	if a == b {
		t.Fatal("two New() calls returned the same pointer")
	}
	if !a.Active() || !b.Active() {
		t.Fatal("New must return an active owner")
	}
}

func TestIsOwnedBySameActivePointer(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	if !ownership.IsOwnedBy(o, o) {
		t.Fatal("same active pointer must be owned")
	}
}

func TestIsOwnedByDifferentPointers(t *testing.T) {
	t.Parallel()

	a, b := ownership.New(), ownership.New()
	if ownership.IsOwnedBy(a, b) {
		t.Fatal("distinct owners must not match")
	}
}

func TestIsOwnedByNilCases(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	cases := []struct {
		name      string
		node, own *ownership.Owner
	}{
		{"both nil", nil, nil},
		{"node nil", nil, o},
		{"owner nil", o, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if ownership.IsOwnedBy(tc.node, tc.own) {
				t.Fatal("nil involvement must not be owned")
			}
		})
	}
}

func TestInvalidateDeactivatesIdempotent(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	if !o.Active() {
		t.Fatal("expected active before Invalidate")
	}
	o.Invalidate()
	if o.Active() {
		t.Fatal("expected inactive after Invalidate")
	}
	// Idempotent.
	o.Invalidate()
	if o.Active() {
		t.Fatal("second Invalidate must leave owner inactive")
	}
}

func TestInvalidateNilReceiverNoOp(t *testing.T) {
	t.Parallel()

	var o *ownership.Owner
	o.Invalidate() // must not panic
	if o.Active() {
		t.Fatal("nil owner is never active")
	}
}

func TestActiveNilFalse(t *testing.T) {
	t.Parallel()

	var o *ownership.Owner
	if o.Active() {
		t.Fatal("nil owner must not be active")
	}
}

func TestIsOwnedByFalseAfterInvalidate(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	// Node still stamped with o after a freeze-like invalidation.
	nodeStamp := o
	o.Invalidate()

	if ownership.IsOwnedBy(nodeStamp, o) {
		t.Fatal("inactive owner must not grant in-place edit even for same pointer")
	}
	if o.Active() {
		t.Fatal("owner must be inactive")
	}
}

func TestIsOwnedByPointerEqualityInsufficientAfterInvalidate(t *testing.T) {
	t.Parallel()

	// Retained pointer from before Freeze: identity alone is not enough.
	o := ownership.New()
	retained := o
	if retained != o {
		t.Fatal("sanity: same pointer")
	}
	o.Invalidate()
	if ownership.IsOwnedBy(retained, retained) {
		t.Fatal("same pointer after Invalidate must not be IsOwnedBy")
	}
}

func TestNewOwnerAfterInvalidateIsIndependent(t *testing.T) {
	t.Parallel()

	old := ownership.New()
	old.Invalidate()
	fresh := ownership.New()
	if ownership.IsOwnedBy(old, fresh) {
		t.Fatal("stale stamp must not match a new active owner")
	}
	if !ownership.IsOwnedBy(fresh, fresh) {
		t.Fatal("fresh owner must own itself")
	}
}
