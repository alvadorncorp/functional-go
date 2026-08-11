package workflow_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow"
	"github.com/alvadorncorp/gopackages/workflow/action"
)

var errInactive = errors.New("user is not active")

func requireActive() workflow.Workflow[user, user] {
	return workflow.MapFallible(
		workflow.From(action.Pure[user]),
		func(loaded user) (user, error) {
			if !loaded.Active {
				return user{}, errInactive
			}
			return loaded, nil
		},
	)
}

func TestCompositionProducesAReusableWorkflow(t *testing.T) {
	t.Parallel()

	// LoadUser + RequireActive = LoadActiveUser, itself an ordinary Workflow
	// that larger workflows compose without knowing how it was built.
	loadActiveUser := workflow.Then(loadUser(fixedUsers("ada", nil)), requireActive())

	greet := workflow.Map(loadActiveUser, func(loaded user) string {
		return "hello " + loaded.Name
	})

	value, err := action.Run(t.Context(), greet(firstID))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != "hello ada" {
		t.Fatalf("Run() value = %q, want %q", value, "hello ada")
	}
}

func TestCompositionExecutesNothingUntilRun(t *testing.T) {
	t.Parallel()

	loads := 0
	composed := workflow.Then(loadUser(fixedUsers("ada", &loads)), requireActive())
	mapped := workflow.Map(composed, func(loaded user) string { return loaded.Name })

	program := mapped(firstID)
	if loads != 0 {
		t.Fatalf("loads = %d, want 0 before Run", loads)
	}

	if _, err := action.Run(t.Context(), program); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loads != 1 {
		t.Fatalf("loads = %d, want 1", loads)
	}
}

func TestThenRunsStagesInOrder(t *testing.T) {
	t.Parallel()

	var order []string
	first := workflow.From(func(id userID) action.Action[userID] {
		return action.Do(func(context.Context) (userID, error) {
			order = append(order, "first")
			return id, nil
		})
	})
	second := workflow.From(func(id userID) action.Action[string] {
		return action.Do(func(context.Context) (string, error) {
			order = append(order, "second")
			return "done", nil
		})
	})

	if _, err := action.Run(t.Context(), workflow.Then(first, second)(firstID)); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	if strings.Join(order, ",") != "first,second" {
		t.Fatalf("order = %v, want [first second]", order)
	}
}

func TestThenShortCircuitsOnFirstFailure(t *testing.T) {
	t.Parallel()

	secondRuns := 0
	failing := workflow.From(func(userID) action.Action[user] { return action.Fail[user](errNotFound) })
	second := workflow.From(func(user) action.Action[string] {
		secondRuns++
		return action.Pure("unreachable")
	})

	_, err := action.Run(t.Context(), workflow.Then(failing, second)(firstID))
	if !errors.Is(err, errNotFound) {
		t.Fatalf("Run() error = %v, want errNotFound", err)
	}
	if secondRuns != 0 {
		t.Fatalf("second stage runs = %d, want 0", secondRuns)
	}
}

func TestMapFallibleShortCircuitsTheDownstream(t *testing.T) {
	t.Parallel()

	inactive := users{
		Find: func(id userID) action.Action[user] {
			return action.Pure(user{ID: id, Name: "ada", Active: false})
		},
	}

	subject := workflow.Then(loadUser(inactive), requireActive())

	if _, err := action.Run(t.Context(), subject(firstID)); !errors.Is(err, errInactive) {
		t.Fatalf("Run() error = %v, want errInactive", err)
	}
}

func TestContramapAdaptsTheInput(t *testing.T) {
	t.Parallel()

	byID := loadUser(fixedUsers("ada", nil))
	byRequest := workflow.Contramap(byID, func(raw string) userID {
		if raw == "first" {
			return firstID
		}
		return secondID
	})

	loaded, err := action.Run(t.Context(), byRequest("first"))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loaded.ID != firstID {
		t.Fatalf("Run() id = %d, want %d", loaded.ID, firstID)
	}
}

func TestFlatMapSequencesADependentAction(t *testing.T) {
	t.Parallel()

	subject := workflow.FlatMap(loadUser(fixedUsers("ada", nil)), func(loaded user) action.Action[string] {
		return action.Pure(strings.ToUpper(loaded.Name))
	})

	value, err := action.Run(t.Context(), subject(firstID))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != "ADA" {
		t.Fatalf("Run() value = %q, want %q", value, "ADA")
	}
}

func TestTapKeepsTheOutput(t *testing.T) {
	t.Parallel()

	observed := ""
	subject := workflow.Tap(loadUser(fixedUsers("ada", nil)), func(loaded user) action.Action[struct{}] {
		return action.Do(func(context.Context) (struct{}, error) {
			observed = loaded.Name
			return struct{}{}, nil
		})
	})

	loaded, err := action.Run(t.Context(), subject(firstID))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loaded.Name != "ada" || observed != "ada" {
		t.Fatalf("value = %q, observed = %q, want both %q", loaded.Name, observed, "ada")
	}
}

func TestRecoverSubstitutesAFallback(t *testing.T) {
	t.Parallel()

	missing := users{
		Find: func(userID) action.Action[user] { return action.Fail[user](errNotFound) },
	}

	subject := workflow.Recover(loadUser(missing), func(error) action.Action[user] {
		return action.Pure(user{Name: "anonymous", Active: true})
	})

	loaded, err := action.Run(t.Context(), subject(firstID))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if loaded.Name != "anonymous" {
		t.Fatalf("Run() name = %q, want %q", loaded.Name, "anonymous")
	}
}

func TestComposeCombinatorsPanicOnNilArguments(t *testing.T) {
	t.Parallel()

	valid := loadUser(fixedUsers("ada", nil))

	assertPanics(t, "workflow: nil workflow", func() {
		_ = workflow.Map(workflow.Workflow[userID, user](nil), func(loaded user) string { return loaded.Name })
	})
	assertPanics(t, "workflow: nil mapper", func() {
		_ = workflow.Map[userID, user, string](valid, nil)
	})
	assertPanics(t, "workflow: nil workflow", func() {
		_ = workflow.Then(valid, workflow.Workflow[user, string](nil))
	})
	assertPanics(t, "workflow: nil mapper", func() {
		_ = workflow.Contramap[string, userID, user](valid, nil)
	})
	assertPanics(t, "workflow: nil binder", func() {
		_ = workflow.FlatMap[userID, user, string](valid, nil)
	})
	assertPanics(t, "workflow: nil observer", func() {
		_ = workflow.Tap[userID, user](valid, nil)
	})
	assertPanics(t, "workflow: nil recovery", func() {
		_ = workflow.Recover[userID, user](valid, nil)
	})
}

func TestApplyingAWorkflowThatReturnsNilPanics(t *testing.T) {
	t.Parallel()

	broken := workflow.Workflow[userID, user](func(userID) action.Action[user] { return nil })

	assertPanics(t, "workflow: nil action", func() {
		_ = workflow.Map(broken, func(loaded user) string { return loaded.Name })(firstID)
	})
}
