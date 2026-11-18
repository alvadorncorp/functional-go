package modeling_test

import (
	"errors"
	"testing"

	"go.alvadorn.org/functional/modeling"
)

func addTransition(state, input int) int {
	return state + input
}

var errNegativeInput = errors.New("negative input")

func addFallibleTransition(state, input int) (int, error) {
	if input < 0 {
		return state, errNegativeInput
	}
	return state + input, nil
}

func TestTransitionAdaptsOrdinaryFunction(t *testing.T) {
	var transition modeling.Transition[int, int, int] = addTransition
	if got := transition(2, 3); got != 5 {
		t.Fatalf("direct transition = %d, want 5", got)
	}
	if got := transition.Apply(2, 3); got != 5 {
		t.Fatalf("Apply = %d, want 5", got)
	}
}

func TestFallibleTransitionAdaptsOrdinaryFunction(t *testing.T) {
	var transition modeling.FallibleTransition[int, int, int] = addFallibleTransition
	if got, err := transition.Apply(2, 3); err != nil || got != 5 {
		t.Fatalf("Apply = (%d, %v), want (5, nil)", got, err)
	}
	got, err := transition(2, -1)
	if got != 2 || !errors.Is(err, errNegativeInput) {
		t.Fatalf("direct transition = (%d, %v), want (2, errNegativeInput)", got, err)
	}
}
