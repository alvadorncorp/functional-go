package predicate_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/predicate"
)

func FuzzPredicateLaws(f *testing.F) {
	for _, seed := range []int{-6, -1, 0, 1, 6} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value int) {
		p := predicate.Predicate[int](func(v int) bool { return v%2 == 0 })
		q := predicate.Predicate[int](func(v int) bool { return v > 0 })
		if predicate.Not(predicate.Not(p))(value) != p(value) {
			t.Fatal("double negation")
		}
		if predicate.And(p, q)(value) != predicate.And(q, p)(value) {
			t.Fatal("And commutativity")
		}
		if predicate.Or(p, q)(value) != predicate.Or(q, p)(value) {
			t.Fatal("Or commutativity")
		}
		if predicate.Not(predicate.And(p, q))(value) != predicate.Or(predicate.Not(p), predicate.Not(q))(value) {
			t.Fatal("And De Morgan")
		}
		if predicate.Not(predicate.Or(p, q))(value) != predicate.And(predicate.Not(p), predicate.Not(q))(value) {
			t.Fatal("Or De Morgan")
		}
	})
}
