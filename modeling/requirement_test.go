package modeling_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

type requirementInput struct{ amount int }

var (
	errRequirementA = errors.New("requirement A")
	errRequirementB = errors.New("requirement B")
	errRequirementC = errors.New("requirement C")
)

func enoughBalance(balance int, input requirementInput) error {
	if balance < input.amount {
		return errRequirementA
	}
	return nil
}

type (
	intRequirement      = modeling.Requirement[int, requirementInput]
	allRequirementsFunc = func(...intRequirement) intRequirement
	anyRequirementFunc  = func(intRequirement, ...intRequirement) intRequirement
)

var (
	_ intRequirement      = enoughBalance
	_ allRequirementsFunc = modeling.AllRequirements[int, requirementInput]
	_ anyRequirementFunc  = modeling.AnyRequirement[int, requirementInput]
)

func TestAllRequirementsForwardsInputsAndStopsAtFirstFailure(t *testing.T) {
	calls := make([]string, 0, 3)
	combined := modeling.AllRequirements(
		func(state int, input requirementInput) error {
			calls = append(calls, "A")
			if state != 10 || input.amount != 7 {
				t.Fatal("state/input changed")
			}
			return nil
		},
		func(int, requirementInput) error { calls = append(calls, "B"); return errRequirementB },
		func(int, requirementInput) error { calls = append(calls, "C"); return errRequirementC },
	)

	err := combined(10, requirementInput{amount: 7})
	if err != errRequirementB {
		t.Fatalf("error = %v, want identical errRequirementB", err)
	}
	if !slices.Equal(calls, []string{"A", "B"}) {
		t.Fatalf("calls = %v, want [A B]", calls)
	}
}

func TestAllRequirementsWithNoRequirementsSucceeds(t *testing.T) {
	if err := modeling.AllRequirements[int, requirementInput]()(10, requirementInput{amount: 7}); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
}

func TestAnyRequirementStopsAtFirstSuccess(t *testing.T) {
	calls := make([]string, 0, 3)
	combined := modeling.AnyRequirement(
		func(int, requirementInput) error { calls = append(calls, "A"); return errRequirementA },
		func(int, requirementInput) error { calls = append(calls, "B"); return nil },
		func(int, requirementInput) error { calls = append(calls, "C"); return errRequirementC },
	)

	if err := combined(10, requirementInput{amount: 7}); err != nil {
		t.Fatalf("error = %v, want nil", err)
	}
	if !slices.Equal(calls, []string{"A", "B"}) {
		t.Fatalf("calls = %v, want [A B]", calls)
	}
}

func TestAnyRequirementReturnsLastFailure(t *testing.T) {
	combined := modeling.AnyRequirement(
		func(int, requirementInput) error { return errRequirementA },
		func(int, requirementInput) error { return errRequirementB },
		func(int, requirementInput) error { return errRequirementC },
	)
	if err := combined(10, requirementInput{amount: 7}); err != errRequirementC {
		t.Fatalf("error = %v, want identical errRequirementC", err)
	}
}

func TestRequirementCompositionRejectsNil(t *testing.T) {
	tests := map[string]func(){
		"all":       func() { modeling.AllRequirements[int, requirementInput](nil) },
		"any first": func() { modeling.AnyRequirement[int, requirementInput](nil) },
		"any rest":  func() { modeling.AnyRequirement(enoughBalance, nil) },
	}
	for name, compose := range tests {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != "modeling: nil requirement" {
					t.Fatalf("panic = %#v, want %q", got, "modeling: nil requirement")
				}
			}()
			compose()
		})
	}
}
