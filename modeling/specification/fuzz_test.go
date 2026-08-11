package specification_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/specification"
)

func FuzzSpecificationLaws(f *testing.F) {
	for _, seed := range []int{-6, -1, 0, 1, 6} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value int) {
		even := specification.Specification[int](func(input int) bool { return input%2 == 0 })
		alwaysTrue := specification.Specification[int](func(int) bool { return true })
		alwaysFalse := specification.Specification[int](func(int) bool { return false })

		if specification.Not(specification.Not(even))(value) != even(value) {
			t.Fatal("double negation")
		}
		if specification.All(even, alwaysTrue)(value) != even(value) {
			t.Fatal("All identity")
		}
		if specification.Any(even, alwaysFalse)(value) != even(value) {
			t.Fatal("Any identity")
		}
	})
}
