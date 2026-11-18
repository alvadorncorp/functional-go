// Command createorder is the memory-backed demo of the create-order workflow,
// ported to the public go.alvadorn.org/functional libraries: one workflow, two
// delivery paths (HTTP handler exercised in process and a message consumer),
// different timeout budgets, and selective retry of transient failures only.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"strings"

	"go.alvadorn.org/functional/workflow"
)

const (
	demoCustomer      = "cust-1"
	demoProduct       = "prod-1"
	demoQuantity      = 2
	overStockQuantity = 999
	firstHTTPOrderID  = "ord-http-1"
	secondHTTPOrderID = "ord-http-2"
	firstMessageID    = "ord-msg-1"
	secondMessageID   = "ord-msg-2"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "createorder demo failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	store := NewStore()
	delivery := CreateOrderWorkflow(store.Dependencies())

	printPlan()
	runHTTPDemo(delivery, store)
	if err := runConsumerDemo(ctx, delivery, store); err != nil {
		return err
	}
	printSummary(store)
	return nil
}

func printPlan() {
	fmt.Println("create-order demo (memory-backed, no network service)")
	fmt.Println()
	fmt.Println("declared plan (independent of the workflow code):")
	fmt.Printf("  %s\n\n", CreateOrderPlan.String())
}

// runHTTPDemo exercises the handler in process with httptest: no listener, no
// network. A valid request returns 201; an over-stock request is a domain
// rejection that maps to 422 and is never retried on the server side.
func runHTTPDemo(delivery workflow.Workflow[CreateOrder, Order], store *Store) {
	handler := Handler{CreateOrder: delivery}
	for _, request := range []CreateOrderRequest{
		{OrderID: firstHTTPOrderID, CustomerID: demoCustomer, ProductID: demoProduct, Quantity: demoQuantity},
		{OrderID: secondHTTPOrderID, CustomerID: demoCustomer, ProductID: demoProduct, Quantity: overStockQuantity},
	} {
		payload := mustJSON(request)
		req := httptest.NewRequest(stdhttp.MethodPost, "/orders", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		fmt.Printf("POST /orders %s\n  -> %d %s\n", payload, rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	orders, _ := store.Totals()
	fmt.Printf("  http path done: orders=%d\n\n", orders)
}

func mustJSON(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err) // demo DTOs always marshal
	}
	return payload
}

// runConsumerDemo delivers three messages. The first hits an armed transient
// outbox failure: attempt 1 rolls back entirely, attempt 2 succeeds, and the
// store holds exactly one order and one event. The second delivery is a broker
// redelivery of the same message; the existing order is left in place. The third is
// a domain rejection, which is not transient and so is never retried.
func runConsumerDemo(ctx context.Context, delivery workflow.Workflow[CreateOrder, Order], store *Store) error {
	consumer := Consumer{CreateOrder: delivery}

	message := OrderMessage{OrderID: firstMessageID, Customer: demoCustomer, Product: demoProduct, Units: demoQuantity}
	store.ResetTrace()
	store.FailNextOutboxAppends(1)
	err := consumer.Handle(ctx, message)
	fmt.Printf("delivery 1 (%s) with a transient outbox failure: Handle returned %v\n", firstMessageID, err)
	fmt.Printf("  trace: %s\n", store.Trace())
	if err != nil {
		return fmt.Errorf("retry did not commit: %w", err)
	}
	fmt.Println("  a failed attempt rolled back, the retry committed once")

	store.ResetTrace()
	err = consumer.Handle(ctx, message)
	fmt.Printf("delivery 2 (redelivery of %s): Handle returned %v\n", firstMessageID, err)
	fmt.Printf("  trace: %s\n", store.Trace())
	if err != nil {
		return fmt.Errorf("redelivery failed: %w", err)
	}
	fmt.Println("  save-order found the existing order and skipped the outbox")

	rejected := OrderMessage{OrderID: secondMessageID, Customer: demoCustomer, Product: demoProduct, Units: 0}
	store.ResetTrace()
	err = consumer.Handle(ctx, rejected)
	fmt.Printf("delivery 3 (%s, quantity 0): Handle returned %v\n", secondMessageID, err)
	fmt.Printf("  trace: %s\n\n", store.Trace())
	if !errors.Is(err, ErrInvalidQuantity) {
		return fmt.Errorf("expected an invalid quantity: %w", err)
	}
	fmt.Println("  a domain rejection is not transient, so no retry ran")
	return nil
}

func printSummary(store *Store) {
	orders, events := store.Totals()
	fmt.Println("retry sat around the whole workflow, the boundary rolled back the failed")
	fmt.Println("attempt, and redelivery found the existing order without creating a duplicate.")
	fmt.Printf("final store: orders=%d, outbox events=%d\n", orders, events)
}
