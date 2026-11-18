// The distributed topology: the SAME two contexts on SEPARATE resource
// managers. There is no transaction spanning both. Banking commits its own
// state together with an outbox message; that message later reaches Ledger,
// which commits its own state in its own transaction. Between the two commits
// the system is legitimately inconsistent, and it converges afterwards. That
// is eventual consistency, and calling it anything else would be a lie the
// type system cannot catch.
//
// Delivery is at-least-once, so the consumer owns idempotency: the inbox is
// an ordinary capability of this application, not something the workflow
// module provides.
package main

import (
	"context"
	"sync"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// IntegrationMessage is one integration message in flight between the two
// resource managers. It carries the Banking integration event as typed data;
// the shared topology instead serializes payloads, because the shape of a
// message is an application decision per deployment, not a library rule.
type IntegrationMessage struct {
	Type      string
	Reference string
	Event     TransferOccurred
}

// BankingReceipt is what the Banking side returns.
type BankingReceipt struct {
	SourceBalance Money
}

// LedgerReceipt is what the consumer side returns. Skipped marks a
// redelivery the inbox refused to apply twice.
type LedgerReceipt struct {
	BookEntries uint64
	Skipped     bool
}

// BankingTxOutbox appends an integration message inside the same Banking
// transaction as the state change it accompanies.
type BankingTxOutbox struct {
	Append func(IntegrationMessage) action.Action[struct{}]
}

// LedgerTxOutbox appends the Ledger's own downstream messages inside its
// transaction.
type LedgerTxOutbox struct {
	Append func(IntegrationMessage) action.Action[struct{}]
}

// TxInbox is the consumer's inbox: the set of references this context has
// already applied, written in the same transaction as the state it guards.
type TxInbox struct {
	Has  func(string) action.Action[bool]
	Mark func(string) action.Action[struct{}]
}

// BankingTxScope is resource manager A's transaction scope.
type BankingTxScope struct {
	Accounts TxAccounts
	Outbox   BankingTxOutbox
}

// LedgerTxScope is resource manager B's transaction scope. The book, the
// inbox mark, and the downstream messages commit together or not at all.
type LedgerTxScope struct {
	Books  TxBooks
	Inbox  TxInbox
	Outbox LedgerTxOutbox
}

// BankingStore is resource manager A: the Banking database, with its own
// outbox awaiting delivery.
type BankingStore struct {
	mu       sync.Mutex
	accounts map[AccountID]Account
	outbox   []IntegrationMessage
}

// NewBankingStore seeds resource manager A.
func NewBankingStore(account Account) *BankingStore {
	return &BankingStore{accounts: map[AccountID]Account{account.ID: account}}
}

// Accounts is the read capability.
func (s *BankingStore) Accounts() Accounts {
	return Accounts{
		Load: func(id AccountID) action.Action[Account] {
			return action.Do(func(ctx context.Context) (Account, error) {
				if err := ctx.Err(); err != nil {
					return Account{}, err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				account, found := s.accounts[id]
				if !found {
					return Account{}, ErrAccountNotFound
				}
				return account, nil
			})
		},
	}
}

// Snapshot returns the committed balance and the pending outbox.
func (s *BankingStore) Snapshot(id AccountID) (Money, []IntegrationMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accounts[id].Balance, append([]IntegrationMessage(nil), s.outbox...)
}

// pendingBanking holds uncommitted writes for one Boundary invocation.
type pendingBanking struct {
	account Account
	saved   bool
	outbox  []IntegrationMessage
}

// Boundary is resource manager A's transaction: the account write and the
// outbox append commit together or not at all.
func (s *BankingStore) Boundary() transaction.Boundary[BankingTxScope] {
	return func(ctx context.Context, fn func(context.Context, BankingTxScope) error) error {
		pending := &pendingBanking{}
		if err := fn(ctx, s.bankingScope(pending)); err != nil {
			return err
		}
		s.publish(pending)
		return nil
	}
}

func (s *BankingStore) bankingScope(pending *pendingBanking) BankingTxScope {
	return BankingTxScope{
		Accounts: TxAccounts{
			Save: func(account Account) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					pending.account, pending.saved = account, true
					return nil
				})
			},
		},
		Outbox: BankingTxOutbox{
			Append: func(message IntegrationMessage) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					pending.outbox = append(pending.outbox, message)
					return nil
				})
			},
		},
	}
}

func (s *BankingStore) publish(pending *pendingBanking) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pending.saved {
		s.accounts[pending.account.ID] = pending.account
	}
	s.outbox = append(s.outbox, pending.outbox...)
}

// LedgerStore is resource manager B — a different database entirely. It owns
// the inbox: the references this context has already applied.
type LedgerStore struct {
	mu          sync.Mutex
	books       map[BookID]Book
	processed   map[string]struct{}
	outbox      []IntegrationMessage
	bookSaveErr error
}

