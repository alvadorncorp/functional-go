// The shared topology: ONE resource manager serves both contexts, so the
// Banking state, the Ledger state, and the integration outbox land in a
// single ACID transaction. Two bounded contexts sharing a database is a
// deliberate deployment choice, not a modelling compromise: a bounded context
// is a semantic boundary, a transaction is a physical one, and nothing forces
// them to coincide. The contexts keep separate models, separate rules, and
// separate sections of this package; only the capabilities they are handed
// happen to write through the same handle.
package main

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// DomainEvent is the small contract the outbox writer needs from either
// context's facts: a stable wire name. Both contexts satisfy it structurally.
type DomainEvent interface {
	EventName() string
}

// OutboxMessage is an integration message awaiting delivery. It belongs to
// this application: the workflow module supplies the mechanisms to compose it
// into a transaction and takes no position on what an outbox is.
type OutboxMessage struct {
	Type      string
	Reference string
	Payload   []byte
}

// EventEncoder translates a domain event into an integration message. It is a
// capability rather than a fixed function so an application can choose JSON,
// Protobuf, or Avro, without any workflow changing.
type EventEncoder func(reference string, event DomainEvent) (OutboxMessage, error)

// NewJSONEncoder builds an EventEncoder that serializes the event as JSON.
func NewJSONEncoder() EventEncoder {
	return func(reference string, event DomainEvent) (OutboxMessage, error) {
		payload, err := json.Marshal(event)
		if err != nil {
			return OutboxMessage{}, err
		}
		return OutboxMessage{Type: event.EventName(), Reference: reference, Payload: payload}, nil
	}
}

// Accounts and Books are read capabilities, available outside any transaction.
// The write capabilities deliberately do not exist here.
type (
	Accounts struct {
		Load func(AccountID) action.Action[Account]
	}
	Books struct {
		Load func(BookID) action.Action[Book]
	}
)

// TxAccounts, TxBooks, and TxOutbox are write capabilities, available only
// inside a transaction scope.
type (
	TxAccounts struct {
		Save func(Account) action.Action[struct{}]
	}
	TxBooks struct {
		Save func(Book) action.Action[struct{}]
	}
	TxOutbox struct {
		Append func(OutboxMessage) action.Action[struct{}]
	}
)

// TransferTxScope is the set of capabilities bound to one concrete
// transaction. Every function it carries writes through the same handle,
// which is what makes the Banking write, the Ledger write, and the outbox
// rows commit or roll back together. This type is what turns "all three land
// or none does" from a hope into a property of the code.
type TransferTxScope struct {
	Accounts TxAccounts
	Books    TxBooks
	Outbox   TxOutbox
}

// Receipt is what a successful shared transfer returns.
type Receipt struct {
	SourceBalance Money
	BookEntries   uint64
}

// Store is one in-memory resource manager serving both contexts, mirroring
// the deployment where Banking and Ledger share a database. It records the
// transaction lifecycle so the demo and the tests can assert exactly which
// operations reached the resource manager.
type Store struct {
	mu       sync.Mutex
	accounts map[AccountID]Account
	books    map[BookID]Book
	outbox   []OutboxMessage
	trace    []string

	bankingSaveErr error
	bookSaveErr    error
}

// NewStore seeds the store with committed state.
func NewStore(account Account, book Book) *Store {
	return &Store{
		accounts: map[AccountID]Account{account.ID: account},
		books:    map[BookID]Book{book.ID: book},
	}
}

// FailNextBankingSave makes exactly the next Banking save fail, then clears
// itself, the way a full disk would; the next transaction reaches a healthy
// store again.
func (s *Store) FailNextBankingSave(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bankingSaveErr = err
}

// FailNextBookSave makes exactly the next Ledger save fail, then clears
// itself.
func (s *Store) FailNextBookSave(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bookSaveErr = err
}

// Trace is the comma-joined lifecycle observed so far.
func (s *Store) Trace() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.trace, ",")
}

// Snapshot returns the committed state of both contexts and the outbox.
func (s *Store) Snapshot(id AccountID, bookID BookID) (Account, Book, []OutboxMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accounts[id], s.books[bookID], append([]OutboxMessage(nil), s.outbox...)
}

