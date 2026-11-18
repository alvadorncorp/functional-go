package transaction_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

var (
	errSave   = errors.New("save failed")
	errCommit = errors.New("commit failed")
)

const savedAmount = 250

// journal records the lifecycle a boundary performs, so a test can assert the
// exact order of begin, write, commit, and rollback.
type journal struct {
	events []string
}

func (j *journal) record(event string) {
	j.events = append(j.events, event)
}

func (j *journal) trace() string {
	return strings.Join(j.events, ",")
}

// accounts is a transaction-bound capability: it exists only inside a scope and
// is therefore impossible to call outside the boundary that created it.
type accounts struct {
	Save func(int) action.Action[struct{}]
}

// scope is the application-defined set of capabilities bound to one concrete
// transaction. The transaction package never names this type.
type scope struct {
	Accounts accounts
}

type fakeResource struct {
	journal    *journal
	saveErr    error
	commitErr  error
	lastSaved  int
	scopeCount int
}

func (r *fakeResource) boundary() transaction.Boundary[scope] {
	return func(ctx context.Context, fn func(context.Context, scope) error) error {
		r.journal.record("begin")
		r.scopeCount++

		committed := false
		defer func() {
			if !committed {
				r.journal.record("rollback")
			}
		}()

		if err := fn(ctx, r.newScope()); err != nil {
			return err
		}
		if r.commitErr != nil {
			return r.commitErr
		}

		r.journal.record("commit")
		committed = true
		return nil
	}
}

func (r *fakeResource) newScope() scope {
	return scope{
		Accounts: accounts{
			Save: func(amount int) action.Action[struct{}] {
				return action.Do(func(context.Context) (struct{}, error) {
					if r.saveErr != nil {
						r.journal.record("save-failed")
						return struct{}{}, r.saveErr
					}
					r.journal.record("save")
					r.lastSaved = amount
					return struct{}{}, nil
				})
			},
		},
	}
}

func newResource() *fakeResource {
	return &fakeResource{journal: &journal{}}
}

func TestWithinCommitsOnSuccess(t *testing.T) {
	t.Parallel()

	resource := newResource()

	program := transaction.Within(resource.boundary(), func(s scope) action.Action[string] {
		return action.Map(s.Accounts.Save(savedAmount), func(struct{}) string { return "receipt" })
	})

	if resource.journal.trace() != "" {
		t.Fatalf("trace = %q, want empty before Run", resource.journal.trace())
	}

	value, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != "receipt" {
		t.Fatalf("Run() value = %q, want %q", value, "receipt")
	}
	if resource.journal.trace() != "begin,save,commit" {
		t.Fatalf("trace = %q, want %q", resource.journal.trace(), "begin,save,commit")
	}
	if resource.lastSaved != savedAmount {
		t.Fatalf("lastSaved = %d, want %d", resource.lastSaved, savedAmount)
	}
}

