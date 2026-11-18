package main

import (
	"context"
	"slices"

	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/values/either"
	"go.alvadorn.org/functional/values/optional"
	"go.alvadorn.org/functional/workflow"
	"go.alvadorn.org/functional/workflow/action"
	"go.alvadorn.org/functional/workflow/transaction"
)

// CheckoutRequest is the application input for one checkout attempt.
type CheckoutRequest struct {
	OrderID    OrderID
	CustomerID CustomerID
	CouponCode optional.Optional[CouponCode]
	Payment    PaymentChoice
}

// OutboxMessage is an integration message awaiting delivery: one row per
// emitted domain event, appended in emission order.
type OutboxMessage struct {
	Type        string
	AggregateID string
}

// CheckoutResult is what a successful run returns to its caller.
type CheckoutResult struct {
	Order  Order
	Outbox immutable.List[OutboxMessage]
}

// eventMessage translates a domain event into its integration message.
func eventMessage(event Event) OutboxMessage {
	id := ""
	switch typed := event.(type) {
	case *OrderCheckedOut:
		id = string(typed.OrderID)
	case DiscountApplied:
		id = string(typed.OrderID)
	}
	return OutboxMessage{Type: event.EventName(), AggregateID: id}
}

// Orders is what the application can do with orders outside a transaction:
// read one. Writes live only on TxOrders inside TxScope.
type Orders struct {
	Load func(OrderID) action.Action[Order]
}

// Customers loads a loyalty profile outside the transaction.
type Customers struct {
	Load func(CustomerID) action.Action[CustomerProfile]
}

// Coupons loads a coupon outside the transaction.
type Coupons struct {
	Load func(CouponCode) action.Action[Coupon]
}

// TxOrders saves an order guarded by the loaded version.
type TxOrders struct {
	Save func(Order, OrderVersion) action.Action[struct{}]
}

// TxPayments records the payment consequence selected by the decision.
type TxPayments struct {
	ReserveImmediate func(OrderID, ImmediatePayment, Money) action.Action[struct{}]
	RecordDeferred   func(OrderID, DeferredPayment, Money) action.Action[struct{}]
}

// TxOutbox appends an integration message inside the same transaction as the
// state change it accompanies.
type TxOutbox struct {
	Append func(OutboxMessage) action.Action[struct{}]
}

// TxCoupons consumes a coupon inside the same transaction as the rest of the
// checkout writes.
type TxCoupons struct {
	Consume func(CouponCode) action.Action[struct{}]
}

// TxScope is the set of write capabilities bound to one concrete transaction.
// Every function writes through the same staged handle, so payment, coupon
// consumption, order, and outbox rows commit or roll back together.
type TxScope struct {
	Orders   TxOrders
	Payments TxPayments
	Outbox   TxOutbox
	Coupons  TxCoupons
}

// CheckoutDependencies supplies the read capabilities and the typed
// transaction boundary at composition time.
type CheckoutDependencies struct {
	Orders    Orders
	Customers Customers
	Coupons   Coupons
	Boundary  transaction.Boundary[TxScope]
}

// LoadedCheckout pairs the request with state loaded outside any transaction.
type LoadedCheckout struct {
	Request  CheckoutRequest
	Order    Order
	Customer CustomerProfile
}

// EnrichedCheckout adds the coupon loaded when the request asked for one.
type EnrichedCheckout struct {
	LoadedCheckout
	Coupon optional.Optional[Coupon]
}

// Checkout builds the lazy checkout workflow. Nothing runs until action.Run:
// construction and applying the workflow to a request perform no I/O.
func Checkout(deps CheckoutDependencies) workflow.Workflow[CheckoutRequest, CheckoutResult] {
	return workflow.From(func(req CheckoutRequest) action.Action[CheckoutResult] {
		return action.FlatMap(loadContext(deps, &req), func(loaded LoadedCheckout) action.Action[CheckoutResult] {
			couponLookup := loadCoupon(deps, &loaded.Request)
			return action.FlatMap(couponLookup, func(coupon optional.Optional[Coupon]) action.Action[CheckoutResult] {
				return commitCheckout(deps, &EnrichedCheckout{LoadedCheckout: loaded, Coupon: coupon})
			})
		})
	})
}

// loadContext loads order and customer independently, then combines them with
// the request. Construction performs no I/O.
func loadContext(deps CheckoutDependencies, req *CheckoutRequest) action.Action[LoadedCheckout] {
	return action.Both(
		deps.Orders.Load(req.OrderID),
		deps.Customers.Load(req.CustomerID),
		func(order Order, customer CustomerProfile) LoadedCheckout {
			return LoadedCheckout{Request: *req, Order: order, Customer: customer}
		},
	)
}

