package modeling_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

var (
	errRuleA = errors.New("rule A")
	errRuleB = errors.New("rule B")
	errRuleC = errors.New("rule C")
)

func positiveRule(value int) error {
	if value <= 0 {
		return errRuleA
	}
	return nil
}

var (
	_ modeling.Rule[int]                                                 = positiveRule
	_ func(...modeling.Rule[int]) modeling.Rule[int]                     = modeling.AllRules[int]
	_ func(modeling.Rule[int], ...modeling.Rule[int]) modeling.Rule[int] = modeling.AnyRule[int]
)

func TestAllRulesPreservesOrderAndStopsAtFirstFailure(t *testing.T) {
	calls := make([]string, 0, 3)
	combined := modeling.AllRules(
		func(int) error { calls = append(calls, "A"); return nil },
		func(int) error { calls = append(calls, "B"); return errRuleB },
		func(int) error { calls = append(calls, "C"); return errRuleC },
	)

	err := combined(1)
	if err != errRuleB {
		t.Fatalf("error = %v, want identical errRuleB", err)
	}
	if !slices.Equal(calls, []string{"A", "B"}) {
		t.Fatalf("calls = %v, want [A B]", calls)
	}
}

func TestAllRulesWithNoRulesSucceeds(t *testing.T) {
	if err := modeling.AllRules[int]()(1); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
}

func TestAnyRuleStopsAtFirstSuccess(t *testing.T) {
	calls := make([]string, 0, 3)
	combined := modeling.AnyRule(
		func(int) error { calls = append(calls, "A"); return errRuleA },
		func(int) error { calls = append(calls, "B"); return nil },
		func(int) error { calls = append(calls, "C"); return errRuleC },
	)

	if err := combined(1); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if !slices.Equal(calls, []string{"A", "B"}) {
		t.Fatalf("calls = %v, want [A B]", calls)
	}
}

func TestAnyRuleReturnsLastFailure(t *testing.T) {
	combined := modeling.AnyRule(
		func(int) error { return errRuleA },
		func(int) error { return errRuleB },
		func(int) error { return errRuleC },
	)
	if err := combined(1); err != errRuleC {
		t.Fatalf("error = %v, want identical errRuleC", err)
	}
}

func TestRuleCompositionRejectsNil(t *testing.T) {
	tests := map[string]func(){
		"all":       func() { modeling.AllRules[int](nil) },
		"any first": func() { modeling.AnyRule[int](nil) },
		"any rest":  func() { modeling.AnyRule(positiveRule, nil) },
	}
	for name, compose := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "modeling: nil rule" {
					t.Fatalf("panic = %#v, want %q", got, "modeling: nil rule")
				}
			}()
			compose()
		})
	}
}
