package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.alvadorn.org/functional/workflow/action"
)

// TestSharedTransactionRollsBackEverything pins the guarantee that makes the
// shared topology safe: when the Ledger write fails mid-transaction, the
// Banking write staged in the same transaction is undone too, so no state
// change exists without its outbox messages and neither store moves.
func TestSharedTransactionRollsBackEverything(t *testing.T) {
	store := NewStore(
		NewAccount(demoAccountID, demoOpeningBalance, 1),
		NewBook(demoBookID, 1),
	)
	transfer := NewTransferWorkflow(store.Accounts(), store.Books(), demoBookID, store.Boundary(), NewJSONEncoder())

	receipt, err := action.Run(context.Background(), transfer(transferCommand()))
	if err != nil {
		t.Fatalf("first transfer: %v", err)
	}
	if receipt.SourceBalance != demoOpeningBalance-demoTransferSize || receipt.BookEntries != 1 {
		t.Fatalf("first transfer: receipt %+v, want the debit applied and one entry", receipt)
	}

	ledgerDown := errors.New("ledger store: disk full")
	store.FailNextBookSave(ledgerDown)

	if _, err := action.Run(context.Background(), transfer(transferCommand())); !errors.Is(err, ledgerDown) {
		t.Fatalf("second transfer: got %v, want %v", err, ledgerDown)
	}

	account, book, outbox := store.Snapshot(demoAccountID, demoBookID)
	if account.Balance != demoOpeningBalance-demoTransferSize {
		t.Errorf("rollback: balance %d, want the first transfer's %d", account.Balance, demoOpeningBalance-demoTransferSize)
	}
	if book.Entries != 1 {
		t.Errorf("rollback: entries %d, want 1", book.Entries)
	}
	if len(outbox) != 3 {
		t.Errorf("rollback: outbox holds %d messages, want only the first transfer's 3", len(outbox))
	}

	// The failed transaction staged the banking write before the ledger write
	// refused; the trace must show that transaction rolling back, leaving the
	// first transfer's commit as the only one.
	joined := store.Trace()
	if !strings.HasSuffix(joined, "begin,banking-save,book-save-failed,rollback") {
		t.Errorf("rollback: trace %q, want it to end at a rolled-back transaction", joined)
	}
	if strings.Count(joined, ",commit") != 1 {
		t.Errorf("rollback: trace %q, want exactly one commit", joined)
	}
}

// TestDistributedRedeliveryAppliesTheLedgerOnce pins the consumer-side
// contract of at-least-once delivery: a failed delivery leaves resource
// manager B untouched, the retry applies the domain exactly once, and a
// duplicate delivery is skipped by the inbox instead of posting a second
// entry.
func TestDistributedRedeliveryAppliesTheLedgerOnce(t *testing.T) {
	bankingStore := NewBankingStore(NewAccount(demoAccountID, demoOpeningBalance, 1))
	ledgerStore := NewLedgerStore(NewBook(demoBookID, 1))
	transfer := NewBankingTransferWorkflow(bankingStore.Accounts(), bankingStore.Boundary())
	record := NewLedgerRecordWorkflow(ledgerStore.Books(), demoBookID, ledgerStore.Boundary())

	if _, err := action.Run(context.Background(), transfer(transferCommand())); err != nil {
		t.Fatalf("banking transfer: %v", err)
	}
	_, pending := bankingStore.Snapshot(demoAccountID)
	if len(pending) != 1 {
		t.Fatalf("banking outbox holds %d messages, want the single integration fact", len(pending))
	}

	ledgerDown := errors.New("ledger store: connection reset")
	ledgerStore.FailNextBookSave(ledgerDown)

	if _, err := action.Run(context.Background(), Relay(pending, record)); !errors.Is(err, ledgerDown) {
		t.Fatalf("first delivery: got %v, want %v", err, ledgerDown)
	}
	book, inboxSize, ledgerOutbox := ledgerStore.Snapshot(demoBookID)
	if book.Entries != 0 || inboxSize != 0 || len(ledgerOutbox) != 0 {
		t.Fatalf("failed delivery: entries=%d inbox=%d outbox=%d, want nothing applied",
			book.Entries, inboxSize, len(ledgerOutbox))
	}

	first, err := action.Run(context.Background(), Relay(pending, record))
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if len(first) != 1 || first[0].Skipped || first[0].BookEntries != 1 {
		t.Fatalf("retry: receipts %+v, want one receipt for the first application", first)
	}

	second, err := action.Run(context.Background(), Relay(pending, record))
	if err != nil {
		t.Fatalf("duplicate delivery: %v", err)
	}
	if len(second) != 1 || !second[0].Skipped {
		t.Fatalf("duplicate delivery: receipts %+v, want the inbox to skip it", second)
	}

	book, inboxSize, ledgerOutbox = ledgerStore.Snapshot(demoBookID)
	if book.Entries != 1 || book.Debits != Amount(demoTransferSize) || book.Credits != Amount(demoTransferSize) {
		t.Errorf("deduplication: book %+v, want exactly one balanced entry", book)
	}
	if inboxSize != 1 || len(ledgerOutbox) != 1 {
		t.Errorf("deduplication: inbox=%d outbox=%d, want one mark and one message", inboxSize, len(ledgerOutbox))
	}
}
