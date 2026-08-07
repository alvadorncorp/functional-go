//go:build !race

package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

// allocRuns is testing.AllocsPerRun's run count. AllocsPerRun is unreliable
// under -race, hence the !race build tag on this whole file.
const allocRuns = 100

// TestZeroAllocations does not run in parallel with itself or its subtests:
// testing.AllocsPerRun panics when called during a parallel test.
func TestZeroAllocations(t *testing.T) {
	const heldValue = 1

	o := optional.Some(heldValue)
	o2 := optional.Some(heldValue)

	tests := []struct {
		name string
		run  func()
	}{
		{
			name: "Get",
			run: func() {
				_, _ = o.Get()
			},
		},
		{
			name: "OrElse",
			run: func() {
				_ = o.OrElse(0)
			},
		},
		{
			name: "Filter",
			run: func() {
				_ = o.Filter(func(v int) bool { return v > 0 })
			},
		},
		{
			name: "Map",
			run: func() {
				_ = optional.Map(o, func(v int) int { return v + 1 })
			},
		},
		{
			name: "Zip",
			run: func() {
				_ = optional.Zip(o, o2, func(a, b int) int { return a + b })
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if allocs := testing.AllocsPerRun(allocRuns, tc.run); allocs != 0 {
				t.Errorf("%s allocates %v times per run, want 0", tc.name, allocs)
			}
		})
	}
}
