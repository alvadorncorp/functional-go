// The in-memory store plays the resource manager a SQL adapter would play in
// production: writes accumulate in a pending buffer and reach the committed
// state only on commit, so a rollback is observable as "nothing changed".
package main

import (
	"context"
	"strings"
	"sync"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// noAppendFailure disables the injected append failure.
const noAppendFailure = -1

// Store holds committed state and records the transaction lifecycle so the
// demo and the tests can assert exactly which operations reached the resource
// manager.
type Store struct {
	mu       sync.Mutex
	accounts map[AccountID]Account
	outbox   []OutboxMessage
	trace    []string

	appendErr   error
	appendFails int
}

// NewStore seeds the store with committed accounts.
func NewStore(accounts ...Account) *Store {
	seeded := make(map[AccountID]Account, len(accounts))
	for i := range accounts {
		seeded[accounts[i].ID] = accounts[i]
	}
	return &Store{accounts: seeded, appendFails: noAppendFailure}
}

// FailAppendAt makes the append at the given zero-based index fail, the way a
// full disk or a lost connection would in production.
func (s *Store) FailAppendAt(index int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.appendFails, s.appendErr = index, err
}

// Trace is the comma-joined lifecycle observed so far.
func (s *Store) Trace() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.trace, ",")
}

// Snapshot returns the committed account for id and the committed outbox.
func (s *Store) Snapshot(id AccountID) (Account, []OutboxMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accounts[id], append([]OutboxMessage(nil), s.outbox...)
}

// Accounts is the read capability, available outside any transaction.
func (s *Store) Accounts() Accounts {
	return Accounts{
		Load: func(id AccountID) action.Action[Account] {
			return action.Do(func(ctx context.Context) (Account, error) {
				if err := ctx.Err(); err != nil {
					return Account{}, err
				}
				s.mu.Lock()
				defer s.mu.Unlock()
				s.trace = append(s.trace, "load")
				account, found := s.accounts[id]
				if !found {
					return Account{}, ErrAccountNotFound
				}
				return account, nil
			})
		},
	}
}

// Boundary is the transactional region. Writes accumulate in a pending buffer
// and are published to the committed state only on commit, which is what makes
// a rollback observable as "nothing changed".
func (s *Store) Boundary() transaction.Boundary[TxScope] {
	return func(ctx context.Context, fn func(context.Context, TxScope) error) error {
		s.appendTrace("begin")

		pending := &pendingTxn{accounts: map[AccountID]Account{}}

		committed := false
		defer func() {
			if !committed {
				s.appendTrace("rollback")
			}
		}()

		scope := TxScope{
			Accounts: TxAccounts{Save: s.pendingSave(pending)},
			Outbox:   TxOutbox{Append: s.pendingAppend(pending)},
		}

		if err := fn(ctx, scope); err != nil {
			return err
		}

		s.publish(pending.accounts, pending.outbox)
		committed = true
		return nil
	}
}

// pendingTxn holds uncommitted writes for one Boundary invocation.
type pendingTxn struct {
	accounts map[AccountID]Account
	outbox   []OutboxMessage
	appends  int
}

// pendingSave stages an account write into the pending buffer.
func (s *Store) pendingSave(pending *pendingTxn) func(Account) action.Action[struct{}] {
	return func(account Account) action.Action[struct{}] {
		return action.Exec(func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			s.appendTrace("save")
			pending.accounts[account.ID] = account
			return nil
		})
	}
}

// pendingAppend stages an outbox message into the pending buffer, or fails at
// the injected append index.
func (s *Store) pendingAppend(pending *pendingTxn) func(OutboxMessage) action.Action[struct{}] {
	return func(message OutboxMessage) action.Action[struct{}] {
		return action.Exec(func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			index := pending.appends
			pending.appends++
			if err := s.appendError(index); err != nil {
				s.appendTrace("append-failed")
				return err
			}
			s.appendTrace("append")
			pending.outbox = append(pending.outbox, message)
			return nil
		})
	}
}

func (s *Store) publish(accounts map[AccountID]Account, outbox []OutboxMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, account := range accounts {
		s.accounts[id] = account
	}
	s.outbox = append(s.outbox, outbox...)
	s.trace = append(s.trace, "commit")
}

func (s *Store) appendTrace(event string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trace = append(s.trace, event)
}

func (s *Store) appendError(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.appendFails == index {
		return s.appendErr
	}
	return nil
}
