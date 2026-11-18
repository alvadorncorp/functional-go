package policy_test

import (
	"strconv"
	"testing"

	"go.alvadorn.org/functional/modeling/policy"
)

func identityInt(value int) int { return value }
func twice(value int) int       { return value * 2 }
func asText(value int) string   { return strconv.Itoa(value) }

func TestMapTransformsOutputOnce(t *testing.T) {
	policyCalls := 0
	mapperCalls := 0
	p := policy.Map(
		policy.Policy[int, int](func(input int) int { policyCalls++; return input + 1 }),
		func(output int) string { mapperCalls++; return strconv.Itoa(output) },
	)
	if got := p(41); got != "42" || policyCalls != 1 || mapperCalls != 1 {
		t.Fatalf("result/calls = %q/%d/%d, want 42/1/1", got, policyCalls, mapperCalls)
	}
}

func TestMapIdentityLaw(t *testing.T) {
	p := policy.Policy[int, int](twice)
	if got, want := policy.Map(p, identityInt)(7), p(7); got != want {
		t.Fatalf("Map identity = %d, want %d", got, want)
	}
}

func TestMapCompositionLaw(t *testing.T) {
	p := policy.Policy[int, int](twice)
	left := policy.Map(policy.Map(p, twice), asText)
	right := policy.Map(p, func(value int) string { return asText(twice(value)) })
	if got, want := left(3), right(3); got != want {
		t.Fatalf("Map composition = %q, want %q", got, want)
	}
}

func TestContramapAdaptsInputOnce(t *testing.T) {
	type context struct{ Value int }
	mapperCalls := 0
	policyCalls := 0
	p := policy.Contramap(
		policy.Policy[int, int](func(input int) int { policyCalls++; return input * 2 }),
		func(input context) int { mapperCalls++; return input.Value },
	)
	if got := p(context{Value: 21}); got != 42 || mapperCalls != 1 || policyCalls != 1 {
		t.Fatalf("result/calls = %d/%d/%d, want 42/1/1", got, mapperCalls, policyCalls)
	}
}

func TestContramapCompositionLaw(t *testing.T) {
	type outer struct{ Raw string }
	parse := func(input string) int { value, _ := strconv.Atoi(input); return value }
	extract := func(input outer) string { return input.Raw }
	p := policy.Policy[int, int](twice)
	left := policy.Contramap(policy.Contramap(p, parse), extract)
	right := policy.Contramap(p, func(input outer) int { return parse(extract(input)) })
	if got, want := left(outer{Raw: "21"}), right(outer{Raw: "21"}); got != want {
		t.Fatalf("Contramap composition = %d, want %d", got, want)
	}
}

func TestTransformsRejectNil(t *testing.T) {
	var nilPolicy policy.Policy[int, int]
	var nilMapper func(int) int
	assertPanic(t, "modeling/policy: nil policy", func() { policy.Map(nilPolicy, twice) })
	assertPanic(t, "modeling/policy: nil mapper", func() {
		policy.Map(policy.Policy[int, int](twice), nilMapper)
	})
	assertPanic(t, "modeling/policy: nil policy", func() { policy.Contramap(nilPolicy, twice) })
	assertPanic(t, "modeling/policy: nil mapper", func() {
		policy.Contramap(policy.Policy[int, int](twice), nilMapper)
	})
}