// Accounts is the Banking read capability.
func (s *Store) Accounts() Accounts {
	return Accounts{
		Load: func(id AccountID) action.Action[Account] {
			return action.Do(func(ctx context.Context) (Account, error) {
				if err := ctx.Err(); err != nil {
					return Account{}, err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				s.trace = append(s.trace, "load-account")
				account, found := s.accounts[id]
				if !found {
					return Account{}, ErrAccountNotFound
				}
				return account, nil
			})
		},
	}
}

// Books is the Ledger read capability.
func (s *Store) Books() Books {
	return Books{
		Load: func(id BookID) action.Action[Book] {
			return action.Do(func(ctx context.Context) (Book, error) {
				if err := ctx.Err(); err != nil {
					return Book{}, err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				s.trace = append(s.trace, "load-book")
				book, found := s.books[id]
				if !found {
					return Book{}, ErrBookNotFound
				}
				return book, nil
			})
		},
	}
}

// pendingTransfer holds uncommitted writes for one Boundary invocation.
type pendingTransfer struct {
	account Account
	saved   bool
	book    Book
	posted  bool
	outbox  []OutboxMessage
}

// Boundary is the one transaction both contexts write through. Writes
// accumulate in a pending buffer and reach the committed state only on
// commit, which is what makes a rollback observable as "nothing changed
// anywhere".
func (s *Store) Boundary() transaction.Boundary[TransferTxScope] {
	return func(ctx context.Context, fn func(context.Context, TransferTxScope) error) error {
		s.appendTrace("begin")

		pending := &pendingTransfer{}
		committed := false
		defer func() {
			if !committed {
				s.appendTrace("rollback")
			}
		}()

		if err := fn(ctx, s.transferScope(pending)); err != nil {
			return err
		}

		s.publish(pending)
		committed = true
		return nil
	}
}

func (s *Store) transferScope(pending *pendingTransfer) TransferTxScope {
	return TransferTxScope{
		Accounts: TxAccounts{
			Save: func(account Account) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					if err := s.takeBankingSaveError(); err != nil {
						s.appendTrace("banking-save-failed")
						return err
					}
					s.appendTrace("banking-save")
					pending.account, pending.saved = account, true
					return nil
				})
			},
		},
		Books: TxBooks{
			Save: func(book Book) action.Action[struct{}] { return s.stageBook(pending, book) },
		},
		Outbox: TxOutbox{
			Append: func(message OutboxMessage) action.Action[struct{}] {
				return action.Exec(func(ctx context.Context) error {
					if err := ctx.Err(); err != nil {
						return err
					}
					s.appendTrace("append")
					pending.outbox = append(pending.outbox, message)
					return nil
				})
			},
		},
	}
}

func (s *Store) stageBook(pending *pendingTransfer, book Book) action.Action[struct{}] {
	return action.Exec(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.takeBookSaveError(); err != nil {
			s.appendTrace("book-save-failed")
			return err
		}
		s.appendTrace("book-save")
		pending.book, pending.posted = book, true
		return nil
	})
}

func (s *Store) takeBankingSaveError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.bankingSaveErr
	s.bankingSaveErr = nil
	return err
}

func (s *Store) takeBookSaveError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.bookSaveErr
	s.bookSaveErr = nil
	return err
}

func (s *Store) publish(pending *pendingTransfer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pending.saved {
		s.accounts[pending.account.ID] = pending.account
	}
	if pending.posted {
		s.books[pending.book.ID] = pending.book
	}
	s.outbox = append(s.outbox, pending.outbox...)
	s.trace = append(s.trace, "commit")
}

func (s *Store) appendTrace(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trace = append(s.trace, event)
}

// LoadAccount is the smallest reusable workflow in this application.
func LoadAccount(accounts Accounts) workflow.Workflow[AccountID, Account] {
	return workflow.From(accounts.Load)
}

// RequireOpenAccount is the pure rule the load workflow applies: an account
// that cannot take part in an operation fails here, before any transaction
// opens.
func RequireOpenAccount(account Account) (Account, error) {
	if account.Closed {
		return Account{}, ErrAccountClosed
	}
	return account, nil
}

// LoadAvailableAccount is LoadAccount followed by RequireOpenAccount. A
// deposit workflow would consume it exactly the same way, which is the point:
// a composed workflow becomes a reusable word in the application's
// vocabulary.
func LoadAvailableAccount(accounts Accounts) workflow.Workflow[AccountID, Account] {
	return workflow.MapFallible(LoadAccount(accounts), RequireOpenAccount)
}

