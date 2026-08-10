package fold_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/primitives/fold"
)

func ExampleScanLeft() {
	states := fold.ScanLeft([]int{1, 2, 3}, 0, func(total, value int) int {
		return total + value
	})
	fmt.Println(states)
	// Output: [0 1 3 6]
}
