package workflow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

type inputContextKey struct{}

func TestProvideAppliesOnceAndDefersEffectsUntilRun(t *testing.T) {
	applied := 0
	effects := 0
	subject := workflow.From(func(input int) action.Action[int] {
		applied++
		return action.Do(func(ctx context.Context) (int, error) {
			effects++
			marker, _ := ctx.Value(inputContextKey{}).(int)
			return input + marker, nil
		})
	})

	program := workflow.Provide(subject, 40)
	if applied != 1 || effects != 0 {
		t.Fatalf("after Provide: applied=%d effects=%d, want 1 0", applied, effects)
	}

	//nolint:staticcheck // private typed key verifies exact context propagation.
	ctx := context.WithValue(t.Context(), inputContextKey{}, 2)
	value, err := action.Run(ctx, program)
	if err != nil || value != 42 || effects != 1 {
		t.Fatalf("Run() = (%d, %v), effects=%d; want (42, nil), 1", value, err, effects)
	}
}

func TestProvidePreservesErrorAndRejectsInvalidWorkflows(t *testing.T) {
	wantErr := errors.New("load failed")
	subject := workflow.From(func(int) action.Action[int] { return action.Fail[int](wantErr) })

	_, err := action.Run(t.Context(), workflow.Provide(subject, 1))
	if err != wantErr {
		t.Fatalf("Run() error = %v, want identical %v", err, wantErr)
	}

	assertInputPanic(t, "workflow: nil workflow", func() {
		workflow.Provide(workflow.Workflow[int, int](nil), 1)
	})
	assertInputPanic(t, "workflow: nil action", func() {
		workflow.Provide(workflow.Workflow[int, int](func(int) action.Action[int] { return nil }), 1)
	})
}

func TestContramapMapsAtApplicationAndLeavesActionLazy(t *testing.T) {
	mapped := 0
	effects := 0
	base := workflow.From(func(input int) action.Action[int] {
		return action.Do(func(context.Context) (int, error) {
			effects++
			return input * 2, nil
		})
	})

	subject := workflow.Contramap(base, func(input string) int {
		mapped++
		return len(input)
	})
	if mapped != 0 || effects != 0 {
		t.Fatalf("after construction: mapped=%d effects=%d, want 0 0", mapped, effects)
	}

	program := subject("four")
	if mapped != 1 || effects != 0 {
		t.Fatalf("after application: mapped=%d effects=%d, want 1 0", mapped, effects)
	}

	value, err := action.Run(t.Context(), program)
	if err != nil || value != 8 || effects != 1 {
		t.Fatalf("Run() = (%d, %v), effects=%d; want (8, nil), 1", value, err, effects)
	}
}

func TestContramapRejectsNilArgumentsAndNilResult(t *testing.T) {
	base := workflow.From(action.Pure[int])
	var nilMapper func(string) int

	assertInputPanic(t, "workflow: nil workflow", func() {
		workflow.Contramap(workflow.Workflow[int, int](nil), func(string) int { return 0 })
	})
	assertInputPanic(t, "workflow: nil mapper", func() {
		workflow.Contramap(base, nilMapper)
	})
	assertInputPanic(t, "workflow: nil action", func() {
		bad := workflow.Contramap(
			workflow.Workflow[int, int](func(int) action.Action[int] { return nil }),
			func(input string) int { return len(input) },
		)
		_ = bad("input")
	})
}

func assertInputPanic(t *testing.T, want string, call func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	call()
}
