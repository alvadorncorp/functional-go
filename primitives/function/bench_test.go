package function_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/function"
)

var benchmarkFunctionResult int

func BenchmarkCompose(b *testing.B) {
	b.ReportAllocs()
	composed := function.Compose(func(v int) int { return v * 2 }, func(v int) int { return v + 1 })
	for b.Loop() {
		benchmarkFunctionResult = composed(41)
	}
}

func BenchmarkPipe(b *testing.B) {
	b.ReportAllocs()
	piped := function.Pipe(func(v int) int { return v + 1 }, func(v int) int { return v * 2 })
	for b.Loop() {
		benchmarkFunctionResult = piped(41)
	}
}

func BenchmarkDirectVsPipe(b *testing.B) {
	inc := func(v int) int { return v + 1 }
	double := func(v int) int { return v * 2 }
	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFunctionResult = double(inc(41))
		}
	})
	b.Run("pipe", func(b *testing.B) {
		b.ReportAllocs()
		piped := function.Pipe(inc, double)
		for b.Loop() {
			benchmarkFunctionResult = piped(41)
		}
	})
}
