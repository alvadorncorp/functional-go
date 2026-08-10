package fold_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/fold"
)

func FuzzFoldScanConsistency(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{1, 2, 3})
	f.Fuzz(func(t *testing.T, input []byte) {
		values := make([]int, len(input))
		for index, value := range input {
			values[index] = int(value)
		}
		scan := fold.ScanLeft(values, 0, add)
		if len(scan) != len(values)+1 {
			t.Fatalf("scan len = %d", len(scan))
		}
		if got, want := scan[len(scan)-1], fold.FoldLeft(values, 0, add); got != want {
			t.Fatalf("scan/fold = %d/%d", got, want)
		}
	})
}
