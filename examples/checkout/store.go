package main

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.alvadorn.org/functional/values/either"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// Injected failure points consumed by FailNext.
const (
	failSave    = "save"
	failPayment = "payment"
	failConsume = "consume"
	failOutbox  = "outbox"
)

// errInjectedFailure is the sentinel wrapped by every injected failure so
// callers can recognize the rollback proof with errors.Is.
var errInjectedFailure = errors.New("injected failure")

var (
	errOrderNotFound    = errors.New("ordering store: order not found")
	errCustomerNotFound = errors.New("ordering store: customer not found")
	errCouponNotFound   = errors.New("ordering store: coupon not found")
	errVersionConflict  = errors.New("ordering store: order version conflict")
)

func injectedFailure(op string) error {
	return fmt.Errorf("ordering store: %s failed: %w", op, errInjectedFailure)
}

// PaymentRecord is one committed payment consequence. The Either carries the
// charged method id, or the deferred terms in days.
type PaymentRecord struct {
	OrderID OrderID
	Total   Money
	Payment either.Either[PaymentMethodID, uint16]
}

// stagedTxn holds the uncommitted writes of one boundary invocation. Every
// staged capability writes into this copy; publish replaces committed state
// only after the callback succeeds, so any failure is a rollback.
type stagedTxn struct {
	orders   map[OrderID]Order
	coupons  map[CouponCode]Coupon
	payments map[OrderID]PaymentRecord
	outbox   []OutboxMessage
}

// Store holds committed checkout state behind a mutex. Transactional writes
// stage on a private clone until the boundary callback succeeds, then publish
// atomically; a failed callback discards the staging area, which is the
// rollback.
type Store struct {
	mu        sync.Mutex
	orders    map[OrderID]Order
	customers map[CustomerID]CustomerProfile
	coupons   map[CouponCode]Coupon
	payments  map[OrderID]PaymentRecord
	outbox    []OutboxMessage
	failures  map[string]bool
}

// NewStore seeds initial committed state.
func NewStore(orders []Order, customers []CustomerProfile, coupons []Coupon) *Store {
	store := &Store{
		orders:    make(map[OrderID]Order, len(orders)),
		customers: make(map[CustomerID]CustomerProfile, len(customers)),
		coupons:   make(map[CouponCode]Coupon, len(coupons)),
		payments:  make(map[OrderID]PaymentRecord),
		failures:  make(map[string]bool),
	}
	for _, order := range orders {
		store.orders[order.ID()] = order
	}
	for _, customer := range customers {
		store.customers[customer.ID] = customer
	}
	for _, coupon := range coupons {
		store.coupons[coupon.Code] = coupon
	}
	return store
}

// FailNext arms a one-shot failure for the next attempt of op: save, payment,
// consume, or outbox.
func (s *Store) FailNext(op string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[op] = true
}

// takeFailure consumes an armed one-shot failure, if any.
func (s *Store) takeFailure(op string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failures[op] {
		delete(s.failures, op)
		return injectedFailure(op)
	}
	return nil
}

// Snapshot is an owned copy of committed state for assertions and printing.
// Order handles, coupons, and payment rows are immutable values; outbox
// payload bytes are never mutated after encoding.
type Snapshot struct {
	Orders    map[OrderID]Order
	Customers map[CustomerID]CustomerProfile
	Coupons   map[CouponCode]Coupon
	Payments  map[OrderID]PaymentRecord
	Outbox    []OutboxMessage
}

// Snapshot copies every committed collection under the store lock.
func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := Snapshot{
		Orders:    make(map[OrderID]Order, len(s.orders)),
		Customers: make(map[CustomerID]CustomerProfile, len(s.customers)),
		Coupons:   make(map[CouponCode]Coupon, len(s.coupons)),
		Payments:  make(map[OrderID]PaymentRecord, len(s.payments)),
	}
	for id, order := range s.orders {
		snap.Orders[id] = order
	}
	for id, customer := range s.customers {
		snap.Customers[id] = customer
	}
	for code, coupon := range s.coupons {
		snap.Coupons[code] = coupon
	}
	for id, record := range s.payments {
		snap.Payments[id] = record
	}
	snap.Outbox = append([]OutboxMessage(nil), s.outbox...)
	return snap
}

// stage clones committed write state into a fresh staging area.
func (s *Store) stage() *stagedTxn {
	s.mu.Lock()
	defer s.mu.Unlock()
	staged := &stagedTxn{
		orders:   make(map[OrderID]Order, len(s.orders)),
		coupons:  make(map[CouponCode]Coupon, len(s.coupons)),
		payments: make(map[OrderID]PaymentRecord, len(s.payments)),
	}
	for id, order := range s.orders {
		staged.orders[id] = order
	}
	for code, coupon := range s.coupons {
		staged.coupons[code] = coupon
	}
	for id, record := range s.payments {
		staged.payments[id] = record
	}
	staged.outbox = append([]OutboxMessage(nil), s.outbox...)
	return staged
}

// publish swaps the staging area into committed state. Staged operations
// already validated against this baseline, so no second conflict pass is
// needed.
func (s *Store) publish(staged *stagedTxn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders = staged.orders
	s.coupons = staged.coupons
	s.payments = staged.payments
	s.outbox = staged.outbox
}

