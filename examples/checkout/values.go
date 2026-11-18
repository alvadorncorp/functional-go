package main

import (
	"errors"

	"go.alvadorn.org/functional/values/result"
)

// USD keeps every amount in the demo in one currency, so currency-mismatch
// math stays a pure-function concern rather than a scenario concern.
const USD Currency = "USD"

var (
	ErrNegativeMoney     = errors.New("money: negative amount")
	ErrCurrencyMismatch  = errors.New("money: currency mismatch")
	ErrInsufficientMoney = errors.New("money: insufficient amount")
	ErrInvalidQuantity   = errors.New("quantity: must be positive")
)

// Currency names the settlement currency of a Money value.
type Currency string

// Money is a value object: cents in exactly one currency, never mutated in
// place. Every arithmetic result is a new Money guarded by a Result.
type Money struct {
	cents    int64
	currency Currency
}

// NewMoney rejects negative amounts.
func NewMoney(cents int64, currency Currency) result.Result[Money, error] {
	if cents < 0 {
		return result.Err[Money](ErrNegativeMoney)
	}
	return result.Ok[Money, error](Money{cents: cents, currency: currency})
}

// ZeroMoney returns the neutral amount of a currency.
func ZeroMoney(currency Currency) Money {
	return Money{currency: currency}
}

// Cents returns the raw minor-unit amount.
func (m Money) Cents() int64 { return m.cents }

// Currency returns the settlement currency.
func (m Money) Currency() Currency { return m.currency }

// IsZero reports whether the amount carries no value.
func (m Money) IsZero() bool { return m.cents == 0 }

// percentScale converts a whole-number percent into a fraction.
const percentScale = 100

// Percentage applies a whole-number percent and truncates the remainder.
func (m Money) Percentage(percent uint8) Money {
	return Money{cents: m.cents * int64(percent) / percentScale, currency: m.currency}
}

// Add sums two amounts of the same currency.
func (m Money) Add(other Money) result.Result[Money, error] {
	if m.currency != other.currency {
		return result.Err[Money](ErrCurrencyMismatch)
	}
	return NewMoney(m.cents+other.cents, m.currency)
}

// Sub subtracts same-currency amounts and rejects overdrafts.
func (m Money) Sub(other Money) result.Result[Money, error] {
	if m.currency != other.currency {
		return result.Err[Money](ErrCurrencyMismatch)
	}
	if other.cents > m.cents {
		return result.Err[Money](ErrInsufficientMoney)
	}
	return NewMoney(m.cents-other.cents, m.currency)
}

// Typed identifiers keep entity references explicit at every boundary.
type (
	OrderID         string
	CustomerID      string
	ProductID       string
	PaymentMethodID string
	CouponCode      string
	PromotionID     string
)

// Quantity is a positive line-item count.
type Quantity uint16

// NewQuantity rejects zero.
func NewQuantity(value uint16) result.Result[Quantity, error] {
	if value == 0 {
		return result.Err[Quantity](ErrInvalidQuantity)
	}
	return result.Ok[Quantity, error](Quantity(value))
}

// Value unwraps the count.
func (q Quantity) Value() uint16 { return uint16(q) }
