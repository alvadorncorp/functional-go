package main

import (
	"errors"

	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/modeling/change"
	"go.alvadorn.org/functional/structural/iterator"
	"go.alvadorn.org/functional/values/result"
)

var (
	ErrOrderNotEditable      = errors.New("order: only draft orders are editable")
	ErrEmptyOrder            = errors.New("order: no line items")
	ErrAlreadyCheckedOut     = errors.New("order: already checked out")
	ErrOrderOwnership        = errors.New("order: belongs to another customer")
	ErrEmptyCheckedOutOrder  = errors.New("order: checked out with no line items")
	ErrUnbalancedOrderTotals = errors.New("order: subtotal does not equal discount plus total")
)

// OrderStatus is the checkout lifecycle position of an order.
type OrderStatus uint8

const (
	// OrderStatusDraft accepts line items.
	OrderStatusDraft OrderStatus = iota
	// OrderStatusCheckedOut is final for this demo.
	OrderStatusCheckedOut
)

// String renders stable, lowercase status names for output.
func (s OrderStatus) String() string {
	if s == OrderStatusCheckedOut {
		return "checked_out"
	}
	return "draft"
}

// OrderVersion drives optimistic concurrency on saves.
type OrderVersion uint64

// initialOrderVersion is the version of every freshly created draft.
const initialOrderVersion OrderVersion = 1

// OrderItem prices one product at a unit price and count.
type OrderItem struct {
	ProductID ProductID
	UnitPrice Money
	Quantity  Quantity
}

// itemTotal multiplies unit price by count.
func itemTotal(item OrderItem) result.Result[Money, error] {
	return NewMoney(item.UnitPrice.Cents()*int64(item.Quantity.Value()), item.UnitPrice.Currency())
}

// calculateSubtotal folds every line item into one amount, threading the
// Result so the first mismatched currency or negative amount stops the fold.
func calculateSubtotal(items immutable.List[OrderItem], currency Currency) result.Result[Money, error] {
	return iterator.Fold(items.Values(), result.Ok[Money, error](ZeroMoney(currency)),
		func(total result.Result[Money, error], item OrderItem) result.Result[Money, error] {
			return result.FlatMap(total, func(running Money) result.Result[Money, error] {
				return result.FlatMap(itemTotal(item), running.Add)
			})
		})
}

// Order is a small copy-on-write handle. Updates clone the payload and return
// a new handle, so published state is never mutated in place. The zero Order
// is inert: getters return zero values and transforms return the receiver.
type Order struct {
	p *orderPayload
}

type orderPayload struct {
	id         OrderID
	customerID CustomerID
	status     OrderStatus
	items      immutable.List[OrderItem]
	promotions immutable.Set[PromotionID]
	subtotal   Money
	discount   Money
	total      Money
	version    OrderVersion
}

// NewOrder creates an empty draft at the initial version.
func NewOrder(id OrderID, customer CustomerID, currency Currency) Order {
	zero := ZeroMoney(currency)
	return Order{p: &orderPayload{
		id:         id,
		customerID: customer,
		status:     OrderStatusDraft,
		subtotal:   zero,
		discount:   zero,
		total:      zero,
		version:    initialOrderVersion,
	}}
}

// edit clones the payload, applies one mutation, and returns the new handle.
func (o Order) edit(mutate func(*orderPayload)) Order {
	if o.p == nil {
		return o
	}
	next := *o.p
	mutate(&next)
	return Order{p: &next}
}

// ID returns the order identity.
func (o Order) ID() OrderID {
	if o.p == nil {
		return ""
	}
	return o.p.id
}

// CustomerID returns the owning customer.
func (o Order) CustomerID() CustomerID {
	if o.p == nil {
		return ""
	}
	return o.p.customerID
}

// Status returns the lifecycle position.
func (o Order) Status() OrderStatus {
	if o.p == nil {
		return OrderStatusDraft
	}
	return o.p.status
}

// ItemsCount returns the number of line items.
func (o Order) ItemsCount() int {
	if o.p == nil {
		return 0
	}
	return o.p.items.Len()
}

// Subtotal returns the folded item total.
func (o Order) Subtotal() Money {
	if o.p == nil {
		return ZeroMoney(USD)
	}
	return o.p.subtotal
}

// Discount returns the applied discount amount.
func (o Order) Discount() Money {
	if o.p == nil {
		return ZeroMoney(USD)
	}
	return o.p.discount
}

// Total returns the payable amount.
func (o Order) Total() Money {
	if o.p == nil {
		return ZeroMoney(USD)
	}
	return o.p.total
}

// Version returns the optimistic-concurrency version.
func (o Order) Version() OrderVersion {
	if o.p == nil {
		return initialOrderVersion
	}
	return o.p.version
}

// PromotionsCount returns how many promotions the order carries.
func (o Order) PromotionsCount() int {
	if o.p == nil {
		return 0
	}
	return o.p.promotions.Len()
}