// NewLedgerStore seeds resource manager B.
func NewLedgerStore(book Book) *LedgerStore {
	return &LedgerStore{
		books:     map[BookID]Book{book.ID: book},
		processed: map[string]struct{}{},
	}
}

// FailNextBookSave makes exactly the next book save fail, then clears itself,
// the way a lost connection would; the relay's retry reaches a healthy
// database again.
func (s *LedgerStore) FailNextBookSave(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookSaveErr = err
}

// Books is the read capability.
func (s *LedgerStore) Books() Books {
	return Books{
		Load: func(id BookID) action.Action[Book] {
			return action.Do(func(ctx context.Context) (Book, error) {
				if err := ctx.Err(); err != nil {
					return Book{}, err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				book, found := s.books[id]
				if !found {
					return Book{}, ErrBookNotFound
				}
				return book, nil
			})
		},
	}
}

// Snapshot returns the committed book, the inbox size, and the pending outbox.
func (s *LedgerStore) Snapshot(id BookID) (Book, int, []IntegrationMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.books[id], len(s.processed), append([]IntegrationMessage(nil), s.outbox...)
}

// pendingLedger holds uncommitted writes for one Boundary invocation.
type pendingLedger struct {
	book   Book
	posted bool
	marks  []string
	outbox []IntegrationMessage
}

// Boundary is resource manager B's transaction, independent of A's.
func (s *LedgerStore) Boundary() transaction.Boundary[LedgerTxScope] {
	return func(ctx context.Context, fn func(context.Context, LedgerTxScope) error) error {
		pending := &pendingLedger{}
		if err := fn(ctx, s.ledgerScope(pending)); err != nil {
			return err
		}
		s.publish(pending)
		return nil
	}
}

func (s *LedgerStore) ledgerScope(pending *pendingLedger) LedgerTxScope {
	return LedgerTxScope{
		Books: TxBooks{
			Save: func(book Book) action.Action[struct{}] { return s.stageBook(pending, book) },
		},
		Inbox: TxInbox{
			Has: func(reference string) action.Action[bool] {
				return action.Do(func(ctx context.Context) (bool, error) {
					if err := ctx.Err(); err != nil {
						return false, err
					}
					s.mu.Lock()
					defer s.mu.Unlock()
					_, seen := s.processed[reference]
					return seen, nil
				})
			},
			Mark: func(reference string) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					pending.marks = append(pending.marks, reference)
					return nil
				})
			},
		},
		Outbox: LedgerTxOutbox{
			Append: func(message IntegrationMessage) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					pending.outbox = append(pending.outbox, message)
					return nil
				})
			},
		},
	}
}

func (s *LedgerStore) stageBook(pending *pendingLedger, book Book) action.Action[struct{}] {
	return action.Exec(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.takeBookSaveError(); err != nil {
			return err
		}
		pending.book, pending.posted = book, true
		return nil
	})
}

func (s *LedgerStore) takeBookSaveError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.bookSaveErr
	s.bookSaveErr = nil
	return err
}

func (s *LedgerStore) publish(pending *pendingLedger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pending.posted {
		s.books[pending.book.ID] = pending.book
	}
	for _, reference := range pending.marks {
		s.processed[reference] = struct{}{}
	}
	s.outbox = append(s.outbox, pending.outbox...)
}

// NewBankingTransferWorkflow commits the Banking state and the integration
// message in resource manager A's own transaction. It publishes nothing: a
// broker call inside a database transaction cannot be rolled back, so
// delivery happens afterwards, from the outbox.
func NewBankingTransferWorkflow(
	accounts Accounts,
	boundary transaction.Boundary[BankingTxScope],
) workflow.Workflow[TransferCommand, BankingReceipt] {
	load := workflow.From(accounts.Load)

	return func(cmd TransferCommand) action.Action[BankingReceipt] {
		return action.FlatMap(load(cmd.From), func(source Account) action.Action[BankingReceipt] {
			decision, err := DecideTransfer(source, cmd)
			if err != nil {
				return action.Fail[BankingReceipt](err)
			}
			return commitBankingTransfer(boundary, cmd.Reference, decision)
		})
	}
}

func commitBankingTransfer(
	boundary transaction.Boundary[BankingTxScope],
	reference string,
	decision modeling.Decision[Account, BankingEvent],
) action.Action[BankingReceipt] {
	return transaction.Within(boundary, func(scope BankingTxScope) action.Action[BankingReceipt] {
		return action.Do(func(ctx context.Context) (BankingReceipt, error) {
			if _, err := action.Run(ctx, scope.Accounts.Save(decision.State())); err != nil {
				return BankingReceipt{}, err
			}
			if err := appendIntegrationMessages(ctx, scope.Outbox, reference, decision); err != nil {
				return BankingReceipt{}, err
			}
			return BankingReceipt{SourceBalance: decision.State().Balance}, nil
		})
	})
}

