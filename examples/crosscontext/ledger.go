// Pure Ledger rules: the double-entry decision is total and effect-free. The
// context knows nothing about banking accounts, balances, or why money moved:
// it receives commands in its own vocabulary and applies its own rules.
package main

import (
	"errors"

	"go.alvadorn.org/functional/modeling"
)

// The Ledger failure vocabulary.
var (
	ErrBookNotFound     = errors.New("ledger: book not found")
	ErrUnbalancedEntry  = errors.New("ledger: entry must have a debit and a credit")
	ErrNonPositiveEntry = errors.New("ledger: amount must be positive")
)

// BookID identifies one ledger book.
type BookID string

// Amount is a bookkeeping amount in minor units.
type Amount int64

// NewBook seeds a book with no entries.
func NewBook(id BookID, version uint64) Book {
	return Book{ID: id, Version: version}
}

// Book is a small immutable value: posting an entry returns a new Book, and
// the previous one keeps its old totals forever.
type Book struct {
	ID      BookID
	Entries uint64
	Debits  Amount
	Credits Amount
	Version uint64
}

// RecordTransferCommand is this context's own command shape. It is not a
// banking event: the translation happens at the boundary, so a change to the
// Banking vocabulary never reaches these rules.
type RecordTransferCommand struct {
	Reference string
	DebitOf   string
	CreditOf  string
	Amount    Amount
}

// LedgerEvent is a fact this context emits.
type LedgerEvent interface {
	EventName() string
}

// EntryRecorded records one balanced double entry.
type EntryRecorded struct {
	Reference string
	DebitOf   string
	CreditOf  string
	Amount    Amount
}

// EventName implements LedgerEvent.
func (EntryRecorded) EventName() string { return "ledger.entry_recorded" }

// The recording rules: both sides present, amount positive. Requirements
// evaluate in declaration order and the first refusal wins.
var (
	BalancedEntry modeling.Requirement[Book, RecordTransferCommand] = func(_ Book, cmd RecordTransferCommand) error {
		if cmd.DebitOf == "" || cmd.CreditOf == "" {
			return ErrUnbalancedEntry
		}
		return nil
	}
	PositiveEntry modeling.Requirement[Book, RecordTransferCommand] = func(_ Book, cmd RecordTransferCommand) error {
		if cmd.Amount <= 0 {
			return ErrNonPositiveEntry
		}
		return nil
	}
)

var recordRequirements = modeling.AllRequirements(BalancedEntry, PositiveEntry)

// Post applies the entry to the book.
var Post modeling.Transition[Book, RecordTransferCommand, Book] = func(book Book, cmd RecordTransferCommand) Book {
	book.Entries++
	book.Debits += cmd.Amount
	book.Credits += cmd.Amount
	book.Version++
	return book
}

// Balanced is the invariant a ledger must always satisfy.
var Balanced modeling.Invariant[Book] = func(book Book) error {
	if book.Debits != book.Credits {
		return ErrUnbalancedEntry
	}
	return nil
}

// DecideRecord is the pure decision for posting a transfer to the book.
func DecideRecord(book Book, cmd RecordTransferCommand) (modeling.Decision[Book, LedgerEvent], error) {
	var refused modeling.Decision[Book, LedgerEvent]

	if err := recordRequirements(book, cmd); err != nil {
		return refused, err
	}

	next := Post.Apply(book, cmd)
	if err := Balanced(next); err != nil {
		return refused, err
	}

	return modeling.Decide[LedgerEvent](next).Emit(EntryRecorded(cmd)), nil
}
