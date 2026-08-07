package transient_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/internal/core/ownership"
	"github.com/alvadorncorp/gopackages/functional/internal/core/transient"
)

func TestOpenProvidesOwner(t *testing.T) {
	t.Parallel()

	w := transient.Open()
	o := w.Owner()
	if o == nil {
		t.Fatal("Owner is nil")
	}
	if !o.Active() {
		t.Fatal("open window owner must be active")
	}
	if !ownership.IsOwnedBy(o, o) {
		t.Fatal("owner must be self-owned")
	}
	// Stable across calls before Freeze.
	if w.Owner() != o {
		t.Fatal("Owner must be stable until Freeze")
	}
}

func TestOpenOwnersAreDistinct(t *testing.T) {
	t.Parallel()

	a, b := transient.Open(), transient.Open()
	if a.Owner() == b.Owner() {
		t.Fatal("separate windows must have distinct owners")
	}
}

func TestOwnerPanicsAfterFreeze(t *testing.T) {
	t.Parallel()

	w := transient.Open()
	w.Freeze()
	assertPanic(t, "transient: window used after Freeze", func() {
		_ = w.Owner()
	})
}

func TestSecondFreezePanics(t *testing.T) {
	t.Parallel()

	w := transient.Open()
	w.Freeze()
	assertPanic(t, "transient: window used after Freeze", func() {
		w.Freeze()
	})
}

func TestFreezeInvalidatesRetainedOwner(t *testing.T) {
	t.Parallel()

	w := transient.Open()
	retained := w.Owner()
	// Simulate a node still stamped with this owner after Freeze.
	nodeStamp := retained

	w.Freeze()

	if retained.Active() {
		t.Fatal("Freeze must Invalidate the owner token")
	}
	if ownership.IsOwnedBy(nodeStamp, retained) {
		t.Fatal("after Freeze, IsOwnedBy must be false for the retained stamp")
	}
}

func TestNilWindowPanics(t *testing.T) {
	t.Parallel()

	var w *transient.Window
	assertPanic(t, "transient: window used after Freeze", func() {
		_ = w.Owner()
	})
	assertPanic(t, "transient: window used after Freeze", func() {
		w.Freeze()
	})
}

func TestZeroValueWindowPanics(t *testing.T) {
	t.Parallel()

	assertPanic(t, "transient: window used after Freeze", func() {
		_ = (&transient.Window{}).Owner()
	})
	assertPanic(t, "transient: window used after Freeze", func() {
		(&transient.Window{}).Freeze()
	})
}

func TestZeroValueWindowVarPanics(t *testing.T) {
	t.Parallel()

	assertPanic(t, "transient: window used after Freeze", func() {
		var w transient.Window
		_ = w.Owner()
	})
	assertPanic(t, "transient: window used after Freeze", func() {
		var w transient.Window
		w.Freeze()
	})
}

func assertPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if s, ok := r.(string); !ok || s != want {
			t.Fatalf("panic = %v, want %q", r, want)
		}
	}()
	fn()
}