// appendIntegrationMessages writes only the facts other contexts subscribe
// to. MoneyDebited is internal and stays inside Banking — the contrast with
// the shared topology, whose single outbox receives every emitted event.
func appendIntegrationMessages(
	ctx context.Context,
	outbox BankingTxOutbox,
	reference string,
	decision modeling.Decision[Account, BankingEvent],
) error {
	for event := range decision.Events() {
		occurred, ok := event.(TransferOccurred)
		if !ok {
			continue
		}
		message := IntegrationMessage{Type: occurred.EventName(), Reference: reference, Event: occurred}
		if _, err := action.Run(ctx, outbox.Append(message)); err != nil {
			return err
		}
	}
	return nil
}

// NewLedgerRecordWorkflow consumes a Banking integration event and commits in
// resource manager B's own transaction.
//
// Its input is the integration event, not a Banking account: the two contexts
// share a message, never a model. It checks the inbox first because delivery
// is at-least-once, so being handed the same message twice is normal rather
// than exceptional.
func NewLedgerRecordWorkflow(
	books Books,
	bookID BookID,
	boundary transaction.Boundary[LedgerTxScope],
) workflow.Workflow[TransferOccurred, LedgerReceipt] {
	return func(occurred TransferOccurred) action.Action[LedgerReceipt] {
		return action.FlatMap(books.Load(bookID), func(book Book) action.Action[LedgerReceipt] {
			decision, err := DecideRecord(book, ToLedgerCommand(occurred))
			if err != nil {
				return action.Fail[LedgerReceipt](err)
			}
			return commitLedgerRecord(boundary, occurred.Reference, book.Entries, decision)
		})
	}
}

// commitLedgerRecord opens resource manager B's transaction and asks the
// inbox whether this reference was already applied.
func commitLedgerRecord(
	boundary transaction.Boundary[LedgerTxScope],
	reference string,
	currentEntries uint64,
	decision modeling.Decision[Book, LedgerEvent],
) action.Action[LedgerReceipt] {
	return transaction.Within(boundary, func(scope LedgerTxScope) action.Action[LedgerReceipt] {
		return action.Do(func(ctx context.Context) (LedgerReceipt, error) {
			seen, err := action.Run(ctx, scope.Inbox.Has(reference))
			if err != nil {
				return LedgerReceipt{}, err
			}
			if seen {
				// A redelivery is a normal event under at-least-once
				// delivery. The domain is not applied a second time.
				return LedgerReceipt{BookEntries: currentEntries, Skipped: true}, nil
			}
			return persistLedgerRecord(ctx, scope, reference, decision)
		})
	})
}

// persistLedgerRecord commits the book, the inbox mark, and the Ledger's own
// outbox in the one transaction: the same all-or-nothing discipline as the
// shared topology, applied inside each resource manager.
func persistLedgerRecord(
	ctx context.Context,
	scope LedgerTxScope,
	reference string,
	decision modeling.Decision[Book, LedgerEvent],
) (LedgerReceipt, error) {
	if _, err := action.Run(ctx, scope.Books.Save(decision.State())); err != nil {
		return LedgerReceipt{}, err
	}
	if _, err := action.Run(ctx, scope.Inbox.Mark(reference)); err != nil {
		return LedgerReceipt{}, err
	}
	if err := appendLedgerMessages(ctx, scope.Outbox, reference, decision); err != nil {
		return LedgerReceipt{}, err
	}
	return LedgerReceipt{BookEntries: decision.State().Entries}, nil
}

func appendLedgerMessages(
	ctx context.Context,
	outbox LedgerTxOutbox,
	reference string,
	decision modeling.Decision[Book, LedgerEvent],
) error {
	for event := range decision.Events() {
		message := IntegrationMessage{Type: event.EventName(), Reference: reference}
		if _, err := action.Run(ctx, outbox.Append(message)); err != nil {
			return err
		}
	}
	return nil
}

// Relay hands pending Banking messages to the Ledger workflow. It stands in
// for whatever actually moves messages in production — a broker, a poller, a
// job runner — and is deliberately a plain function: the point is that a
// workflow is not tied to an HTTP request, not that this demo ships a worker
// runtime. Delivery is at-least-once: the same message may arrive again, and
// the consumer's inbox decides what a duplicate means.
func Relay(
	pending []IntegrationMessage,
	record workflow.Workflow[TransferOccurred, LedgerReceipt],
) action.Action[[]LedgerReceipt] {
	return action.Do(func(ctx context.Context) ([]LedgerReceipt, error) {
		receipts := make([]LedgerReceipt, 0, len(pending))
		for i := range pending {
			receipt, err := action.Run(ctx, record(pending[i].Event))
			if err != nil {
				return receipts, err
			}
			receipts = append(receipts, receipt)
		}
		return receipts, nil
	})
}
