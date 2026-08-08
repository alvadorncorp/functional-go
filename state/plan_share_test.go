package state

import "testing"

// TestApplySharesOlderNode asserts reverse-linked structural sharing:
// branches from the same parent point their older link at that parent's
// head, and each branch allocates a distinct new head. This is the
// structural property that makes Apply O(1) plan growth with prefix sharing.
func TestApplySharesOlderNode(t *testing.T) {
	t.Parallel()

	s1 := Of(0).Apply(func(v int) int { return v + 1 })
	s2 := s1.Apply(func(v int) int { return v * 2 })
	s3 := s1.Apply(func(v int) int { return v + 100 })

	if s1.head == nil {
		t.Fatal("s1.head is nil")
	}
	if s2.head == nil || s3.head == nil {
		t.Fatal("branch heads are nil")
	}
	if s2.head.older != s1.head {
		t.Fatal("s2.head.older must share s1.head (prefix sharing)")
	}
	if s3.head.older != s1.head {
		t.Fatal("s3.head.older must share s1.head (prefix sharing)")
	}
	if s2.head == s3.head {
		t.Fatal("sibling branches must have distinct heads")
	}
	if s2.head == s1.head || s3.head == s1.head {
		t.Fatal("Apply must allocate a new head, not mutate the parent")
	}
}
