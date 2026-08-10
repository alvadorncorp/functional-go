//go:build !race

package fold_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/fold"
	"github.com/alvadorncorp/gopackages/functional/values/nonempty"
	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

var (
	foldTestInt      int
	foldTestOptional optional.Optional[int]
	foldTestScan     []int
)

func TestFoldAndReduceDoNotAllocate(t *testing.T) {
	values := []int{1, 2, 3, 4}
	if allocs := testing.AllocsPerRun(1000, func() {
		foldTestInt = fold.FoldLeft(values, 0, add)
	}); allocs != 0 {
		t.Fatalf("FoldLeft allocs = %v", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		foldTestInt = fold.FoldRight(values, 0, func(value, acc int) int { return value + acc })
	}); allocs != 0 {
		t.Fatalf("FoldRight allocs = %v", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		foldTestOptional = fold.Reduce(values, add)
	}); allocs != 0 {
		t.Fatalf("Reduce allocs = %v", allocs)
	}
	nonEmpty := nonempty.Of(1, 2, 3, 4)
	if allocs := testing.AllocsPerRun(1000, func() {
		foldTestInt = fold.ReduceNonEmpty(nonEmpty, add)
	}); allocs != 0 {
		t.Fatalf("ReduceNonEmpty allocs = %v", allocs)
	}
	if allocs := testing.AllocsPerRun(1000, func() {
		foldTestScan = fold.ScanLeft(values, 0, add)
	}); allocs != 1 {
		t.Fatalf("ScanLeft allocs = %v, want 1", allocs)
	}
}