// AddItem appends a line item and recalculates subtotal and total. The
// receiver is never mutated.
func (o Order) AddItem(item OrderItem) result.Result[Order, error] {
	if o.p == nil || o.p.status != OrderStatusDraft {
		return result.Err[Order](ErrOrderNotEditable)
	}
	items := o.p.items.Append(item)
	subtotal, err := result.ToPair(calculateSubtotal(items, o.p.subtotal.Currency()))
	if err != nil {
		return result.Err[Order](err)
	}
	return result.Ok[Order, error](o.edit(func(payload *orderPayload) {
		payload.items = items
		payload.subtotal = subtotal
		payload.total = subtotal
	}))
}

// MarkCheckedOut closes the draft.
func (o Order) MarkCheckedOut() Order {
	return o.edit(func(payload *orderPayload) { payload.status = OrderStatusCheckedOut })
}

// WithPromotion records a promotion id.
func (o Order) WithPromotion(id PromotionID) Order {
	return o.edit(func(payload *orderPayload) { payload.promotions = payload.promotions.Add(id) })
}

// WithDiscount stores the applied discount amount.
func (o Order) WithDiscount(amount Money) Order {
	return o.edit(func(payload *orderPayload) { payload.discount = amount })
}

// WithTotal stores the payable amount.
func (o Order) WithTotal(amount Money) Order {
	return o.edit(func(payload *orderPayload) { payload.total = amount })
}

// BumpVersion advances optimistic concurrency by one.
func (o Order) BumpVersion() Order {
	return o.edit(func(payload *orderPayload) { payload.version++ })
}

// ApplyCheckedOutState applies checkout totals during event replay. Used by
// ReduceOrder; the receiver is never mutated.
func (o Order) ApplyCheckedOutState(subtotal, discount, total Money, version OrderVersion) Order {
	return o.edit(func(payload *orderPayload) {
		payload.status = OrderStatusCheckedOut
		payload.subtotal = subtotal
		payload.discount = discount
		payload.total = total
		payload.version = version
	})
}

// SameState reports whether two handles carry equivalent domain state.
// Handles are pointers, so equality compares payload fields, not identity.
func (o Order) SameState(other Order) bool {
	if o.p == nil || other.p == nil {
		return o.p == other.p
	}
	return o.p.id == other.p.id &&
		o.p.customerID == other.p.customerID &&
		o.p.status == other.p.status &&
		o.p.items.Equal(other.p.items) &&
		o.p.promotions.Equal(other.p.promotions) &&
		o.p.subtotal == other.p.subtotal &&
		o.p.discount == other.p.discount &&
		o.p.total == other.p.total &&
		o.p.version == other.p.version
}

// markCheckedOutChange exposes MarkCheckedOut as a named state change.
func markCheckedOutChange() change.Change[Order] {
	return Order.MarkCheckedOut
}

// discountChange returns a change that stores the discount amount.
func discountChange(amount Money) change.Change[Order] {
	return func(o Order) Order { return o.WithDiscount(amount) }
}

// totalChange returns a change that stores the payable amount.
func totalChange(amount Money) change.Change[Order] {
	return func(o Order) Order { return o.WithTotal(amount) }
}

// promotionChange returns a change that records a promotion.
func promotionChange(id PromotionID) change.Change[Order] {
	return func(o Order) Order { return o.WithPromotion(id) }
}

// versionChange returns a change that advances the version.
func versionChange() change.Change[Order] {
	return Order.BumpVersion
}

var requireCheckedOutHasItems modeling.Invariant[Order] = func(o Order) error {
	if o.Status() == OrderStatusCheckedOut && o.ItemsCount() == 0 {
		return ErrEmptyCheckedOutOrder
	}
	return nil
}

var requireEqualCurrencies modeling.Invariant[Order] = func(o Order) error {
	subtotal, discount, total := o.Subtotal(), o.Discount(), o.Total()
	if subtotal.Currency() != discount.Currency() || subtotal.Currency() != total.Currency() {
		return ErrCurrencyMismatch
	}
	return nil
}

var requireNonNegativeComponents modeling.Invariant[Order] = func(o Order) error {
	subtotal, discount, total := o.Subtotal(), o.Discount(), o.Total()
	if subtotal.Cents() < 0 || discount.Cents() < 0 || total.Cents() < 0 {
		return ErrNegativeMoney
	}
	return nil
}

// requireBalancedTotals requires subtotal == discount + total. Currency
// mismatches are reported by requireEqualCurrencies, so balance is skipped.
var requireBalancedTotals modeling.Invariant[Order] = func(o Order) error {
	subtotal, discount, total := o.Subtotal(), o.Discount(), o.Total()
	if subtotal.Currency() != total.Currency() {
		return nil
	}
	if subtotal.Cents() != discount.Cents()+total.Cents() {
		return ErrUnbalancedOrderTotals
	}
	return nil
}

// checkoutInvariants accumulates the post-checkout conditions. Decide runs
// them on the next state before any event is emitted.
var checkoutInvariants = modeling.Accumulate(
	requireCheckedOutHasItems,
	requireEqualCurrencies,
	requireNonNegativeComponents,
	requireBalancedTotals,
)