// NewTransferWorkflow coordinates both contexts.
//
// Each context decides for itself, purely, before anything is written. Only
// then does one transaction open, and inside it the Banking state, the Ledger
// state, and the integration outbox are written through capabilities bound to
// the same handle.
func NewTransferWorkflow(
	accounts Accounts,
	books Books,
	bookID BookID,
	boundary transaction.Boundary[TransferTxScope],
	encode EventEncoder,
) workflow.Workflow[TransferCommand, Receipt] {
	load := LoadAvailableAccount(accounts)

	return func(cmd TransferCommand) action.Action[Receipt] {
		return action.FlatMap(load(cmd.From), func(source Account) action.Action[Receipt] {
			bankingDecision, err := DecideTransfer(source, cmd)
			if err != nil {
				// The Banking domain refused. No transaction is ever opened.
				return action.Fail[Receipt](err)
			}
			occurred, found := integrationEvent(bankingDecision)
			if !found {
				// DecideTransfer always emits the integration fact; this
				// guard keeps the workflow honest if the domain ever stops.
				return action.Fail[Receipt](ErrAccountNotFound)
			}
			return action.FlatMap(books.Load(bookID), func(book Book) action.Action[Receipt] {
				bookDecision, err := DecideRecord(book, ToLedgerCommand(occurred))
				if err != nil {
					// The Ledger domain refused. Still no transaction.
					return action.Fail[Receipt](err)
				}
				return commitTransfer(boundary, encode, cmd.Reference, bankingDecision, bookDecision)
			})
		})
	}
}

// integrationEvent picks the fact other contexts subscribe to. The workflow
// reads what the domain emitted rather than reconstructing it. The Decision
// is passed by value deliberately: a pointer would invite mutation of a
// completed decision.
func integrationEvent(decision modeling.Decision[Account, BankingEvent]) (TransferOccurred, bool) {
	for event := range decision.Events() {
		if occurred, ok := event.(TransferOccurred); ok {
			return occurred, true
		}
	}
	return TransferOccurred{}, false
}

// commitTransfer persists both decisions inside one transaction. Decision is
// passed by value deliberately: a pointer would invite mutation of a
// completed decision.
func commitTransfer(
	boundary transaction.Boundary[TransferTxScope],
	encode EventEncoder,
	reference string,
	bankingDecision modeling.Decision[Account, BankingEvent],
	bookDecision modeling.Decision[Book, LedgerEvent],
) action.Action[Receipt] {
	return transaction.Within(boundary, func(scope TransferTxScope) action.Action[Receipt] {
		return action.Do(func(ctx context.Context) (Receipt, error) {
			if _, err := action.Run(ctx, scope.Accounts.Save(bankingDecision.State())); err != nil {
				return Receipt{}, err
			}
			if _, err := action.Run(ctx, scope.Books.Save(bookDecision.State())); err != nil {
				return Receipt{}, err
			}
			// Banking's facts travel first, then Ledger's, each decision in
			// its own emission order — one transaction carries them all.
			if _, err := action.Run(ctx, appendOutbox(scope.Outbox, encode, reference, bankingDecision)); err != nil {
				return Receipt{}, err
			}
			if _, err := action.Run(ctx, appendOutbox(scope.Outbox, encode, reference, bookDecision)); err != nil {
				return Receipt{}, err
			}
			state := bankingDecision.State()
			return Receipt{SourceBalance: state.Balance, BookEntries: bookDecision.State().Entries}, nil
		})
	})
}

// appendOutbox writes one message per emitted event of one decision, in
// emission order. The finite event sequence is materialized for Traverse,
// which snapshots it, so encoding and appending proceed in order and stop at
// the first failure. No events means no outbox rows.
func appendOutbox[S any, E DomainEvent](
	outbox TxOutbox,
	encode EventEncoder,
	reference string,
	decision modeling.Decision[S, E],
) action.Action[struct{}] {
	events := slices.Collect(decision.Events())
	appended := action.Traverse(events, func(event E) action.Action[struct{}] {
		message, err := encode(reference, event)
		if err != nil {
			return action.Fail[struct{}](err)
		}
		return outbox.Append(message)
	})
	return action.Map(appended, func([]struct{}) struct{} { return struct{}{} })
}
