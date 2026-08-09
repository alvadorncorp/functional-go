//go:build !race

package result_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/result"
)

func TestZeroAllocations(t *testing.T) {
	r := result.Ok[int, string](1)
	tests := map[string]func(){
		"Ok":     func() { _ = result.Ok[int, string](1) },
		"Err":    func() { _ = result.Err[int]("bad") },
		"Map":    func() { _ = result.Map(r, func(v int) int { return v + 1 }) },
		"MapErr": func() { _ = result.MapErr(r, func(err string) int { return len(err) }) },
		"FlatMap": func() {
			_ = result.FlatMap(r, func(v int) result.Result[int, string] {
				return result.Ok[int, string](v)
			})
		},
	}
	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			if got := testing.AllocsPerRun(100, run); got != 0 {
				t.Fatalf("allocs/run = %v, want 0", got)
			}
		})
	}
}
