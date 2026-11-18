package main

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"time"

	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/resilience"
)

const (
	httpBudget     = 3 * time.Second
	consumerBudget = 30 * time.Second
	retryAttempts  = 3
)

// Path one: HTTP. The server maps the JSON DTO to the command, budgets the
// whole exchange at 3s, and never retries on the server side; a failed request
// is the client's cue to resend, which is what the idempotency key is for.
type CreateOrderRequest struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
}

type orderResponse struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	ProductID  string `json:"product_id"`
	Quantity   int    `json:"quantity"`
}

type Handler struct {
	CreateOrder workflow.Workflow[CreateOrder, Order]
}

func (h Handler) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var dto CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		writeError(w, stdhttp.StatusBadRequest, err)
		return
	}
	input := CreateOrder{
		OrderID:    OrderID(dto.OrderID),
		CustomerID: CustomerID(dto.CustomerID),
		ProductID:  ProductID(dto.ProductID),
		Quantity:   dto.Quantity,
	}
	program := workflow.Provide(h.CreateOrder, input)
	program = resilience.Timeout(httpBudget, program)
	order, err := action.Run(r.Context(), program)
	if err != nil {
		writeError(w, statusFor(err), err)
		return
	}
	writeJSON(w, stdhttp.StatusCreated, orderResponse{
		OrderID:    string(order.ID),
		CustomerID: string(order.CustomerID),
		ProductID:  string(order.ProductID),
		Quantity:   order.Quantity,
	})
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return stdhttp.StatusGatewayTimeout
	case errors.Is(err, context.Canceled):
		return stdhttp.StatusRequestTimeout
	case errors.Is(err, ErrCustomerInactive),
		errors.Is(err, ErrInvalidQuantity),
		errors.Is(err, ErrInsufficientStock):
		return stdhttp.StatusUnprocessableEntity
	default:
		return stdhttp.StatusInternalServerError
	}
}

// Path two: a message consumer. It maps its own wire DTO to the same command,
// retries transient failures up to 3 attempts inside one 30s budget, and never
// retries domain rejections. The broker may redeliver at any time; the store's
// upsert plus the transaction boundary make every delivery idempotent.
type OrderMessage struct {
	OrderID  string `json:"order_id"`
	Customer string `json:"customer"`
	Product  string `json:"product"`
	Units    int    `json:"units"`
}

type Consumer struct {
	CreateOrder workflow.Workflow[CreateOrder, Order]
	Backoff     resilience.Backoff // optional composition-root override
}

func (c Consumer) Handle(ctx context.Context, message OrderMessage) error {
	input := CreateOrder{
		OrderID:    OrderID(message.OrderID),
		CustomerID: CustomerID(message.Customer),
		ProductID:  ProductID(message.Product),
		Quantity:   message.Units,
	}
	backoff := c.Backoff
	if backoff == nil {
		backoff = exponentialBackoff
	}
	// Retry sits around the whole effectful workflow, outside the boundary: a
	// failed attempt rolls back entirely, so the retry re-runs from a clean
	// slate instead of repeating half-finished effects.
	program := workflow.Provide(c.CreateOrder, input)
	program = resilience.Retry(resilience.RetryPolicy{
		Attempts: retryAttempts,
		When:     isTransient,
		Backoff:  backoff,
	}, program)
	program = resilience.Timeout(consumerBudget, program) // one budget covers every attempt
	_, err := action.Run(ctx, program)
	return err
}

// exponentialBackoff is a capped doubling delay: 100ms, 200ms, 400ms, up to 1s.
func exponentialBackoff(attempt int) time.Duration {
	const (
		baseDelay    = 100 * time.Millisecond
		maxDelay     = time.Second
		maxDoublings = 4 // caps the growth so large attempt counts cannot overflow
	)
	delay := baseDelay
	for range min(max(attempt-1, 0), maxDoublings) {
		delay = min(delay*2, maxDelay)
	}
	return delay
}

func writeJSON(w stdhttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w stdhttp.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

type errorResponse struct {
	Error string `json:"error"`
}
