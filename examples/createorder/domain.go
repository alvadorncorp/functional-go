// Pure business logic for the create-order demo: command, order value,
// order-created event, and the DecideOrder decision.
package main

import (
	"errors"
	"fmt"

	"go.alvadorn.org/functional/modeling"
)

var (
	ErrCustomerInactive  = errors.New("createorder: customer inactive")
	ErrInvalidQuantity   = errors.New("createorder: quantity must be positive")
	ErrInsufficientStock = errors.New("createorder: insufficient stock")
)

type CustomerID string

type ProductID string

type OrderID string

type Customer struct {
	ID     CustomerID
	Active bool
}

type Product struct {
	ID    ProductID
	Stock int
}

// CreateOrder is the transport-independent command. OrderID is the client's
// idempotency key: every retry and redelivery of the same request carries it,
// which is what lets the store accept repeated saves without duplicating.
type CreateOrder struct {
	OrderID    OrderID
	CustomerID CustomerID
	ProductID  ProductID
	Quantity   int
}

// Order is a small value type. Decisions pass it by value and nothing mutates
// it after publication.
type Order struct {
	ID         OrderID
	CustomerID CustomerID
	ProductID  ProductID
	Quantity   int
}

// OrderCreated records that an order was accepted.
type OrderCreated struct {
	OrderID    OrderID
	CustomerID CustomerID
	ProductID  ProductID
	Quantity   int
}

func NewOrderCreated(order Order) OrderCreated {
	return OrderCreated{
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		ProductID:  order.ProductID,
		Quantity:   order.Quantity,
	}
}

// DecideOrder is the pure business decision: validate, build the order, and
// emit its event. It touches no store and returns errors with the sentinel
// identity the transports map from.
func DecideOrder(
	customer Customer,
	product Product,
	command CreateOrder,
) (modeling.Decision[Order, OrderCreated], error) {
	var empty modeling.Decision[Order, OrderCreated]
	if !customer.Active {
		return empty, fmt.Errorf("%w: %s", ErrCustomerInactive, customer.ID)
	}
	if command.Quantity <= 0 {
		return empty, fmt.Errorf("%w: %d", ErrInvalidQuantity, command.Quantity)
	}
	if product.Stock < command.Quantity {
		return empty, fmt.Errorf("%w: have %d, want %d", ErrInsufficientStock, product.Stock, command.Quantity)
	}
	order := Order{
		ID:         command.OrderID,
		CustomerID: command.CustomerID,
		ProductID:  command.ProductID,
		Quantity:   command.Quantity,
	}
	return modeling.Decide[OrderCreated](order).Emit(NewOrderCreated(order)), nil
}
