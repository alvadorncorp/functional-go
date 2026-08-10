package lazy_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/lazy"
)

var lazyBenchmarkInt int

func directValue() int { return 42 }

func BenchmarkLazyDirectCall(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		lazyBenchmarkInt = directValue()
	}
}

func BenchmarkLazyColdGet(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		lazyBenchmarkInt = lazy.New(func() int { return 42 }).Get()
	}
}

func BenchmarkLazyWarmGet(b *testing.B) {
	value := lazy.Of(42)
	lazyBenchmarkInt = value.Get()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		lazyBenchmarkInt = value.Get()
	}
}

func BenchmarkLazyMapChain(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		value := lazy.Map(lazy.Map(lazy.Of(1), func(v int) int { return v + 1 }), func(v int) int { return v * 2 })
		lazyBenchmarkInt = value.Get()
	}
}
