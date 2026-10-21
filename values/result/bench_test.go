package result_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/result"
)

func BenchmarkResultOk(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = result.Ok[int, string](1)
	}
}

func BenchmarkResultMap(b *testing.B) {
	b.ReportAllocs()
	r := result.Ok[int, string](1)
	for b.Loop() {
		_ = result.Map(r, func(v int) int { return v + 1 })
	}
}

func BenchmarkResultFlatMap(b *testing.B) {
	b.ReportAllocs()
	r := result.Ok[int, string](1)
	for b.Loop() {
		_ = result.FlatMap(r, func(v int) result.Result[int, string] {
			return result.Ok[int, string](v + 1)
		})
	}
}

func BenchmarkResultFromPair(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = result.FromPair(1, nil)
	}
}

func BenchmarkResultToPair(b *testing.B) {
	b.ReportAllocs()
	r := result.Ok[int, error](1)
	for b.Loop() {
		_, _ = result.ToPair(r)
	}
}

func BenchmarkResultMatch(b *testing.B) {
	b.ReportAllocs()
	r := result.Ok[int, string](1)
	for b.Loop() {
		_ = result.Match(r, func(err string) int { return len(err) }, func(v int) int { return v })
	}
}
