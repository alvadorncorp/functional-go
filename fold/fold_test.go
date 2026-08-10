package fold_test

import (
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/fold"
	"github.com/alvadorncorp/gopackages/functional/values/nonempty"
)

func add(total, value int) int { return total + value }

func TestFoldLeftAndRight(t *testing.T) {
	t.Parallel()
	values := []int{1, 2, 3}
	left := fold.FoldLeft(values, "z", func(acc string, value int) string {
		return "(" + acc + "+" + strconv.Itoa(value) + ")"
	})
	if left != "(((z+1)+2)+3)" {
		t.Fatalf("FoldLeft = %q", left)
	}
	right := fold.FoldRight(values, "z", func(value int, acc string) string {
		return "(" + strconv.Itoa(value) + "+" + acc + ")"
	})
	if right != "(1+(2+(3+z)))" {
		t.Fatalf("FoldRight = %q", right)
	}
	if got := fold.FoldLeft([]int(nil), 7, add); got != 7 {
		t.Fatalf("empty FoldLeft = %d", got)
	}
	if got := fold.FoldRight([]int(nil), 7, func(value, acc int) int { return value + acc }); got != 7 {
		t.Fatalf("empty FoldRight = %d", got)
	}
	got := fold.FoldLeft([]string{"a", "bb"}, 0, func(total int, value string) int {
		return total + len(value)
	})
	if got != 3 {
		t.Fatalf("different accumulator type = %d", got)
	}
}

func TestFoldRightTraversesFromRight(t *testing.T) {
	t.Parallel()

	got := fold.FoldRight([]int{1, 2, 3}, 0, func(value, acc int) int {
		return value - acc
	})
	if got != 2 { // 1 - (2 - (3 - 0))
		t.Fatalf("FoldRight = %d, want 2", got)
	}
}

func TestReduceAndReduceNonEmpty(t *testing.T) {
	t.Parallel()
	if got := fold.Reduce([]int(nil), add); !got.IsNone() {
		t.Fatal("Reduce(empty) must be None")
	}
	if got := fold.Reduce([]int{4}, add); got.MustGet() != 4 {
		t.Fatalf("Reduce singleton = %d", got.MustGet())
	}
	if got := fold.Reduce([]int{10, 3, 2}, func(acc, value int) int {
		return acc - value
	}); got.MustGet() != 5 { // (10 - 3) - 2
		t.Fatalf("Reduce = %d, want 5", got.MustGet())
	}
	if got := fold.ReduceNonEmpty(nonempty.Of(10, 3, 2), func(acc, value int) int {
		return acc - value
	}); got != 5 { // (10 - 3) - 2
		t.Fatalf("ReduceNonEmpty = %d, want 5", got)
	}
}

func TestScanLeftIncludesInitial(t *testing.T) {
	t.Parallel()
	got := fold.ScanLeft([]int{1, 2, 3}, 0, add)
	want := []int{0, 1, 3, 6}
	if len(got) != len(want) {
		t.Fatalf("ScanLeft len = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("ScanLeft[%d] = %d, want %d", index, got[index], want[index])
		}
	}
	empty := fold.ScanLeft([]int(nil), 9, add)
	if len(empty) != 1 || empty[0] != 9 {
		t.Fatalf("empty ScanLeft = %v", empty)
	}
}

func TestFoldRightIsIterativeAndStackSafe(t *testing.T) {
	t.Parallel()
	values := make([]int, 100_000)
	for index := range values {
		values[index] = 1
	}
	if got := fold.FoldRight(values, 0, func(value, acc int) int { return value + acc }); got != len(values) {
		t.Fatalf("large FoldRight = %d, want %d", got, len(values))
	}
}
