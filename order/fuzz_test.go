package order_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/order"
)

func FuzzNaturalOrderAntisymmetry(f *testing.F) {
	f.Add(0, 0, 0)
	f.Add(-1, 0, 1)
	f.Add(3, 2, 1)
	f.Fuzz(func(t *testing.T, a, b, c int) {
		o := order.Natural[int]()
		if o(a, a) != 0 {
			t.Fatal("reflexivity")
		}
		if sign(o(a, b)) != -sign(o(b, a)) {
			t.Fatal("sign antisymmetry")
		}
		if o(a, b) <= 0 && o(b, c) <= 0 && o(a, c) > 0 {
			t.Fatal("transitivity")
		}
		if order.Reverse(order.Reverse(o))(a, b) != o(a, b) {
			t.Fatal("reverse involution")
		}
	})
}
