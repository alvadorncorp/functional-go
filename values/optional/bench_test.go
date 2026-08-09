package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

func BenchmarkOptionalSome(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = optional.Some(1)
	}
}

func BenchmarkOptionalMap(b *testing.B) {
	b.ReportAllocs()
	o := optional.Some(1)
	for b.Loop() {
		_ = optional.Map(o, func(v int) int { return v + 1 })
	}
}
