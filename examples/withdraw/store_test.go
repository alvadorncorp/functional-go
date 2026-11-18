package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.alvadorn.org/functional/workflow/action"
)

// TestOutboxFailureRollsBackCommit pins the guarantee that makes the outbox
// safe: when appending an outbox message fails, the state change staged in the
// same transaction is undone too, so no debit exists without its message and
// no message exists without its debit.
func TestOutboxFailureRollsBackCommit(t *testing.T) {
	outboxDown := errors.New("outbox unavailable")
	store := NewStore(Account{ID: demoAccountID, Balance: demoStartBalance})
	seeded, _ := store.Snapshot(demoAccountID)

	store.FailAppendAt(0, outboxDown)

	encode := NewJSONEncoder(
		func() time.Time { return time.Unix(0, 0).UTC() },
		func() MessageID { return "msg-test" },
	)
	withdraw := NewWithdrawWorkflow(store.Accounts(), store.Boundary(), encode)

	_, err := action.Run(context.Background(), withdraw(WithdrawCommand{
		AccountID: demoAccountID,
		Amount:    caseImmediateAmount, // immediate path: save, then append
	}))
	if !errors.Is(err, outboxDown) {
		t.Fatalf("outbox failure: got %v, want %v", err, outboxDown)
	}

	account, outbox := store.Snapshot(demoAccountID)
	if account != seeded {
		t.Fatalf("rollback: account changed to %+v, want the seeded %+v", account, seeded)
	}
	if len(outbox) != 0 {
		t.Fatalf("rollback: outbox holds %d messages, want none", len(outbox))
	}

	want := "load,begin,save,append-failed,rollback"
	if got := store.Trace(); got != want {
		t.Fatalf("rollback: trace %q, want %q", got, want)
	}
}

// TestDomainRejectionOpensNoTransaction pins the other half of the contract: a
// pure decision that fails never reaches the resource manager, so a rejected
// withdrawal leaves no trace beyond the read that fed it.
func TestDomainRejectionOpensNoTransaction(t *testing.T) {
	store := NewStore(Account{ID: demoAccountID, Balance: demoStartBalance})

	encode := NewJSONEncoder(
		func() time.Time { return time.Unix(0, 0).UTC() },
		func() MessageID { return "msg-test" },
	)
	withdraw := NewWithdrawWorkflow(store.Accounts(), store.Boundary(), encode)

	_, err := action.Run(context.Background(), withdraw(WithdrawCommand{
		AccountID: demoAccountID,
		Amount:    caseRejectedAmount, // larger than the balance
	}))
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("rejection: got %v, want %v", err, ErrInsufficientBalance)
	}

	account, outbox := store.Snapshot(demoAccountID)
	if account.Balance != demoStartBalance {
		t.Fatalf("rejection: balance changed to %d, want %d", account.Balance, demoStartBalance)
	}
	if len(outbox) != 0 {
		t.Fatalf("rejection: outbox holds %d messages, want none", len(outbox))
	}

	if got, want := store.Trace(), "load"; got != want {
		t.Fatalf("rejection: trace %q, want %q", got, want)
	}
}
