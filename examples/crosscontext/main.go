// Command crosscontext is a memory-backed demo of the crosscontext example:
// one money transfer seen by two bounded contexts, Banking and Ledger, that
// share neither model nor vocabulary, joined only by one pure anti-corruption
// translation. The default mode runs both contexts on ONE resource manager,
// so the Banking state, the Ledger state, and the integration outbox commit
// in a single ACID transaction — or, when a write fails mid-transaction, not
// at all. With -mode=distributed each context keeps its own resource manager:
// Banking commits its state plus one outbox message, a relay hands the
// message to Ledger at-least-once, and Ledger's own inbox deduplicates a
// redelivered message. No database is required.
//
// Run it from this directory with: go run .
// Or with: go run . -mode=distributed
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
)

const (
	demoAccountID   = AccountID("acc-source")
	demoCounterpart = AccountID("acc-destination")
	demoBookID      = BookID("book-main")
	demoReference   = "transfer-demo"
)

// The demo amounts, in minor units so arithmetic stays exact.
const (
	demoOpeningBalance = Money(100_000) // 1000.00
	demoTransferSize   = Money(25_000)  // 250.00
)

func main() {
	mode := flag.String("mode", "shared", "shared | distributed")
	flag.Parse()

	if err := run(context.Background(), *mode); err != nil {
		fmt.Fprintf(os.Stderr, "crosscontext demo failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, mode string) error {
	switch mode {
	case "shared":
		return runShared(ctx)
	case "distributed":
		return runDistributed(ctx)
	default:
		return fmt.Errorf("unknown mode %q (want shared or distributed)", mode)
	}
}

// transferCommand is the demo's one command. Its reference is the idempotency
// key that follows the operation across both contexts.
func transferCommand() TransferCommand {
	return TransferCommand{
		Reference: demoReference,
		From:      demoAccountID,
		To:        demoCounterpart,
		Amount:    demoTransferSize,
	}
}

func runShared(ctx context.Context) error {
	store := NewStore(
		NewAccount(demoAccountID, demoOpeningBalance, 1),
		NewBook(demoBookID, 1),
	)
	transfer := NewTransferWorkflow(store.Accounts(), store.Books(), demoBookID, store.Boundary(), NewJSONEncoder())

	fmt.Println("shared: one resource manager commits Banking, Ledger, and the outbox together")
	fmt.Println()

	if err := runSharedTransfer(ctx, transfer, store); err != nil {
		return err
	}
	fmt.Println()

	return runSharedRollback(ctx, transfer, store)
}

func runSharedTransfer(ctx context.Context, transfer workflow.Workflow[TransferCommand, Receipt], store *Store) error {
	mark := markSharedState(store)
	fmt.Printf("case 1: transfer %d from %s to %s\n", demoTransferSize, demoAccountID, demoCounterpart)

	receipt, err := action.Run(ctx, transfer(transferCommand()))
	if err != nil {
		return err
	}
	fmt.Printf("  ok      receipt balance=%d book_entries=%d\n", receipt.SourceBalance, receipt.BookEntries)
	printSharedState(store, mark)
	return nil
}

func runSharedRollback(ctx context.Context, transfer workflow.Workflow[TransferCommand, Receipt], store *Store) error {
	mark := markSharedState(store)
	fmt.Println("case 2: the same transfer again, with the ledger write failing mid-transaction")

	store.FailNextBookSave(errors.New("ledger store: disk full"))

	_, err := action.Run(ctx, transfer(transferCommand()))
	if err == nil {
		return errors.New("case 2: expected the transfer to fail")
	}
	fmt.Printf("  failed  %v\n", err)
	printSharedState(store, mark)
	return nil
}

// sharedMark remembers how much committed state a case started with, so the
// demo prints only the work that case caused.
type sharedMark struct {
	trace  int
	outbox int
}

func markSharedState(store *Store) sharedMark {
	_, _, outbox := store.Snapshot(demoAccountID, demoBookID)
	return sharedMark{trace: len(traceEntries(store)), outbox: len(outbox)}
}

func printSharedState(store *Store, mark sharedMark) {
	account, book, outbox := store.Snapshot(demoAccountID, demoBookID)
	fmt.Printf("  state   balance=%d entries=%d\n", account.Balance, book.Entries)
	for _, message := range outbox[mark.outbox:] {
		fmt.Printf("  outbox  %s %s\n", message.Type, message.Reference)
	}
	fmt.Printf("  trace   %s\n", traceSince(store, mark.trace))
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

func runDistributed(ctx context.Context) error {
	bankingStore := NewBankingStore(NewAccount(demoAccountID, demoOpeningBalance, 1))
	ledgerStore := NewLedgerStore(NewBook(demoBookID, 1))
	transfer := NewBankingTransferWorkflow(bankingStore.Accounts(), bankingStore.Boundary())
	record := NewLedgerRecordWorkflow(ledgerStore.Books(), demoBookID, ledgerStore.Boundary())

	fmt.Println("distributed: separate resource managers, joined by an at-least-once relay")
	fmt.Println()

	// Banking commits its own state together with the integration message.
	// No broker call happens inside the transaction; delivery happens from
	// the outbox, and only the integration fact enters it — the internal
	// debit stays inside Banking.
	if _, err := action.Run(ctx, transfer(transferCommand())); err != nil {
		return fmt.Errorf("case 1: banking transfer: %w", err)
	}
	balance, pending := bankingStore.Snapshot(demoAccountID)
	if len(pending) != 1 {
		return fmt.Errorf("case 1: expected one outbox message, got %d", len(pending))
	}
	fmt.Printf("case 1: transfer %d committed in the Banking resource manager\n", demoTransferSize)
	fmt.Printf("  ok      balance=%d outbox=%d (%s)\n", balance, len(pending), pending[0].Type)
	fmt.Println()

	return runRelayCases(ctx, record, ledgerStore, pending)
}

func runRelayCases(
	ctx context.Context,
	record workflow.Workflow[TransferOccurred, LedgerReceipt],
	ledgerStore *LedgerStore,
	pending []IntegrationMessage,
) error {
	// Delivery is at-least-once: the first attempt reaches a broken ledger
	// database, the retry succeeds, and a duplicate delivery is skipped by
	// the consumer's inbox.
	ledgerStore.FailNextBookSave(errors.New("ledger store: connection reset"))

	fmt.Println("case 2: the relay delivers the pending message (the ledger database is down)")
	before := markLedgerState(ledgerStore)
	_, err := action.Run(ctx, Relay(pending, record))
	if err == nil {
		return errors.New("case 2: expected the relay to fail")
	}
	fmt.Printf("  failed  %v\n", err)
	printLedgerState(ledgerStore, before)
	fmt.Println()

	fmt.Println("case 3: the relay retries the same pending message")
	before = markLedgerState(ledgerStore)
	receipts, err := action.Run(ctx, Relay(pending, record))
	if err != nil {
		return fmt.Errorf("case 3: relay: %w", err)
	}
	printLedgerReceipts(receipts)
	printLedgerState(ledgerStore, before)
	fmt.Println()

	fmt.Println("case 4: the relay redelivers the same message once more (at-least-once)")
	before = markLedgerState(ledgerStore)
	receipts, err = action.Run(ctx, Relay(pending, record))
	if err != nil {
		return fmt.Errorf("case 4: relay: %w", err)
	}
	printLedgerReceipts(receipts)
	printLedgerState(ledgerStore, before)
	return nil
}

// ledgerMark remembers how much committed state a case started with, so the
// demo prints only the work that case caused.
type ledgerMark struct {
	entries uint64
	debits  Amount
	inbox   int
	outbox  int
}

func markLedgerState(store *LedgerStore) ledgerMark {
	book, inboxSize, outbox := store.Snapshot(demoBookID)
	return ledgerMark{entries: book.Entries, debits: book.Debits, inbox: inboxSize, outbox: len(outbox)}
}

func printLedgerState(store *LedgerStore, mark ledgerMark) {
	book, inboxSize, outbox := store.Snapshot(demoBookID)
	fmt.Printf("  changes entries=+%d debits=+%d inbox=+%d outbox=+%d\n",
		book.Entries-mark.entries, book.Debits-mark.debits, inboxSize-mark.inbox, len(outbox)-mark.outbox)
}

func printLedgerReceipts(receipts []LedgerReceipt) {
	for i := range receipts {
		fmt.Printf("  ok      receipt book_entries=%d skipped=%t\n", receipts[i].BookEntries, receipts[i].Skipped)
	}
}
