package either_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/either"
)

func BenchmarkEitherLeft(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = either.Left[string, int]("left")
	}
}

func BenchmarkEitherRight(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = either.Right[string, int](1)
	}
}

func BenchmarkEitherMapRight(b *testing.B) {
	b.ReportAllocs()
	e := either.Right[string, int](1)
	for b.Loop() {
		_ = either.MapRight(e, func(v int) int { return v + 1 })
	}
}

func BenchmarkEitherMatch(b *testing.B) {
	b.ReportAllocs()
	e := either.Right[string, int](1)
	for b.Loop() {
		_ = either.Match(e, func(v string) int { return len(v) }, func(v int) int { return v })
	}
}
