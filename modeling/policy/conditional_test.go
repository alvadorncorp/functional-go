package policy_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/primitives/predicate"
	"github.com/alvadorncorp/gopackages/functional/values/optional"
	"github.com/alvadorncorp/gopackages/modeling/policy"
)

type conditionalPolicy[I, O any] = func(I) optional.Optional[O]

func TestWhenIsLazyAndPreservesSomeZero(t *testing.T) {
	calls := 0
	p := policy.Policy[int, int](func(int) int { calls++; return 0 })
	whenFalse := policy.When(predicate.Predicate[int](func(int) bool { return false }), p)
	if got := whenFalse(1); !got.IsNone() || calls != 0 {
		t.Fatalf("false result/calls = %v/%d, want None/0", got, calls)
	}
	whenTrue := policy.When(predicate.Predicate[int](func(int) bool { return true }), p)
	if got := whenTrue(1); !got.IsSome() || got.OrZero() != 0 || calls != 1 {
		t.Fatalf("true result/calls = %v/%d, want Some(0)/1", got, calls)
	}
}

func TestFirstApplicablePreservesOrderAndShortCircuits(t *testing.T) {
	var order []string
	none := func(name string) conditionalPolicy[int, string] {
		return func(int) optional.Optional[string] { order = append(order, name); return optional.None[string]() }
	}
	some := func(name, value string) conditionalPolicy[int, string] {
		return func(int) optional.Optional[string] { order = append(order, name); return optional.Some(value) }
	}
	selected := policy.FirstApplicable(none("first"), some("second", "chosen"), some("third", "late"))
	value, ok := selected(0).Get()
	if !ok || value != "chosen" || len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("result/order = %q/%v/%v, want chosen/true/[first second]", value, ok, order)
	}
}

func TestFirstApplicableReturnsNoneWhenAllAreNone(t *testing.T) {
	none := func(int) optional.Optional[int] { return optional.None[int]() }
	if got := policy.FirstApplicable(none, none)(1); !got.IsNone() {
		t.Fatalf("result = %v, want None", got)
	}
}

func TestFallbackIsLazyAndForwardsInput(t *testing.T) {
	fallbackCalls := 0
	fallback := policy.Policy[int, int](func(input int) int { fallbackCalls++; return input * 2 })
	some := func(int) optional.Optional[int] { return optional.Some(7) }
	if got := policy.Fallback(some, fallback)(21); got != 7 || fallbackCalls != 0 {
		t.Fatalf("Some result/calls = %d/%d, want 7/0", got, fallbackCalls)
	}
	none := func(int) optional.Optional[int] { return optional.None[int]() }
	if got := policy.Fallback(none, fallback)(21); got != 42 || fallbackCalls != 1 {
		t.Fatalf("None result/calls = %d/%d, want 42/1", got, fallbackCalls)
	}
}

func TestConditionalsRejectNil(t *testing.T) {
	truePredicate := predicate.Predicate[int](func(int) bool { return true })
	p := policy.Policy[int, int](twice)
	var nilPredicate predicate.Predicate[int]
	var nilPolicy policy.Policy[int, int]
	var nilConditional conditionalPolicy[int, int]
	assertPanic(t, "modeling/policy: nil predicate", func() { policy.When(nilPredicate, p) })
	assertPanic(t, "modeling/policy: nil policy", func() { policy.When(truePredicate, nilPolicy) })
	assertPanic(t, "modeling/policy: nil conditional", func() { policy.FirstApplicable(nilConditional) })
	assertPanic(t, "modeling/policy: nil conditional", func() {
		policy.FirstApplicable(func(int) optional.Optional[int] { return optional.None[int]() }, nilConditional)
	})
	assertPanic(t, "modeling/policy: nil conditional", func() { policy.Fallback(nilConditional, p) })
	assertPanic(t, "modeling/policy: nil policy", func() {
		policy.Fallback(func(int) optional.Optional[int] { return optional.None[int]() }, nilPolicy)
	})
}
