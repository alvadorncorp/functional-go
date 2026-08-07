package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

func BenchmarkSomeOrElse(b *testing.B) {
	b.ReportAllocs()

	const heldValue = 1
	const fallbackValue = 0

	o := optional.Some(heldValue)
	for b.Loop() {
		_ = o.OrElse(fallbackValue)
	}
}

func BenchmarkMap(b *testing.B) {
	b.ReportAllocs()

	const heldValue = 1

	o := optional.Some(heldValue)
	for b.Loop() {
		_ = optional.Map(o, func(v int) int { return v + 1 })
	}
}

func BenchmarkZip(b *testing.B) {
	b.ReportAllocs()

	const heldValueA = 1
	const heldValueB = 2

	a := optional.Some(heldValueA)
	other := optional.Some(heldValueB)
	for b.Loop() {
		_ = optional.Zip(a, other, func(x, y int) int { return x + y })
	}
}
