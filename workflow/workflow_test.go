package workflow_test

import (
	"context"
	"errors"
	"testing"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
)

type userID int

type user struct {
	ID     userID
	Name   string
	Active bool
}

const (
	firstID  userID = 1
	secondID userID = 2
)

var errNotFound = errors.New("user not found")

// users is the capability shape the module recommends: a small struct of
// functions, injected by the composition root, never a container lookup.
type users struct {
	Find func(userID) action.Action[user]
}

func fixedUsers(name string, loads *int) users {
	return users{
		Find: func(id userID) action.Action[user] {
			return action.Do(func(context.Context) (user, error) {
				if loads != nil {
					*loads++
				}
				return user{ID: id, Name: name, Active: true}, nil
			})
		},
	}
}

func loadUser(dependency users) workflow.Workflow[userID, user] {
	return workflow.From(dependency.Find)
}

func TestWorkflowAppliesWithoutEffects(t *testing.T) {
	t.Parallel()

	loads := 0
	subject := loadUser(fixedUsers("ada", &loads))

	program := subject(firstID)
	if loads != 0 {
		t.Fatalf("loads = %d, want 0: applying a workflow must not execute", loads)
	}

	loaded, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loaded.ID != firstID || loaded.Name != "ada" {
		t.Fatalf("Run() user = %+v, want {ID:1 Name:ada}", loaded)
	}
	if loads != 1 {
		t.Fatalf("loads = %d, want 1", loads)
	}
}

func TestWorkflowReuseAcrossInputs(t *testing.T) {
	t.Parallel()

	subject := loadUser(fixedUsers("ada", nil))

	first, err := action.Run(t.Context(), subject(firstID))
	if err != nil {
		t.Fatalf("first Run() error = %v, want nil", err)
	}
	second, err := action.Run(t.Context(), subject(secondID))
	if err != nil {
		t.Fatalf("second Run() error = %v, want nil", err)
	}

	if first.ID != firstID || second.ID != secondID {
		t.Fatalf("ids = (%d, %d), want (1, 2)", first.ID, second.ID)
	}
}

func TestWorkflowDependencyCaptureIsStable(t *testing.T) {
	t.Parallel()

	dependency := fixedUsers("captured", nil)
	subject := loadUser(dependency)

	// Reassigning the local after construction must not affect the workflow:
	// the dependency graph is fixed by the closure, never resolved late.
	dependency = fixedUsers("replaced", nil)

	loaded, err := action.Run(t.Context(), subject(firstID))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loaded.Name != "captured" {
		t.Fatalf("Run() name = %q, want %q", loaded.Name, "captured")
	}
	if dependency.Find == nil {
		t.Fatal("guard: the replacement dependency should still be usable")
	}
}

func TestConstructorsPanicOnNilArguments(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow: nil function", func() {
		_ = workflow.From[userID, user](nil)
	})
}

func assertPanics(t *testing.T, want string, call func()) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("call did not panic, want panic %q", want)
		}
		message, ok := got.(string)
		if !ok || message != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()

	call()
}

type contextKey struct{}
