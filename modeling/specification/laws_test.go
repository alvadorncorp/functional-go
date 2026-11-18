package specification_test

import (
	"testing"

	"go.alvadorn.org/functional/modeling/specification"
)

func TestSpecificationLaws(t *testing.T) {
	even := specification.Specification[int](func(value int) bool { return value%2 == 0 })
	alwaysTrue := specification.Specification[int](func(int) bool { return true })
	alwaysFalse := specification.Specification[int](func(int) bool { return false })

	assertSame := func(name string, left, right specification.Specification[int]) {
		t.Helper()
		for _, value := range []int{-2, -1, 0, 1, 2} {
			if left(value) != right(value) {
				t.Fatalf("%s failed at %d", name, value)
			}
		}
	}

	assertSame("double negation", specification.Not(specification.Not(even)), even)
	assertSame("All unary", specification.All(even), even)
	assertSame("Any unary", specification.Any(even), even)
	assertSame("All true identity", specification.All(even, alwaysTrue), even)
	assertSame("Any false identity", specification.Any(even, alwaysFalse), even)
	if !specification.All[int]()(0) {
		t.Fatal("All() identity failed")
	}
	if specification.Any[int]()(0) {
		t.Fatal("Any() identity failed")
	}
}
