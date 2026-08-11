package action_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func TestAttemptReifiesSuccessAndFailure(t *testing.T) {
	success, outerErr := action.Run(t.Context(), action.Attempt(action.Pure(42)))
	if outerErr != nil || success.Value != 42 || success.Err != nil {
		t.Fatalf("success = (%+v, %v), want value 42 and nil errors", success, outerErr)
	}

	wantErr := errors.New("failed")
	source := action.Do(func(context.Context) (int, error) { return 7, wantErr })
	failure, outerErr := action.Run(t.Context(), action.Attempt(source))
	if outerErr != nil || failure.Value != 7 || failure.Err != wantErr {
		t.Fatalf("failure = (%+v, %v), want value 7 and identical captured error", failure, outerErr)
	}
}

func TestAttemptIsReExecutable(t *testing.T) {
	runs := 0
	program := action.Attempt(action.Do(func(context.Context) (int, error) {
		runs++
		return runs, nil
	}))
	first, _ := action.Run(t.Context(), program)
	second, _ := action.Run(t.Context(), program)
	if first.Value != 1 || second.Value != 2 || runs != 2 {
		t.Fatalf("values = (%d, %d), runs=%d; want (1, 2), 2", first.Value, second.Value, runs)
	}
}

func TestAttemptDoesNotRecoverPanics(t *testing.T) {
	program := action.Attempt(action.Do(func(context.Context) (int, error) {
		panic("attempt panic")
	}))
	assertPanics(t, "attempt panic", func() {
		_, _ = action.Run(t.Context(), program)
	})
}

func TestAttemptRejectsNilAction(t *testing.T) {
	var nilAction action.Action[int]
	assertPanics(t, "workflow/action: nil action", func() { action.Attempt(nilAction) })
}
