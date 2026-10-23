package action_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

func TestPureYieldsItsValueWithoutEffects(t *testing.T) {
	t.Parallel()

	value, err := action.Run(t.Context(), action.Pure(answer))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != answer {
		t.Fatalf("Run() value = %d, want %d", value, answer)
	}
}

func TestFailPreservesErrorIdentity(t *testing.T) {
	t.Parallel()

	wrapped := errors.Join(errBoom, errors.New("context"))

	_, err := action.Run(t.Context(), action.Fail[int](wrapped))
	if !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want errors.Is(err, errBoom)", err)
	}
}

func TestFailYieldsTheZeroValue(t *testing.T) {
	t.Parallel()

	value, err := action.Run(t.Context(), action.Fail[string](errBoom))
	if err == nil {
		t.Fatal("Run() error = nil, want errBoom")
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
}

func TestDoPanicsOnNilFunction(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil function", func() {
		_ = action.Do[int](nil)
	})
}

func TestFailPanicsOnNilError(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil error", func() {
		_ = action.Fail[int](nil)
	})
}

func TestDoDoesNotCallItsFunction(t *testing.T) {
	t.Parallel()

	called := false
	_ = action.Do(func(context.Context) (int, error) {
		called = true
		return answer, nil
	})

	if called {
		t.Fatal("Do called its function during construction")
	}
}

func TestExecPanicsOnNilFunctionDuringConstruction(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil function", func() {
		_ = action.Exec(nil)
	})
}

func TestExecDefersEffectsUntilEachRun(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	program := action.Exec(func(context.Context) error {
		_, err := output.WriteString("saved\n")
		return err
	})
	if output.Len() != 0 {
		t.Fatal("Exec wrote output during construction")
	}

	for range 2 {
		if _, err := action.Run(t.Context(), program); err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	}
	if got := output.String(); got != "saved\nsaved\n" {
		t.Fatalf("output = %q, want one write per Run", got)
	}
}

func TestExecPreservesErrorAndStopsDependentAction(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	program := action.FlatMap(action.Exec(func(context.Context) error {
		output.WriteString("begin\n")
		return errBoom
	}), func(struct{}) action.Action[struct{}] {
		output.WriteString("commit\n")
		return action.Pure(struct{}{})
	})

	_, err := action.Run(t.Context(), program)
	if err != errBoom {
		t.Fatalf("Run() error = %v, want identical errBoom", err)
	}
	if got := output.String(); got != "begin\n" {
		t.Fatalf("output = %q, want no commit after failure", got)
	}
}

func TestExecUsesScopedResourceDespitePreCancellation(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	errAborted := errors.New("save aborted")
	ctx, cancel := context.WithCancel(context.WithValue(t.Context(), contextKey{}, &output))
	cancel()

	program := action.Exec(func(ctx context.Context) error {
		writer := ctx.Value(contextKey{}).(*bytes.Buffer)
		writer.WriteString("begin\n")
		if ctx.Err() == context.Canceled {
			writer.WriteString("abort\n")
			return errAborted
		}
		writer.WriteString("commit\n")
		return nil
	})

	_, err := action.Run(ctx, program)
	if err != errAborted {
		t.Fatalf("Run() error = %v, want identical save aborted error", err)
	}
	if got := output.String(); got != "begin\nabort\n" {
		t.Fatalf("output = %q, want scoped abort write despite pre-cancellation", got)
	}
}

func TestExecPanicPreservesPartialEffectAndStopsCommit(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	program := action.FlatMap(action.Exec(func(context.Context) error {
		output.WriteString("begin\n")
		panic("exec panic")
	}), func(struct{}) action.Action[struct{}] {
		output.WriteString("commit\n")
		return action.Pure(struct{}{})
	})

	assertPanics(t, "exec panic", func() {
		_, _ = action.Run(t.Context(), program)
	})
	if got := output.String(); got != "begin\n" {
		t.Fatalf("output = %q, want partial write and no commit after panic", got)
	}
}