// Boundary stages writes on a clone of committed state, runs the callback
// without holding the store lock, and publishes only when the callback
// succeeds. A failed callback leaves the staging area to be discarded.
func (s *Store) Boundary() transaction.Boundary[TxScope] {
	return func(ctx context.Context, run func(context.Context, TxScope) error) error {
		staged := s.stage()
		if err := run(ctx, s.scope(staged)); err != nil {
			return err
		}
		s.publish(staged)
		return nil
	}
}

// scope binds the staged write capabilities to one boundary invocation.
func (s *Store) scope(staged *stagedTxn) TxScope {
	return TxScope{
		Orders:   TxOrders{Save: s.stageSave(staged)},
		Payments: TxPayments{ReserveImmediate: s.stageImmediate(staged), RecordDeferred: s.stageDeferred(staged)},
		Outbox:   TxOutbox{Append: s.stageAppend(staged)},
		Coupons:  TxCoupons{Consume: s.stageConsume(staged)},
	}
}

// Orders exposes the committed order read capability.
func (s *Store) Orders() Orders {
	return Orders{Load: func(id OrderID) action.Action[Order] {
		return action.MapFallible(action.Pure(id), func(loaded OrderID) (Order, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			order, ok := s.orders[loaded]
			if !ok {
				return Order{}, errOrderNotFound
			}
			return order, nil
		})
	}}
}

// Customers exposes the committed loyalty read capability.
func (s *Store) Customers() Customers {
	return Customers{Load: func(id CustomerID) action.Action[CustomerProfile] {
		return action.MapFallible(action.Pure(id), func(loaded CustomerID) (CustomerProfile, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			customer, ok := s.customers[loaded]
			if !ok {
				return CustomerProfile{}, errCustomerNotFound
			}
			return customer, nil
		})
	}}
}

// Coupons exposes the committed coupon read capability.
func (s *Store) Coupons() Coupons {
	return Coupons{Load: func(code CouponCode) action.Action[Coupon] {
		return action.MapFallible(action.Pure(code), func(loaded CouponCode) (Coupon, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			coupon, ok := s.coupons[loaded]
			if !ok {
				return Coupon{}, errCouponNotFound
			}
			return coupon, nil
		})
	}}
}

func (s *Store) stageSave(staged *stagedTxn) func(Order, OrderVersion) action.Action[struct{}] {
	return func(order Order, expected OrderVersion) action.Action[struct{}] {
		return action.MapFallible(action.Pure(order), func(saved Order) (struct{}, error) {
			if err := s.takeFailure(failSave); err != nil {
				return struct{}{}, err
			}
			current, ok := staged.orders[saved.ID()]
			if ok && current.Version() != expected {
				return struct{}{}, errVersionConflict
			}
			staged.orders[saved.ID()] = saved
			return struct{}{}, nil
		})
	}
}

func (s *Store) stageImmediate(staged *stagedTxn) func(OrderID, ImmediatePayment, Money) action.Action[struct{}] {
	return func(id OrderID, payment ImmediatePayment, total Money) action.Action[struct{}] {
		return action.MapFallible(action.Pure(payment), func(charged ImmediatePayment) (struct{}, error) {
			if err := s.takeFailure(failPayment); err != nil {
				return struct{}{}, err
			}
			staged.payments[id] = PaymentRecord{
				OrderID: id,
				Total:   total,
				Payment: either.Left[PaymentMethodID, uint16](charged.Method),
			}
			return struct{}{}, nil
		})
	}
}

func (s *Store) stageDeferred(staged *stagedTxn) func(OrderID, DeferredPayment, Money) action.Action[struct{}] {
	return func(id OrderID, payment DeferredPayment, total Money) action.Action[struct{}] {
		return action.MapFallible(action.Pure(payment), func(recorded DeferredPayment) (struct{}, error) {
			if err := s.takeFailure(failPayment); err != nil {
				return struct{}{}, err
			}
			staged.payments[id] = PaymentRecord{
				OrderID: id,
				Total:   total,
				Payment: either.Right[PaymentMethodID](recorded.Days),
			}
			return struct{}{}, nil
		})
	}
}

// stageAppend stages one integration message after the state change it
// accompanies.
func (s *Store) stageAppend(staged *stagedTxn) func(OutboxMessage) action.Action[struct{}] {
	return func(message OutboxMessage) action.Action[struct{}] {
		return action.MapFallible(action.Pure(message), func(appended OutboxMessage) (struct{}, error) {
			if err := s.takeFailure(failOutbox); err != nil {
				return struct{}{}, err
			}
			staged.outbox = append(staged.outbox, appended)
			return struct{}{}, nil
		})
	}
}

// stageConsume redeems a coupon atomically with the rest of the writes. It
// re-checks validity at consume time, so expired or exhausted coupons fail
// the transaction even when the read path saw them as usable.
func (s *Store) stageConsume(staged *stagedTxn) func(CouponCode) action.Action[struct{}] {
	return func(code CouponCode) action.Action[struct{}] {
		return action.MapFallible(action.Pure(code), func(consumed CouponCode) (struct{}, error) {
			coupon, ok := staged.coupons[consumed]
			if !ok {
				return struct{}{}, errCouponNotFound
			}
			if err := s.takeFailure(failConsume); err != nil {
				return struct{}{}, err
			}
			redeemed, err := coupon.Redeem()
			if err != nil {
				return struct{}{}, err
			}
			staged.coupons[consumed] = redeemed
			return struct{}{}, nil
		})
	}
}
