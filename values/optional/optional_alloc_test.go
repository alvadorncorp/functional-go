//go:build !race

package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

func TestZeroAllocations(t *testing.T) {
	some := optional.Some(1)
	tests := map[string]func(){
		"Some":    func() { _ = optional.Some(1) },
		"Map":     func() { _ = optional.Map(some, func(v int) int { return v + 1 }) },
		"FlatMap": func() { _ = optional.FlatMap(some, optional.Some[int]) },
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if got := testing.AllocsPerRun(100, run); got != 0 {
				t.Fatalf("allocs/run = %v, want 0", got)
			}
		})
	}
}
