// Command withdraw is a memory-backed demo of the withdraw workflow example:
// one pure decision per case, committed — or refused — through a transactional
// boundary that carries an outbox alongside the state change. It seeds one
// account, runs an immediate withdrawal, a withdrawal held for approval, and a
// rejected one, and prints what the store committed after each case. No
// database is required.
//
// Run it from this directory with: go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
)

const (
	demoAccountID    = AccountID("acc-ada")
	demoStartBalance = Money(200_000)
	demoClockStart   = "2026-01-15T09:00:00Z"
)

// The three demo commands: one the domain applies immediately, one the policy
// holds for approval, one the requirements refuse outright.
const (
	caseImmediateAmount = Money(120_00)
	caseApprovalAmount  = Money(1_500_00)
	caseRejectedAmount  = Money(250_000)
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "withdraw demo failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	start, err := time.Parse(time.RFC3339, demoClockStart)
	if err != nil {
		return fmt.Errorf("parse demo clock: %w", err)
	}

	store := NewStore(Account{ID: demoAccountID, Balance: demoStartBalance})
	fmt.Printf("seeded %s with balance %d\n\n", demoAccountID, demoStartBalance)

	withdraw := NewWithdrawWorkflow(store.Accounts(), store.Boundary(), demoEncoder(stepClock(start)))

	if err := runImmediateCase(ctx, withdraw, store); err != nil {
		return err
	}
	fmt.Println()

	if err := runApprovalCase(ctx, withdraw, store); err != nil {
		return err
	}
	fmt.Println()

	return runRejectedCase(ctx, withdraw, store)
}

// demoEncoder fixes the identifiers so the printed outbox is deterministic;
// stepClock advances one minute per message for the same reason.
func demoEncoder(clock Clock) EventEncoder {
	issued := 0
	return NewJSONEncoder(clock, func() MessageID {
		issued++
		return MessageID(fmt.Sprintf("msg-%d", issued))
	})
}

func stepClock(start time.Time) Clock {
	tick := 0
	return func() time.Time {
		tick++
		return start.Add(time.Duration(tick) * time.Minute)
	}
}

func runImmediateCase(ctx context.Context, withdraw workflow.Workflow[WithdrawCommand, Receipt], store *Store) error {
	mark := markState(store)
	fmt.Printf("case 1: withdraw %d from %s (immediate)\n", caseImmediateAmount, demoAccountID)

	receipt, err := action.Run(ctx, withdraw(WithdrawCommand{AccountID: demoAccountID, Amount: caseImmediateAmount}))
	if err != nil {
		return err
	}
	fmt.Printf("  ok      balance=%d mode=%s\n", receipt.Balance, receipt.Mode)
	printState(store, mark)
	return nil
}

func runApprovalCase(ctx context.Context, withdraw workflow.Workflow[WithdrawCommand, Receipt], store *Store) error {
	mark := markState(store)
	fmt.Printf("case 2: withdraw %d from %s (above the %d approval threshold)\n",
		caseApprovalAmount, demoAccountID, approvalThreshold)

	receipt, err := action.Run(ctx, withdraw(WithdrawCommand{AccountID: demoAccountID, Amount: caseApprovalAmount}))
	if err != nil {
		return err
	}
	fmt.Printf("  held    balance=%d mode=%s (no debit until a human approves)\n", receipt.Balance, receipt.Mode)
	printState(store, mark)
	return nil
}

func runRejectedCase(ctx context.Context, withdraw workflow.Workflow[WithdrawCommand, Receipt], store *Store) error {
	mark := markState(store)
	fmt.Printf("case 3: withdraw %d from %s (more than the balance)\n", caseRejectedAmount, demoAccountID)

	// The domain refused, so Run fails with the domain's error and the trace
	// gains only the read: no transaction was ever opened.
	_, err := action.Run(ctx, withdraw(WithdrawCommand{AccountID: demoAccountID, Amount: caseRejectedAmount}))
	if !errors.Is(err, ErrInsufficientBalance) {
		return fmt.Errorf("case 3: expected insufficient balance, got %w", err)
	}
	fmt.Printf("  refused %v\n", err)
	fmt.Printf("  trace   %s (no transaction was opened)\n", traceSince(store, mark.trace))
	return nil
}

// stateMark remembers how much committed state a case started with, so the
// demo prints only the work that case caused.
type stateMark struct {
	trace  int
	outbox int
}

func markState(store *Store) stateMark {
	_, outbox := store.Snapshot(demoAccountID)
	return stateMark{trace: len(traceEntries(store)), outbox: len(outbox)}
}

func printState(store *Store, mark stateMark) {
	fmt.Printf("  trace   %s\n", traceSince(store, mark.trace))
	_, outbox := store.Snapshot(demoAccountID)
	for _, message := range outbox[mark.outbox:] {
		fmt.Printf("  outbox  %s %s %s %s\n", message.ID, message.Type, message.AggregateID, message.Payload)
	}
}

func traceEntries(store *Store) []string {
	joined := store.Trace()
	if joined == "" {
		return nil
	}
	return strings.Split(joined, ",")
}

func traceSince(store *Store, before int) string {
	entries := traceEntries(store)
	if before >= len(entries) {
		return "(none)"
	}
	return strings.Join(entries[before:], ",")
}