// loadCoupon returns a pure None when the request has no coupon code;
// otherwise it maps the coupon load into Some.
func loadCoupon(deps CheckoutDependencies, req *CheckoutRequest) action.Action[optional.Optional[Coupon]] {
	code, ok := req.CouponCode.Get()
	if !ok {
		return action.Pure(optional.None[Coupon]())
	}
	return action.Map(deps.Coupons.Load(code), optional.Some)
}

// commitCheckout opens one typed boundary after the reads. Inside the same
// scope the pure decision runs first, then coupon consumption, payment, save,
// and ordered outbox appends. Any error leaves the staging area to be
// discarded, which is the rollback.
func commitCheckout(deps CheckoutDependencies, enriched *EnrichedCheckout) action.Action[CheckoutResult] {
	return transaction.Within(deps.Boundary, func(scope TxScope) action.Action[CheckoutResult] {
		return action.Do(func(ctx context.Context) (CheckoutResult, error) {
			return persistCheckout(ctx, deps, scope, enriched)
		})
	})
}

// persistCheckout runs the pure decision and the staged write sequence in
// fixed order; the first error stops everything before commit.
func persistCheckout(
	ctx context.Context,
	deps CheckoutDependencies,
	scope TxScope,
	enriched *EnrichedCheckout,
) (CheckoutResult, error) {
	decision, err := action.Run(ctx, decideCheckout(enriched))
	if err != nil {
		return CheckoutResult{}, err
	}
	if consumeErr := consumeCouponIfPresent(ctx, scope, enriched); consumeErr != nil {
		return CheckoutResult{}, consumeErr
	}
	if payErr := applyPayment(ctx, scope.Payments, enriched.Request.Payment, decision); payErr != nil {
		return CheckoutResult{}, payErr
	}
	if _, saveErr := action.Run(ctx, scope.Orders.Save(decision.State(), enriched.Order.Version())); saveErr != nil {
		return CheckoutResult{}, saveErr
	}
	messages, appendErr := appendOutbox(ctx, scope.Outbox, decision)
	if appendErr != nil {
		return CheckoutResult{}, appendErr
	}
	return CheckoutResult{Order: decision.State(), Outbox: immutable.ListOf(messages...)}, nil
}

// decideCheckout bridges the pure domain Decide into an Action via
// MapFallible; construction performs no I/O.
func decideCheckout(enriched *EnrichedCheckout) action.Action[modeling.Decision[Order, Event]] {
	return action.MapFallible(action.Pure(enriched), func(e *EnrichedCheckout) (modeling.Decision[Order, Event], error) {
		return Decide(e.Order, &DecideInput{Customer: e.Customer, Payment: e.Request.Payment, Coupon: e.Coupon})
	})
}

// consumeCouponIfPresent consumes the loaded coupon selected by
// optional.Match.
func consumeCouponIfPresent(ctx context.Context, scope TxScope, enriched *EnrichedCheckout) error {
	return optional.Match(enriched.Coupon,
		func() error { return nil },
		func(coupon Coupon) error {
			_, err := action.Run(ctx, scope.Coupons.Consume(coupon.Code))
			return err
		},
	)
}

// applyPayment executes exactly one payment branch selected by either.Match.
func applyPayment(
	ctx context.Context,
	payments TxPayments,
	choice PaymentChoice,
	decision modeling.Decision[Order, Event],
) error {
	state := decision.State()
	return either.Match(choice,
		func(immediate ImmediatePayment) error {
			_, err := action.Run(ctx, payments.ReserveImmediate(state.ID(), immediate, state.Total()))
			return err
		},
		func(deferred DeferredPayment) error {
			_, err := action.Run(ctx, payments.RecordDeferred(state.ID(), deferred, state.Total()))
			return err
		},
	)
}

// appendOutbox translates and appends each decision event in emission order.
// action.Traverse binds the next event only after the previous append
// succeeded, so the first failure stops the sequence and rolls back the
// whole transaction.
func appendOutbox(
	ctx context.Context,
	outbox TxOutbox,
	decision modeling.Decision[Order, Event],
) ([]OutboxMessage, error) {
	appended := action.Traverse(slices.Collect(decision.Events()), func(event Event) action.Action[OutboxMessage] {
		return action.Tap(action.Map(action.Pure(event), eventMessage), outbox.Append)
	})
	return action.Run(ctx, appended)
}
