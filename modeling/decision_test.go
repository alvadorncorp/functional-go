package modeling_test

import (
	"slices"
	"testing"

	"go.alvadorn.org/functional/modeling"
)

type decisionState struct {
	status string
}

type decisionEvent interface {
	decisionEvent()
}

type itemAdded struct {
	product string
}

func (itemAdded) decisionEvent() {}

type orderClosed struct{}

func (orderClosed) decisionEvent() {}

func TestDecisionStartsWithoutEvents(t *testing.T) {
	state := decisionState{status: "open"}
	decision := modeling.Decide[decisionEvent](state)

	if got := decision.State(); got != state {
		t.Fatalf("State() = %#v, want %#v", got, state)
	}
	if got := decision.EventCount(); got != 0 {
		t.Fatalf("EventCount() = %d, want 0", got)
	}
	if got := slices.Collect(decision.Events()); len(got) != 0 {
		t.Fatalf("Events() = %#v, want empty", got)
	}
}

func TestDecisionZeroValueIsValid(t *testing.T) {
	var decision modeling.Decision[decisionState, decisionEvent]

	if got := decision.State(); got != (decisionState{}) {
		t.Fatalf("State() = %#v, want zero state", got)
	}
	if decision.EventCount() != 0 {
		t.Fatalf("EventCount() = %d, want 0", decision.EventCount())
	}
	for event := range decision.Events() {
		t.Fatalf("unexpected event %#v", event)
	}
}

func TestDecisionEmitPreservesReceiverAndEventOrder(t *testing.T) {
	state := decisionState{status: "open"}
	original := modeling.Decide[decisionEvent](state)
	added := itemAdded{product: "book"}
	closed := orderClosed{}

	next := original.Emit(added, closed)

	if original.EventCount() != 0 {
		t.Fatalf("original EventCount() = %d, want 0", original.EventCount())
	}
	if next.State() != state {
		t.Fatalf("next State() = %#v, want %#v", next.State(), state)
	}
	events := slices.Collect(next.Events())
	if len(events) != 2 {
		t.Fatalf("len(Events()) = %d, want 2", len(events))
	}
	if got, ok := events[0].(itemAdded); !ok || got != added {
		t.Fatalf("Events()[0] = %#v, want %#v", events[0], added)
	}
	if _, ok := events[1].(orderClosed); !ok {
		t.Fatalf("Events()[1] = %#v, want orderClosed", events[1])
	}
}
