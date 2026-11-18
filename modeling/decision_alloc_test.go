//go:build !race

package modeling_test

import (
	"testing"

	"go.alvadorn.org/functional/modeling"
)

const decisionAllocRuns = 100

var (
	decisionStateSink decisionState
	decisionCountSink int
)

func TestDecisionTrivialPathsDoNotAllocate(t *testing.T) {
	allocs := testing.AllocsPerRun(decisionAllocRuns, func() {
		decision := modeling.Decide[decisionEvent](decisionState{status: "open"})
		decisionStateSink = decision.State()
		decisionCountSink = decision.EventCount()
	})
	if allocs != 0 {
		t.Fatalf("trivial Decision paths allocate %v times per run, want 0", allocs)
	}
}
