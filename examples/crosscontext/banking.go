// Pure Banking rules: requirements, a transition, and an invariant compose
// into one total decision function. Nothing here has an effect, so every rule
// is testable without a store, a workflow, or a transport. Banking knows
// nothing about ledgers, books, debits, or credits, and none of its
// vocabulary says how its facts reach any other context.
package main

import (
	"errors"

	"go.alvadorn.org/functional/modeling"
)

// The Banking failure vocabulary. The workflow propagates these unchanged;
// whatever transport exists is the only layer that decides what each one
// deserves.
var (
	ErrAccountNotFound   = errors.New("banking: account not found")
	ErrAccountClosed     = errors.New("banking: account is closed")
	ErrInsufficientFunds = errors.New("banking: insufficient funds")
	ErrNonPositiveAmount = errors.New("banking: amount must be positive")
	ErrSameAccount       = errors.New("banking: source and destination are the same")
)

// AccountID identifies one banking account.
type AccountID string

// Money is an amount in minor units, so arithmetic stays exact.
type Money int64

// NewAccount seeds an open account.
func NewAccount(id AccountID, balance Money, version uint64) Account {
	return Account{ID: id, Balance: balance, Version: version}
}

// Account is a small immutable value. Nothing mutates an account in place:
// every update computes and returns a new value, and the previous one keeps
// its old balance forever.
type Account struct {
	ID      AccountID
	Balance Money
	Version uint64
	Closed  bool
}

// TransferCommand is the intent to move money out of one account. Reference
// is the caller-supplied idempotency key that follows the operation across
// every context it touches.
type TransferCommand struct {
	Reference string
	From      AccountID
	To        AccountID
	Amount    Money
}

// BankingEvent is a fact this context emits. It is a domain event, not an
// integration message: it carries no topic, no payload bytes, and no
// transport metadata. Translating it for the outside world is somebody else's
// job.
type BankingEvent interface {
	EventName() string
}

// MoneyDebited is an internal fact: money left the source account.
type MoneyDebited struct {
	AccountID AccountID
	Amount    Money
	Balance   Money
}

// EventName implements BankingEvent.
func (MoneyDebited) EventName() string { return "banking.money_debited" }

// TransferOccurred is the integration fact other contexts care about. It is
// deliberately a different type from MoneyDebited: an internal event and an
// integration event serve different audiences and need not be one to one.
type TransferOccurred struct {
	Reference string
	From      AccountID
	To        AccountID
	Amount    Money
}

// EventName implements BankingEvent.
func (TransferOccurred) EventName() string { return "banking.transfer_occurred" }

// The transfer rules. Requirements evaluate in declaration order and the
// first refusal wins.
var (
	PositiveAmount modeling.Requirement[Account, TransferCommand] = func(_ Account, cmd TransferCommand) error {
		if cmd.Amount <= 0 {
			return ErrNonPositiveAmount
		}
		return nil
	}
	DistinctAccounts modeling.Requirement[Account, TransferCommand] = func(_ Account, cmd TransferCommand) error {
		if cmd.From == cmd.To {
			return ErrSameAccount
		}
		return nil
	}
	SourceIsOpen modeling.Requirement[Account, TransferCommand] = func(account Account, _ TransferCommand) error {
		if account.Closed {
			return ErrAccountClosed
		}
		return nil
	}
	SufficientFunds modeling.Requirement[Account, TransferCommand] = func(account Account, cmd TransferCommand) error {
		if account.Balance < cmd.Amount {
			return ErrInsufficientFunds
		}
		return nil
	}
)

var transferRequirements = modeling.AllRequirements(
	PositiveAmount, DistinctAccounts, SourceIsOpen, SufficientFunds,
)

// Debit removes the amount from the source account.
var Debit modeling.Transition[Account, TransferCommand, Account] = func(account Account, cmd TransferCommand) Account {
	account.Balance -= cmd.Amount
	account.Version++
	return account
}

// SolventAccount is the invariant the resulting account must satisfy.
var SolventAccount modeling.Invariant[Account] = func(account Account) error {
	if account.Balance < 0 {
		return ErrInsufficientFunds
	}
	return nil
}

// DecideTransfer is the pure decision for a transfer out of the source
// account. It emits the internal debit first and the integration fact
// second; anything persisting these must keep that order.
func DecideTransfer(source Account, cmd TransferCommand) (modeling.Decision[Account, BankingEvent], error) {
	var refused modeling.Decision[Account, BankingEvent]

	if err := transferRequirements(source, cmd); err != nil {
		return refused, err
	}

	next := Debit.Apply(source, cmd)
	if err := SolventAccount(next); err != nil {
		return refused, err
	}

	return modeling.Decide[BankingEvent](next).Emit(
		MoneyDebited{AccountID: next.ID, Amount: cmd.Amount, Balance: next.Balance},
		TransferOccurred(cmd),
	), nil
}
