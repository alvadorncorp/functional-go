package predicate_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/predicate"
)

func TestBooleanAndQuantifierLaws(t *testing.T) {
	t.Parallel()
	p := predicate.Predicate[int](func(v int) bool { return v%2 == 0 })
	q := predicate.Predicate[int](func(v int) bool { return v > 0 })
	r := predicate.Predicate[int](func(v int) bool { return v%3 == 0 })
	trueP := predicate.Predicate[int](func(int) bool { return true })
	falseP := predicate.Predicate[int](func(int) bool { return false })
	for _, value := range []int{-6, -1, 0, 1, 6} {
		assertSame := func(name string, left, right predicate.Predicate[int]) {
			t.Helper()
			if left(value) != right(value) {
				t.Fatalf("%s failed at %d", name, value)
			}
		}
		assertSame("double negation", predicate.Not(predicate.Not(p)), p)
		assertSame("And identity", predicate.And(p, trueP), p)
		assertSame("Or identity", predicate.Or(p, falseP), p)
		assertSame("And commutativity", predicate.And(p, q), predicate.And(q, p))
		assertSame("Or commutativity", predicate.Or(p, q), predicate.Or(q, p))
		assertSame("And associativity", predicate.And(p, predicate.And(q, r)), predicate.And(predicate.And(p, q), r))
		assertSame("Or associativity", predicate.Or(p, predicate.Or(q, r)), predicate.Or(predicate.Or(p, q), r))
		assertSame("And De Morgan", predicate.Not(predicate.And(p, q)), predicate.Or(predicate.Not(p), predicate.Not(q)))
		assertSame("Or De Morgan", predicate.Not(predicate.Or(p, q)), predicate.And(predicate.Not(p), predicate.Not(q)))
	}
	values := []int{-2, 0, 2}
	if predicate.None(values, p) != !predicate.Any(values, p) {
		t.Fatal("None(xs,p) != !Any(xs,p)")
	}
	if predicate.All(values, p) != !predicate.Any(values, predicate.Not(p)) {
		t.Fatal("All(xs,p) != !Any(xs,Not(p))")
	}
}
