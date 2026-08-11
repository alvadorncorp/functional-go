package specification_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/specification"
)

func TestContramapAdaptsInputOnce(t *testing.T) {
	type order struct{ Total int }
	mapperCalls := 0
	specificationCalls := 0
	highValue := specification.Contramap(
		specification.Specification[int](func(value int) bool {
			specificationCalls++
			return value >= 1_000
		}),
		func(input order) int {
			mapperCalls++
			return input.Total
		},
	)
	if !highValue(order{Total: 1_200}) || mapperCalls != 1 || specificationCalls != 1 {
		t.Fatalf("result/calls = true/%d/%d, want true/1/1", mapperCalls, specificationCalls)
	}
}

func TestContramapComposition(t *testing.T) {
	type order struct{ Total int }
	type request struct{ Order order }
	positive := specification.Specification[int](func(value int) bool { return value > 0 })
	orderTotal := func(input order) int { return input.Total }
	requestOrder := func(input request) order { return input.Order }

	left := specification.Contramap(specification.Contramap(positive, orderTotal), requestOrder)
	right := specification.Contramap(positive, func(input request) int {
		return orderTotal(requestOrder(input))
	})
	for _, input := range []request{{Order: order{Total: -1}}, {Order: order{Total: 1}}} {
		if got, want := left(input), right(input); got != want {
			t.Fatalf("Contramap composition at %+v = %t, want %t", input, got, want)
		}
	}
}

func TestContramapRejectsNilEagerly(t *testing.T) {
	valid := specification.Specification[int](func(value int) bool { return value > 0 })
	var nilSpecification specification.Specification[int]
	var nilMapper func(string) int
	assertSpecificationPanic(t, "modeling/specification: nil specification", func() {
		specification.Contramap(nilSpecification, func(value string) int { return len(value) })
	})
	assertSpecificationPanic(t, "modeling/specification: nil mapper", func() {
		specification.Contramap(valid, nilMapper)
	})
}
