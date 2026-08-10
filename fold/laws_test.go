package fold_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/fold"
)

func TestFoldScanLaws(t *testing.T) {
	t.Parallel()
	for _, values := range [][]int{nil, {1}, {1, 2, 3}, {-3, 4, 9}} {
		scan := fold.ScanLeft(values, 10, add)
		if len(scan) != len(values)+1 {
			t.Fatalf("ScanLeft len = %d, want %d", len(scan), len(values)+1)
		}
		if got, want := scan[len(scan)-1], fold.FoldLeft(values, 10, add); got != want {
			t.Fatalf("last ScanLeft = %d, FoldLeft = %d", got, want)
		}
	}
	if got := fold.FoldLeft([]int{5}, 10, add); got != add(10, 5) {
		t.Fatal("singleton FoldLeft law")
	}
}
