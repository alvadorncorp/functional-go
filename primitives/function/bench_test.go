package function_test

import (
	"strconv"
	"testing"

	"go.alvadorn.org/functional/primitives/function"
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

func BenchmarkTransformVsDirect(b *testing.B) {
	direct := func(v int) int { return v + 1 }
	named := function.Transform[int](func(v int) int { return v + 1 })
	b.Run("direct", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFunctionResult = direct(41)
		}
	})
	b.Run("transform", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkFunctionResult = named(41)
		}
	})
}

func BenchmarkPipeTransformsChain(b *testing.B) {
	makeChain := func(n int) function.Transform[int] {
		steps := make([]function.Transform[int], n)
		for i := range steps {
			steps[i] = function.Transform[int](func(v int) int { return v + 1 })
		}
		return function.PipeTransforms(steps...)
	}
	for _, n := range []int{1, 2, 4, 8, 16} {
		b.Run("n="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			transform := makeChain(n)
			for b.Loop() {
				benchmarkFunctionResult = transform(0)
			}
		})
	}
}
