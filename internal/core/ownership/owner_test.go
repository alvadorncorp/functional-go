package ownership_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
)

func TestNewReturnsNonNilDistinct(t *testing.T) {
	t.Parallel()

	a := ownership.New()
	b := ownership.New()
	if a == nil || b == nil {
		t.Fatal("New returned nil")
	}
	if a == b {
		t.Fatal("two New() calls returned the same pointer")
	}
}

func TestIsOwnedBySamePointer(t *testing.T) {
	t.Parallel()

	o := ownership.New()
	if !ownership.IsOwnedBy(o, o) {
		t.Fatal("same pointer must be owned")
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
