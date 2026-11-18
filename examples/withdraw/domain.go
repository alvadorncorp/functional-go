// Pure withdraw rules: requirements, a specification, a policy, a transition,
// and an invariant compose into one total decision function. Nothing here has
// an effect, so every rule is testable without a store, a workflow, or a
// transport.
package main

import (
	"errors"
	"fmt"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/modeling/policy"
	"go.alvadorn.org/functional/modeling/specification"
)

// The domain's failure vocabulary. The workflow propagates these unchanged;
// whatever transport exists is the only layer that decides what each one
// deserves.
var (
	ErrAccountNotFound       = errors.New("withdraw: account not found")
	ErrAccountClosed         = errors.New("withdraw: account is closed")
	ErrInsufficientBalance   = errors.New("withdraw: insufficient balance")
	ErrNonPositiveAmount     = errors.New("withdraw: amount must be positive")
	ErrNegativeBalance       = errors.New("withdraw: resulting balance would be negative")
	ErrUnknownWithdrawalMode = errors.New("withdraw: unknown withdrawal mode")
)

// AccountID identifies one account.
type AccountID string

// Money is an amount in minor units, so arithmetic stays exact.
type Money int64

// Account is a small immutable value. Nothing mutates an account in place:
// every update computes and returns a new value, and the previous one keeps its
// old balance forever.
type Account struct {
	ID      AccountID
	Balance Money
	Closed  bool
}

// WithdrawalMode is the path the domain selects for a withdrawal.
type WithdrawalMode int

const (
	// ModeImmediate applies the debit right away.
	ModeImmediate WithdrawalMode = iota
	// ModeRequiresApproval holds the withdrawal for manual approval.
	ModeRequiresApproval
)

// String names the mode for receipts and logs.
func (m WithdrawalMode) String() string {
	switch m {
	case ModeImmediate:
		return "immediate"
	case ModeRequiresApproval:
		return "requires_approval"
	default:
		return fmt.Sprintf("withdrawal-mode(%d)", int(m))
	}
}

// WithdrawCommand is the intent to take money out of an account.
type WithdrawCommand struct {
	AccountID AccountID
	Amount    Money
}

// AccountEvent is a fact this domain emits. It is a domain event, not an
// integration message: it carries no topic, no payload bytes, and no transport
// metadata. Translating it for the outside world is the application's job.
type AccountEvent interface {
	EventName() string
}

// MoneyWithdrawn records a completed debit.
type MoneyWithdrawn struct {
	AccountID AccountID
	Amount    Money
	Balance   Money
}

// EventName implements AccountEvent.
func (MoneyWithdrawn) EventName() string { return "account.money_withdrawn" }

// ApprovalRequested records that a withdrawal was held for manual approval
// instead of being applied.
type ApprovalRequested struct {
	AccountID AccountID
	Amount    Money
}

// EventName implements AccountEvent.
func (ApprovalRequested) EventName() string { return "account.approval_requested" }

// PositiveAmount rejects a withdrawal of zero or less.
var PositiveAmount modeling.Requirement[Account, WithdrawCommand] = func(_ Account, cmd WithdrawCommand) error {
	if cmd.Amount <= 0 {
		return ErrNonPositiveAmount
	}
	return nil
}

// HasEnoughBalance rejects a withdrawal larger than the current balance.
var HasEnoughBalance modeling.Requirement[Account, WithdrawCommand] = func(account Account, cmd WithdrawCommand) error {
	if account.Balance < cmd.Amount {
		return ErrInsufficientBalance
	}
	return nil
}

// AccountIsOpen rejects any withdrawal from a closed account.
var AccountIsOpen modeling.Requirement[Account, WithdrawCommand] = func(account Account, _ WithdrawCommand) error {
	if account.Closed {
		return ErrAccountClosed
	}
	return nil
}

// withdrawRequirements composes the rules every withdrawal must satisfy before
// the domain asks which path to take.
var withdrawRequirements = modeling.AllRequirements(AccountIsOpen, PositiveAmount, HasEnoughBalance)

// approvalThreshold is a business rule and therefore lives here. A workflow
// that wrote `if cmd.Amount > approvalThreshold` would be hiding a policy
// inside orchestration.
const approvalThreshold Money = 1_000_00

// RequiresManualReview classifies withdrawals whose amount exceeds the
// automatic-processing threshold.
var RequiresManualReview specification.Specification[WithdrawCommand] = func(cmd WithdrawCommand) bool {
	return cmd.Amount > approvalThreshold
}

// SelectWithdrawalMode decides which path a command takes. The explicit
// closure keeps specification classification and policy selection distinct.
var SelectWithdrawalMode = policy.Choose(
	func(cmd WithdrawCommand) bool { return RequiresManualReview(cmd) },
	policy.Constant[WithdrawCommand](ModeRequiresApproval),
	policy.Constant[WithdrawCommand](ModeImmediate),
)

// DebitAccount computes the next account state. The input account keeps its old
// balance; the returned value carries the debit.
var DebitAccount modeling.Transition[Account, WithdrawCommand, Account] = func(
	account Account, cmd WithdrawCommand,
) Account {
	return Account{ID: account.ID, Balance: account.Balance - cmd.Amount, Closed: account.Closed}
}

// NonNegativeBalance is the invariant every resulting account must satisfy.
var NonNegativeBalance modeling.Invariant[Account] = func(account Account) error {
	if account.Balance < 0 {
		return ErrNegativeBalance
	}
	return nil
}

// Withdraw decides what happens when cmd is applied to account.
//
// It is a total pure function: same inputs, same outputs, no effects. It
// checks the requirements, consults the policy, applies the transition on the
// immediate path, validates the resulting invariant, and returns the new state
// together with the events that state change implies. It never persists
// anything — the returned Decision is a value the caller is free to commit,
// log, or discard.
func Withdraw(account Account, cmd WithdrawCommand) (modeling.Decision[Account, AccountEvent], error) {
	var refused modeling.Decision[Account, AccountEvent]

	if err := withdrawRequirements(account, cmd); err != nil {
		return refused, err
	}

	switch mode := SelectWithdrawalMode(cmd); mode {
	case ModeImmediate:
		next := DebitAccount(account, cmd)
		if err := NonNegativeBalance(next); err != nil {
			return refused, err
		}
		return modeling.Decide[AccountEvent](next).Emit(MoneyWithdrawn{
			AccountID: next.ID,
			Amount:    cmd.Amount,
			Balance:   next.Balance,
		}), nil

	case ModeRequiresApproval:
		// The account is unchanged: nothing is debited until a human approves.
		return modeling.Decide[AccountEvent](account).Emit(ApprovalRequested{
			AccountID: account.ID,
			Amount:    cmd.Amount,
		}), nil

	default:
		return refused, ErrUnknownWithdrawalMode
	}
}
