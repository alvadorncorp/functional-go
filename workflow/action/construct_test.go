package action_test

import (
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
