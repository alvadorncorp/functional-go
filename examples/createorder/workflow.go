package main

import (
	"slices"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/plan"
	"go.alvadorn.org/functional/workflow/transaction"
)

// Capabilities are functions supplied by the composition root. The workflow
// captures them at construction; the input and the context arrive later.
type Customers struct {
	Find func(CustomerID) action.Action[Customer]
}

type Products struct {
	Find func(ProductID) action.Action[Product]
}

// TxScope carries the capabilities bound to one concrete transaction. SaveOrder
// returns whether the upsert created the order, so a redelivery can skip the
// outbox instead of emitting a second OrderCreated for the same order.
type TxScope struct {
	SaveOrder    func(Order) action.Action[bool]
	AppendOutbox func(...OrderCreated) action.Action[struct{}]
}

type Dependencies struct {
	Customers   Customers
	Products    Products
	Transaction transaction.Boundary[TxScope]
}

// CreateOrderWorkflow loads the customer, loads the product, decides, and
// commits inside one transaction. It is delivery-agnostic: the HTTP handler and
// the message consumer below run this exact workflow.
func CreateOrderWorkflow(deps Dependencies) workflow.Workflow[CreateOrder, Order] {
	return func(input CreateOrder) action.Action[Order] {
		return action.FlatMap(deps.Customers.Find(input.CustomerID),
			func(customer Customer) action.Action[Order] {
				return action.FlatMap(deps.Products.Find(input.ProductID),
					func(product Product) action.Action[Order] {
						decision, err := DecideOrder(customer, product, input)
						if err != nil {
							return action.Fail[Order](err)
						}
						return commitOrder(deps.Transaction, decision.State(), slices.Collect(decision.Events()))
					})
			})
	}
}

// commitOrder persists the decision inside one transaction. Order is passed by
// value deliberately; nothing mutates it after publication.
func commitOrder(
	boundary transaction.Boundary[TxScope],
	order Order,
	events []OrderCreated,
) action.Action[Order] {
	return transaction.Within(boundary, func(scope TxScope) action.Action[Order] {
		return action.FlatMap(scope.SaveOrder(order), func(created bool) action.Action[Order] {
			if !created {
				return action.Pure(order) // redelivery: the order exists; emit no second event
			}
			return action.Map(scope.AppendOutbox(events...), func(struct{}) Order { return order })
		})
	})
}

// CreateOrderPlan documents the workflow's shape. It is declared by hand, apart
// from CreateOrderWorkflow: the library never derives a plan from code, so this
// declaration stays an honest summary only while the plan and the workflow are
// edited together. The demo prints it so any drift is visible.
var CreateOrderPlan = plan.Sequence(
	plan.Step("load-customer"),
	plan.Step("load-product"),
	plan.Step("decide-order"),
	plan.Transaction(
		plan.Step("save-order"),
		plan.Step("append-outbox"),
	),
)
