package modeling_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

// Local aliases keep compile-time signature checks under the lll budget while
// still pinning the public Decision constructor and method expressions.
type (
	facadeDecision = modeling.Decision[decisionState, decisionEvent]
	facadeDecide   = func(decisionState) facadeDecision
	facadeStateOf  = func(facadeDecision) decisionState
	facadeEmit     = func(facadeDecision, ...decisionEvent) facadeDecision
	facadeCount    = func(facadeDecision) int
	facadeEvents   = func(facadeDecision) iter.Seq[decisionEvent]
)

var (
	_ facadeDecide  = modeling.Decide[decisionEvent, decisionState]
	_ facadeStateOf = modeling.Decision[decisionState, decisionEvent].State
	_ facadeEmit    = modeling.Decision[decisionState, decisionEvent].Emit
	_ facadeCount   = modeling.Decision[decisionState, decisionEvent].EventCount
	_ facadeEvents  = modeling.Decision[decisionState, decisionEvent].Events
)

func TestDecisionFacadePreservesZeroValueAndSnapshots(t *testing.T) {
	var zero modeling.Decision[decisionState, decisionEvent]
	first := zero.Emit(itemAdded{product: "book"})
	second := first.Emit(orderClosed{})

	if zero.State() != (decisionState{}) || zero.EventCount() != 0 {
		t.Fatalf("zero = (%#v, %d), want zero state and no events", zero.State(), zero.EventCount())
	}
	if first.EventCount() != 1 {
		t.Fatalf("first EventCount() = %d, want 1", first.EventCount())
	}
	if got := slices.Collect(second.Events()); len(got) != 2 {
		t.Fatalf("second Events() = %#v, want two ordered events", got)
	}
}
