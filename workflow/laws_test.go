package workflow_test

import (
	"context"
	"sync"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

const (
	deepChainDepth = 10000

	markerA userID = 10
	markerB userID = 20

	wantA = firstID + markerA
	wantB = secondID + markerB
)

func TestWorkflowExecutionIsolation(t *testing.T) {
	t.Parallel()

	subject := workflow.From(func(id userID) action.Action[userID] {
		return action.Do(func(ctx context.Context) (userID, error) {
			marker, _ := ctx.Value(contextKey{}).(userID)
			return id + marker, nil
		})
	})

	programA := subject(firstID)
	programB := subject(secondID)

	//nolint:staticcheck // the key is a private struct type, not a bare string.
	ctxA := context.WithValue(t.Context(), contextKey{}, markerA)
	//nolint:staticcheck // the key is a private struct type, not a bare string.
	ctxB := context.WithValue(t.Context(), contextKey{}, markerB)

	var wait sync.WaitGroup
	results := make([]userID, 2)
	errs := make([]error, 2)

	wait.Add(2)
	go func() {
		defer wait.Done()
		results[0], errs[0] = action.Run(ctxA, programA)
	}()
	go func() {
		defer wait.Done()
		results[1], errs[1] = action.Run(ctxB, programB)
	}()
	wait.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("Run(%d) error = %v, want nil", i, err)
		}
	}
	if results[0] != wantA {
		t.Fatalf("first result = %d, want %d (input %d + context %d)", results[0], wantA, firstID, markerA)
	}
	if results[1] != wantB {
		t.Fatalf("second result = %d, want %d (input %d + context %d)", results[1], wantB, secondID, markerB)
	}
}

func TestWorkflowContextIndependence(t *testing.T) {
	t.Parallel()

	subject := workflow.From(func(userID) action.Action[string] {
		return action.Do(func(ctx context.Context) (string, error) {
			marker, _ := ctx.Value(contextKey{}).(string)
			return marker, nil
		})
	})

	// One program value, two sequential runs: an Action is a computation, not a
	// memoized future, so each Run observes only the context it was given.
	program := subject(firstID)

	//nolint:staticcheck // the key is a private struct type, not a bare string.
	ctxA := context.WithValue(t.Context(), contextKey{}, "A")
	//nolint:staticcheck // the key is a private struct type, not a bare string.
	ctxB := context.WithValue(t.Context(), contextKey{}, "B")

	valueA, errA := action.Run(ctxA, program)
	valueB, errB := action.Run(ctxB, program)

	if errA != nil || errB != nil {
		t.Fatalf("errors = (%v, %v), want (nil, nil)", errA, errB)
	}
	if valueA != "A" || valueB != "B" {
		t.Fatalf("values = (%q, %q), want (\"A\", \"B\")", valueA, valueB)
	}
}

func TestDeepCompositionIsStable(t *testing.T) {
	t.Parallel()

	step := workflow.From(func(id userID) action.Action[userID] {
		return action.Pure(id + 1)
	})

	chain := step
	for range deepChainDepth - 1 {
		chain = workflow.Then(chain, step)
	}

	value, err := action.Run(t.Context(), chain(0))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != deepChainDepth {
		t.Fatalf("Run() value = %d, want %d", value, deepChainDepth)
	}
}
