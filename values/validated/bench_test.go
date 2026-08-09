package validated_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/validated"
)

func BenchmarkValidatedValid(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = validated.Valid[int, string](1)
	}
}

func BenchmarkValidatedInvalid(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = validated.Invalid[int]("bad")
	}
}

func BenchmarkValidatedMap(b *testing.B) {
	b.ReportAllocs()
	v := validated.Valid[int, string](1)
	for b.Loop() {
		_ = validated.Map(v, func(value int) int { return value + 1 })
	}
}

func BenchmarkValidatedMap2Valid(b *testing.B) {
	b.ReportAllocs()
	a, other := validated.Valid[int, string](1), validated.Valid[int, string](2)
	for b.Loop() {
		_ = validated.Map2(a, other, func(x, y int) int { return x + y })
	}
}

func BenchmarkValidatedMap2Invalid(b *testing.B) {
	b.ReportAllocs()
	a, other := validated.Invalid[int]("a"), validated.Invalid[int]("b")
	for b.Loop() {
		_ = validated.Map2(a, other, func(x, y int) int { return x + y })
	}
}