func TestWithinRollsBackOnError(t *testing.T) {
	t.Parallel()

	resource := newResource()
	resource.saveErr = errSave

	program := transaction.Within(resource.boundary(), func(s scope) action.Action[string] {
		return action.Map(s.Accounts.Save(savedAmount), func(struct{}) string { return "receipt" })
	})

	value, err := action.Run(t.Context(), program)
	if !errors.Is(err, errSave) {
		t.Fatalf("Run() error = %v, want errSave", err)
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
	if resource.journal.trace() != "begin,save-failed,rollback" {
		t.Fatalf("trace = %q, want %q", resource.journal.trace(), "begin,save-failed,rollback")
	}
}

func TestWithinPropagatesACommitFailure(t *testing.T) {
	t.Parallel()

	resource := newResource()
	resource.commitErr = errCommit

	program := transaction.Within(resource.boundary(), func(s scope) action.Action[string] {
		return action.Map(s.Accounts.Save(savedAmount), func(struct{}) string { return "receipt" })
	})

	value, err := action.Run(t.Context(), program)
	if err != errCommit {
		t.Fatalf("Run() error = %v, want unchanged errCommit", err)
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
	if resource.journal.trace() != "begin,save,rollback" {
		t.Fatalf("trace = %q, want %q", resource.journal.trace(), "begin,save,rollback")
	}
}

func TestWithinRollsBackThenRepanics(t *testing.T) {
	t.Parallel()

	resource := newResource()

	program := transaction.Within(resource.boundary(), func(scope) action.Action[string] {
		return action.Do(func(context.Context) (string, error) {
			panic("workflow/transaction_test: deliberate panic")
		})
	})

	func() {
		defer func() {
			recovered := recover()
			if recovered != "workflow/transaction_test: deliberate panic" {
				t.Fatalf("recovered = %v, want the original panic value", recovered)
			}
		}()
		_, _ = action.Run(t.Context(), program)
	}()

	if resource.journal.trace() != "begin,rollback" {
		t.Fatalf("trace = %q, want %q", resource.journal.trace(), "begin,rollback")
	}
}

func TestWithinPropagatesCancellation(t *testing.T) {
	t.Parallel()

	resource := newResource()

	program := transaction.Within(resource.boundary(), func(scope) action.Action[string] {
		return action.Do(func(ctx context.Context) (string, error) {
			return "", ctx.Err()
		})
	})

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := action.Run(ctx, program); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if resource.journal.trace() != "begin,rollback" {
		t.Fatalf("trace = %q, want %q", resource.journal.trace(), "begin,rollback")
	}
}

func TestWithinExecutesNothingUntilRun(t *testing.T) {
	t.Parallel()

	resource := newResource()

	_ = transaction.Within(resource.boundary(), func(s scope) action.Action[struct{}] {
		return s.Accounts.Save(savedAmount)
	})

	if resource.scopeCount != 0 {
		t.Fatalf("scopeCount = %d, want 0: constructing Within must not begin a transaction", resource.scopeCount)
	}
}

func TestWithinScopedCapabilitiesAreFreshPerRun(t *testing.T) {
	t.Parallel()

	resource := newResource()

	program := transaction.Within(resource.boundary(), func(s scope) action.Action[struct{}] {
		return s.Accounts.Save(savedAmount)
	})

	if _, err := action.Run(t.Context(), program); err != nil {
		t.Fatalf("first Run() error = %v, want nil", err)
	}
	if _, err := action.Run(t.Context(), program); err != nil {
		t.Fatalf("second Run() error = %v, want nil", err)
	}

	if resource.scopeCount != 2 {
		t.Fatalf("scopeCount = %d, want 2: each Run must receive its own scope", resource.scopeCount)
	}
	if resource.journal.trace() != "begin,save,commit,begin,save,commit" {
		t.Fatalf("trace = %q, want two full lifecycles", resource.journal.trace())
	}
}

func TestWithinPanicsWhenBoundarySkipsFunctionOnSuccess(t *testing.T) {
	t.Parallel()

	boundary := func(context.Context, func(context.Context, scope) error) error {
		return nil
	}
	called := false
	program := transaction.Within(boundary, func(scope) action.Action[string] {
		called = true
		return action.Pure("receipt")
	})

	assertPanics(t, "workflow/transaction: boundary returned without running fn", func() {
		_, _ = action.Run(t.Context(), program)
	})
	if called {
		t.Fatal("fn was called despite the boundary skipping its callback")
	}
}

func TestWithinPropagatesErrorWhenBoundarySkipsFunction(t *testing.T) {
	t.Parallel()

	boundaryErr := errors.New("begin failed")
	boundary := func(context.Context, func(context.Context, scope) error) error {
		return boundaryErr
	}
	called := false
	program := transaction.Within(boundary, func(scope) action.Action[string] {
		called = true
		return action.Pure("receipt")
	})

	value, err := action.Run(t.Context(), program)
	if err != boundaryErr {
		t.Fatalf("Run() error = %v, want unchanged boundary error", err)
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
	if called {
		t.Fatal("fn was called despite the boundary skipping its callback")
	}
}

func TestWithinPanicsOnNilArguments(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/transaction: nil boundary", func() {
		_ = transaction.Within(transaction.Boundary[scope](nil), func(scope) action.Action[int] {
			return action.Pure(0)
		})
	})
	assertPanics(t, "workflow/transaction: nil function", func() {
		_ = transaction.Within[scope, int](newResource().boundary(), nil)
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
