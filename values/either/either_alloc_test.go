//go:build !race

package either_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/either"
)

func TestZeroAllocations(t *testing.T) {
	e := either.Right[string, int](1)
	left := either.Left[int, string](1)
	tests := map[string]func(){
		"Left":     func() { _ = either.Left[string, int]("left") },
		"Right":    func() { _ = either.Right[string, int](1) },
		"MapRight": func() { _ = either.MapRight(e, func(v int) int { return v + 1 }) },
		"MapLeft":  func() { _ = either.MapLeft(left, func(int) string { return "mapped" }) },
		"Bimap":    func() { _ = either.Bimap(e, func(v string) int { return len(v) }, func(v int) int { return v + 1 }) },
		"FlatMapRight": func() {
			_ = either.FlatMapRight(e, func(v int) either.Either[string, int] { return either.Right[string](v + 1) })
		},
		"Match": func() {
			_ = either.Match(e, func(v string) int { return len(v) }, func(v int) int { return v })
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
