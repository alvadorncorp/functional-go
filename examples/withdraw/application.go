// The application wires the pure withdraw rules to capabilities the
// composition root supplies. A capability is a small struct of functions, and
// the write capabilities exist only inside a transaction scope, so the
// compiler — not a convention — keeps writes inside the boundary.
package main

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// Accounts is what the application can do with accounts OUTSIDE a transaction:
// read one. The write capability deliberately does not exist here.
type Accounts struct {
	Load func(AccountID) action.Action[Account]
}

// TxAccounts is the account write capability, available only INSIDE a
// transaction scope.
type TxAccounts struct {
	Save func(Account) action.Action[struct{}]
}

// TxOutbox appends an outbox message inside the same transaction as the state
// change it accompanies.
type TxOutbox struct {
	Append func(OutboxMessage) action.Action[struct{}]
}

// TxScope is the set of capabilities bound to one concrete transaction. Every
// function it carries writes through the same handle, which is what makes the
// account row and its outbox rows commit or roll back together.
type TxScope struct {
	Accounts TxAccounts
	Outbox   TxOutbox
}

// Receipt is what a successful withdrawal returns to whichever transport asked
// for it.
type Receipt struct {
	AccountID AccountID
	Balance   Money
	Mode      WithdrawalMode
}

// MessageID identifies one outbox message.
type MessageID string

// OutboxMessage is an integration message awaiting delivery. It belongs to
// this application: the workflow module supplies the mechanisms to compose it
// into a transaction and takes no position on what an outbox is.
type OutboxMessage struct {
	ID          MessageID
	Type        string
	AggregateID string
	Payload     []byte
	CreatedAt   time.Time
}

// EventEncoder translates a domain event into an integration message. It is a
// capability rather than a fixed function so an application can choose JSON,
// Protobuf, or Avro, and can map a domain event to a differently shaped
// integration event, without any workflow changing.
type EventEncoder func(AccountEvent) (OutboxMessage, error)

// Clock and NewMessageID are ordinary function dependencies. Not every
// dependency is an effect: neither needs a context, so neither returns an
// Action.
type (
	Clock        func() time.Time
	NewMessageID func() MessageID
)

// NewJSONEncoder builds an EventEncoder that serializes the event as JSON and
// stamps it with the supplied clock and identifier generator.
func NewJSONEncoder(clock Clock, newID NewMessageID) EventEncoder {
	return func(event AccountEvent) (OutboxMessage, error) {
		payload, err := json.Marshal(event)
		if err != nil {
			return OutboxMessage{}, err
		}
		return OutboxMessage{
			ID:          newID(),
			Type:        event.EventName(),
			AggregateID: aggregateIDOf(event),
			Payload:     payload,
			CreatedAt:   clock(),
		}, nil
	}
}

func aggregateIDOf(event AccountEvent) string {
	switch typed := event.(type) {
	case MoneyWithdrawn:
		return string(typed.AccountID)
	case ApprovalRequested:
		return string(typed.AccountID)
	default:
		return ""
	}
}

// LoadAccount is the smallest reusable workflow in this application.
func LoadAccount(accounts Accounts) workflow.Workflow[AccountID, Account] {
	return workflow.From(accounts.Load)
}

// RequireAvailable is the pure rule the load workflow applies: an account that
// cannot take part in an operation fails here, before any transaction opens.
func RequireAvailable(account Account) (Account, error) {
	if account.Closed {
		return Account{}, ErrAccountClosed
	}
	return account, nil
}

// LoadAvailableAccount is LoadAccount followed by RequireAvailable. It is an
// ordinary Workflow, indistinguishable from a primitive one, and larger
// operations compose it without knowing it is composite.
func LoadAvailableAccount(accounts Accounts) workflow.Workflow[AccountID, Account] {
	return workflow.MapFallible(LoadAccount(accounts), RequireAvailable)
}

// NewWithdrawWorkflow builds the withdraw operation from its capabilities.
//
// Construction performs no I/O and captures no context. Applying the result to
// a command performs no I/O either: it returns an Action describing load,
// decide, and commit. Only action.Run starts any of it.
func NewWithdrawWorkflow(
	accounts Accounts,
	boundary transaction.Boundary[TxScope],
	encode EventEncoder,
) workflow.Workflow[WithdrawCommand, Receipt] {
	load := LoadAvailableAccount(accounts)

	return func(cmd WithdrawCommand) action.Action[Receipt] {
		return action.FlatMap(load(cmd.AccountID), func(account Account) action.Action[Receipt] {
			decision, err := Withdraw(account, cmd)
			if err != nil {
				// The domain refused. No transaction is ever opened.
				return action.Fail[Receipt](err)
			}
			return commitWithdrawal(boundary, encode, decision, SelectWithdrawalMode(cmd))
		})
	}
}

// commitWithdrawal persists the decision inside one transaction. The mode came
// from the domain policy; this function only executes the path the domain
// chose. Decision is passed by value deliberately: a pointer would invite
// mutation of a completed decision.
func commitWithdrawal(
	boundary transaction.Boundary[TxScope],
	encode EventEncoder,
	decision modeling.Decision[Account, AccountEvent],
	mode WithdrawalMode,
) action.Action[Receipt] {
	return transaction.Within(boundary, func(scope TxScope) action.Action[Receipt] {
		return action.Do(func(ctx context.Context) (Receipt, error) {
			switch mode {
			case ModeImmediate:
				if _, err := action.Run(ctx, scope.Accounts.Save(decision.State())); err != nil {
					return Receipt{}, err
				}
			case ModeRequiresApproval:
				// Nothing was debited, so there is no state change to persist:
				// the approval request travels as an outbox message alone.
			default:
				return Receipt{}, ErrUnknownWithdrawalMode
			}

			if _, err := action.Run(ctx, appendOutbox(scope.Outbox, encode, decision)); err != nil {
				return Receipt{}, err
			}

			state := decision.State()
			return Receipt{AccountID: state.ID, Balance: state.Balance, Mode: mode}, nil
		})
	})
}

// appendOutbox writes one message per emitted event, in emission order. The
// finite event sequence is materialized for Traverse, which snapshots it, so
// encoding and appending proceed in order and stop at the first failure. No
// events means no outbox rows.
func appendOutbox(
	outbox TxOutbox,
	encode EventEncoder,
	decision modeling.Decision[Account, AccountEvent],
) action.Action[struct{}] {
	events := slices.Collect(decision.Events())
	appended := action.Traverse(events, func(event AccountEvent) action.Action[struct{}] {
		message, err := encode(event)
		if err != nil {
			return action.Fail[struct{}](err)
		}
		return outbox.Append(message)
	})
	return action.Map(appended, func([]struct{}) struct{} { return struct{}{} })
}
