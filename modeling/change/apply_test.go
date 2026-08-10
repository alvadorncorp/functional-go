package change_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/modeling/change"
)

type (
	applyFunction[S any]    func(S, change.Change[S]) S
	applyAllFunction[S any] func(S, ...change.Change[S]) S
)

var (
	_ applyFunction[int]    = change.Apply[int]
	_ applyAllFunction[int] = change.ApplyAll[int]
)

func appendStep(step string) change.Change[string] {
	return func(state string) string { return state + step }
}

func TestApplyUsesValueFirstOrder(t *testing.T) {
	transformation := appendStep("A")
	if got, want := change.Apply("", transformation), transformation(""); got != want {
		t.Fatalf("Apply result = %q, want %q", got, want)
	}
}

func TestApplyAllRunsLeftToRight(t *testing.T) {
	got := change.ApplyAll("", appendStep("A"), appendStep("B"), appendStep("C"))
	if want := "ABC"; got != want {
		t.Fatalf("ApplyAll result = %q, want %q", got, want)
	}
}

func TestApplyAllWithNoChangesReturnsState(t *testing.T) {
	if got, want := change.ApplyAll("unchanged"), "unchanged"; got != want {
		t.Fatalf("ApplyAll(state) = %q, want %q", got, want)
	}
}

func TestApplyAllEqualsPipeApplication(t *testing.T) {
	changes := []change.Change[string]{appendStep("A"), appendStep("B")}
	if got, want := change.ApplyAll("", changes...), change.Pipe(changes...)(""); got != want {
		t.Fatalf("ApplyAll result = %q, Pipe result = %q", got, want)
	}
}
