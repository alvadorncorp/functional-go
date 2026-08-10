package decision

import (
	"slices"
	"testing"
)

func TestValueZeroAndNew(t *testing.T) {
	var zero Value[string, int]
	if got := zero.State(); got != "" {
		t.Fatalf("zero State() = %q, want empty", got)
	}
	if got := zero.EventCount(); got != 0 {
		t.Fatalf("zero EventCount() = %d, want 0", got)
	}
	if got := slices.Collect(zero.Events()); len(got) != 0 {
		t.Fatalf("zero Events() = %v, want empty", got)
	}

	value := New[int]("open")
	if got := value.State(); got != "open" {
		t.Fatalf("State() = %q, want open", got)
	}
	if got := value.EventCount(); got != 0 {
		t.Fatalf("EventCount() = %d, want 0", got)
	}
}

func TestValueAppendPreservesReceiverAndOrder(t *testing.T) {
	original := New[int]("open")
	next := original.Append(1, 2, 3)

	if got := original.EventCount(); got != 0 {
		t.Fatalf("original EventCount() = %d, want 0", got)
	}
	if got := next.State(); got != "open" {
		t.Fatalf("next State() = %q, want open", got)
	}
	if got := slices.Collect(next.Events()); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("next Events() = %v, want [1 2 3]", got)
	}
	if got := next.Append().EventCount(); got != 3 {
		t.Fatalf("empty Append EventCount() = %d, want 3", got)
	}
}
