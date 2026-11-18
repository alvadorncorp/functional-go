package transaction_test

import (
	"context"
	"errors"
	"testing"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// readAccounts is what the application holds OUTSIDE a transaction: it can load
// and nothing else. The write capability exists only inside a scope, so the
// type system — not a convention — keeps writes inside the boundary.
type readAccounts struct {
	Load func(int) action.Action[int]
}

func TestReadCapabilityOutsideWriteCapabilityInside(t *testing.T) {
	t.Parallel()

	stored := 0
	outside := readAccounts{
		Load: func(int) action.Action[int] { return action.Pure(savedAmount) },
	}

	inside := func(_ context.Context, fn func(context.Context, accounts) error) error {
		return fn(t.Context(), accounts{
			Save: func(amount int) action.Action[struct{}] {
				return action.Do(func(context.Context) (struct{}, error) {
					stored = amount
					return struct{}{}, nil
				})
			},
		})
	}

	program := action.FlatMap(outside.Load(1), func(balance int) action.Action[int] {
		return transaction.Within(inside, func(writable accounts) action.Action[int] {
			return action.Map(writable.Save(balance), func(struct{}) int { return balance })
		})
	})

	value, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != savedAmount || stored != savedAmount {
		t.Fatalf("value = %d, stored = %d, want both %d", value, stored, savedAmount)
	}
}

func TestWithinScopedCapabilitiesAreNotReachableFromContext(t *testing.T) {
	t.Parallel()

	resource := newResource()

	program := transaction.Within(resource.boundary(), func(s scope) action.Action[bool] {
		return action.Do(func(ctx context.Context) (bool, error) {
			// The scope arrives as a typed argument. Nothing in this module puts
			// it — or the handle behind it — into the context, so a workflow
			// cannot fish a transaction out of ctx.
			if s.Accounts.Save == nil {
				return false, errors.New("scope was not provided")
			}
			return ctx.Value(scope{}) == nil, nil
		})
	})

	reachable, err := action.Run(t.Context(), program)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if !reachable {
		t.Fatal("the scope was reachable through the context, want unreachable")
	}
}

func TestTwoBoundariesAreTwoIndependentTransactions(t *testing.T) {
	t.Parallel()

	first := newResource()
	second := newResource()
	second.commitErr = errCommit

	program := action.FlatMap(
		transaction.Within(first.boundary(), func(s scope) action.Action[struct{}] {
			return s.Accounts.Save(savedAmount)
		}),
		func(struct{}) action.Action[struct{}] {
			return transaction.Within(second.boundary(), func(s scope) action.Action[struct{}] {
				return s.Accounts.Save(savedAmount)
			})
		},
	)

	if _, err := action.Run(t.Context(), program); !errors.Is(err, errCommit) {
		t.Fatalf("Run() error = %v, want errCommit", err)
	}

	// This is the outcome the module refuses to hide: the first resource
	// committed and the second rolled back. Sequencing two boundaries composes
	// effects; it does not make them atomic together.
	if first.journal.trace() != "begin,save,commit" {
		t.Fatalf("first trace = %q, want %q", first.journal.trace(), "begin,save,commit")
	}
	if second.journal.trace() != "begin,save,rollback" {
		t.Fatalf("second trace = %q, want %q", second.journal.trace(), "begin,save,rollback")
	}
}
