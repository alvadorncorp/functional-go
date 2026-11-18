package main

import (
	"context"
	"errors"
	"strings"
	"sync"

	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// Transient marks a failure a dependency may recover from on a later attempt:
// a connection that comes back, a broker that unblocks. This local type is the
// demo's stand-in for a fault-classification package, which it deliberately
// avoids: retry policy needs exactly one predicate, isTransient.
type Transient struct{ Cause error }

func (e *Transient) Error() string { return "transient: " + e.Cause.Error() }

func (e *Transient) Unwrap() error { return e.Cause }

func transient(cause error) error { return &Transient{Cause: cause} }

func isTransient(err error) bool {
	var marker *Transient
	return errors.As(err, &marker)
}

// Store is the demo database. Committed state lives in orders and outbox; a
// transaction stages its writes and only the commit merges them, so a failed
// attempt leaves the store untouched and a retry starts from a clean slate.
type Store struct {
	mu             sync.Mutex
	trace          []string
	customers      map[CustomerID]Customer
	products       map[ProductID]Product
	orders         map[OrderID]Order
	outbox         []OrderCreated
	outboxFailures int // remaining transient failures to inject into outbox appends
}

func NewStore() *Store {
	return &Store{
		customers: map[CustomerID]Customer{
			"cust-1": {ID: "cust-1", Active: true},
		},
		products: map[ProductID]Product{
			"prod-1": {ID: "prod-1", Stock: 10},
		},
		orders: make(map[OrderID]Order),
	}
}

// FailNextOutboxAppends arms a deterministic transient fault: the next n outbox
// appends fail, forcing a rollback and letting the retry show a clean recovery.
func (s *Store) FailNextOutboxAppends(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outboxFailures = n
}

func (s *Store) Dependencies() Dependencies {
	return Dependencies{
		Customers:   Customers{Find: s.findCustomer},
		Products:    Products{Find: s.findProduct},
		Transaction: s.Boundary(),
	}
}

func (s *Store) findCustomer(id CustomerID) action.Action[Customer] {
	return action.Do(func(ctx context.Context) (Customer, error) {
		if err := ctx.Err(); err != nil {
			return Customer{}, err
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.recordLocked("load-customer")
		customer, ok := s.customers[id]
		if !ok {
			return Customer{}, errors.New("store: customer " + string(id) + " not found")
		}
		return customer, nil
	})
}

func (s *Store) findProduct(id ProductID) action.Action[Product] {
	return action.Do(func(ctx context.Context) (Product, error) {
		if err := ctx.Err(); err != nil {
			return Product{}, err
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.recordLocked("load-product")
		product, ok := s.products[id]
		if !ok {
			return Product{}, errors.New("store: product " + string(id) + " not found")
		}
		return product, nil
	})
}

// Boundary stages all writes for the body and merges them only on success. Any
// error or panic drops the stage and records a rollback: effects are all or
// nothing, which is what makes a retry around the whole workflow safe.
func (s *Store) Boundary() transaction.Boundary[TxScope] {
	return func(ctx context.Context, body func(context.Context, TxScope) error) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		tx := &storeTx{store: s, orders: make(map[OrderID]Order)}
		s.recordLocked("transaction")
		committed := false
		defer func() {
			if !committed {
				s.recordLocked("rollback")
			}
		}()
		if err := body(ctx, tx.scope()); err != nil {
			return err
		}
		tx.commit()
		s.recordLocked("commit")
		committed = true
		return nil
	}
}

type storeTx struct {
	store  *Store
	orders map[OrderID]Order
	outbox []OrderCreated
}

func (tx *storeTx) scope() TxScope {
	return TxScope{
		SaveOrder: func(order Order) action.Action[bool] {
			return action.Do(func(context.Context) (bool, error) {
				return tx.saveOrder(order)
			})
		},
		AppendOutbox: func(events ...OrderCreated) action.Action[struct{}] {
			return action.Do(func(context.Context) (struct{}, error) {
				return struct{}{}, tx.appendOutbox(events)
			})
		},
	}
}

func (tx *storeTx) saveOrder(order Order) (bool, error) {
	s := tx.store
	s.recordLocked("save-order")
	existing, exists := tx.orders[order.ID]
	if !exists {
		existing, exists = s.orders[order.ID]
	}
	if exists {
		if existing != order {
			return false, errors.New("store: order ID already used with different input")
		}
		return false, nil
	}
	tx.orders[order.ID] = order
	return true, nil
}

func (tx *storeTx) appendOutbox(events []OrderCreated) error {
	s := tx.store
	s.recordLocked("append-outbox")
	if s.outboxFailures > 0 {
		s.outboxFailures--
		return transient(errors.New("outbox unavailable"))
	}
	tx.outbox = append(tx.outbox, events...)
	return nil
}

func (tx *storeTx) commit() {
	s := tx.store
	for id, order := range tx.orders {
		s.orders[id] = order
	}
	s.outbox = append(s.outbox, tx.outbox...)
}

func (s *Store) recordLocked(step string) {
	s.trace = append(s.trace, step)
}

func (s *Store) Trace() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.trace, ", ")
}

func (s *Store) ResetTrace() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trace = nil
}

func (s *Store) Totals() (orders, events int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.orders), len(s.outbox)
}
