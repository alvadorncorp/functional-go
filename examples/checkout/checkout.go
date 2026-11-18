package main

import (
	"errors"
	"iter"

	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/modeling/change"
	"go.alvadorn.org/functional/modeling/policy"
	"go.alvadorn.org/functional/modeling/projection"
	"go.alvadorn.org/functional/modeling/reducer"
	"go.alvadorn.org/functional/modeling/specification"
	"go.alvadorn.org/functional/primitives/predicate"
	"go.alvadorn.org/functional/values/either"
	"go.alvadorn.org/functional/values/optional"
	"go.alvadorn.org/functional/values/result"
)

// minimumCompletedOrders is the loyalty threshold for premium benefits.
const minimumCompletedOrders = 5

// CustomerTier classifies a customer's commercial relationship.
type CustomerTier uint8

const (
	// CustomerStandard customers earn no standing discount.
	CustomerStandard CustomerTier = iota
	// CustomerPremium customers qualify for the loyalty discount.
	CustomerPremium
)

// CustomerProfile is an immutable loyalty snapshot loaded outside the
// transaction.
type CustomerProfile struct {
	ID              CustomerID
	Tier            CustomerTier
	CompletedOrders int
	Blocked         bool
}

var (
	isPremiumTier specification.Specification[CustomerProfile] = func(c CustomerProfile) bool {
		return c.Tier == CustomerPremium
	}
	hasEnoughOrders specification.Specification[CustomerProfile] = func(c CustomerProfile) bool {
		return c.CompletedOrders >= minimumCompletedOrders
	}
	isBlocked specification.Specification[CustomerProfile] = func(c CustomerProfile) bool {
		return c.Blocked
	}

	// LoyaltyEligibility composes the loyalty predicates in fixed order.
	LoyaltyEligibility = specification.All(hasEnoughOrders, specification.Not(isBlocked), isPremiumTier)
)

var (
	ErrInvalidCouponPercent = errors.New("coupon: percent above the maximum")
	ErrCouponExpired        = errors.New("coupon: expired")
	ErrCouponExhausted      = errors.New("coupon: no remaining uses")
)

// maxCouponPercent bounds coupon percentages.
const maxCouponPercent uint8 = 100

// Coupon is a value object describing a reusable discount code.
type Coupon struct {
	Code          CouponCode
	Percent       uint8
	Expired       bool
	RemainingUses uint32
}

// NewCoupon rejects percentages above the maximum.
func NewCoupon(code CouponCode, percent uint8, expired bool, remainingUses uint32) result.Result[Coupon, error] {
	if percent > maxCouponPercent {
		return result.Err[Coupon](ErrInvalidCouponPercent)
	}
	return result.Ok[Coupon, error](Coupon{
		Code:          code,
		Percent:       percent,
		Expired:       expired,
		RemainingUses: remainingUses,
	})
}

var couponNotExpired modeling.Rule[Coupon] = func(c Coupon) error {
	if c.Expired {
		return ErrCouponExpired
	}
	return nil
}

var couponHasRemainingUses modeling.Rule[Coupon] = func(c Coupon) error {
	if c.RemainingUses == 0 {
		return ErrCouponExhausted
	}
	return nil
}

// ValidCoupon rejects an unusable coupon: expiry first, then remaining uses.
var ValidCoupon = modeling.AllRules(couponNotExpired, couponHasRemainingUses)

// Redeem validates the coupon and consumes one use, returning the updated
// value. The receiver is never mutated.
func (c Coupon) Redeem() (Coupon, error) {
	if err := ValidCoupon(c); err != nil {
		return Coupon{}, err
	}
	next := c
	next.RemainingUses--
	return next, nil
}

// ImmediatePayment charges the configured method right away.
type ImmediatePayment struct {
	Method PaymentMethodID
}

// DeferredPayment records net-terms days for later settlement.
type DeferredPayment struct {
	Days uint16
}

// PaymentChoice is exactly one of the two payment effects.
type PaymentChoice = either.Either[ImmediatePayment, DeferredPayment]

// ImmediatePaymentChoice selects the immediate branch.
func ImmediatePaymentChoice(p ImmediatePayment) PaymentChoice {
	return either.Left[ImmediatePayment, DeferredPayment](p)
}

// DeferredPaymentChoice selects the deferred branch.
func DeferredPaymentChoice(p DeferredPayment) PaymentChoice {
	return either.Right[ImmediatePayment](p)
}

// Event is a sealed domain fact emitted by checkout decisions. The unexported
// marker keeps implementations inside this package.
type Event interface {
	EventName() string
	eventMarker()
}

// OrderCheckedOut records the completed checkout transition.
type OrderCheckedOut struct {
	OrderID    OrderID
	CustomerID CustomerID
	Subtotal   Money
	Discount   Money
	Total      Money
	Payment    PaymentChoice
	Version    OrderVersion
}

// EventName implements Event.
func (*OrderCheckedOut) EventName() string { return "order.checked_out" }

func (*OrderCheckedOut) eventMarker() {}

// DiscountApplied records a non-zero discount selected during checkout.
type DiscountApplied struct {
	OrderID   OrderID
	Amount    Money
	Reason    string
	Promotion optional.Optional[PromotionID]
}

// EventName implements Event.
func (DiscountApplied) EventName() string { return "discount.applied" }

func (DiscountApplied) eventMarker() {}

var requireOrderDraft modeling.Requirement[Order, CustomerID] = func(o Order, _ CustomerID) error {
	if o.Status() != OrderStatusDraft {
		return ErrAlreadyCheckedOut
	}
	return nil
}

var requireOrderOwnedBy modeling.Requirement[Order, CustomerID] = func(o Order, customer CustomerID) error {
	if o.CustomerID() != customer {
		return ErrOrderOwnership
	}
	return nil
}

var requireOrderHasItems modeling.Requirement[Order, CustomerID] = func(o Order, _ CustomerID) error {
	if o.ItemsCount() == 0 {
		return ErrEmptyOrder
	}
	return nil
}

// checkoutRequirements composes the ordered preconditions: draft status,
// ownership match, and at least one line item.
var checkoutRequirements = modeling.AllRequirements(
	requireOrderDraft,
	requireOrderOwnedBy,
	requireOrderHasItems,
)

// DecideInput is the pure context required to decide a checkout.
type DecideInput struct {
	Customer CustomerProfile
	Payment  PaymentChoice
	Coupon   optional.Optional[Coupon]
}

// DiscountInput is the explicit context for discount selection.
type DiscountInput struct {
	Order    Order
	Customer CustomerProfile
	Coupon   optional.Optional[Coupon]
}

// Discount is a pure monetary adjustment with a stable reason code.
type Discount struct {
	Amount Money
	Reason string
}

const (
	// reasonNone marks the zero-discount fallback.
	reasonNone = "none"
	// reasonPremiumLoyalty marks the standing loyalty discount.
	reasonPremiumLoyalty = "premium-loyalty"
	// premiumLoyaltyPercent is the standing premium-loyalty discount.
	premiumLoyaltyPercent uint8 = 5
)

// premiumLoyaltyPromotionID is recorded when the loyalty discount applies.
const premiumLoyaltyPromotionID PromotionID = "premium-loyalty"

func couponApplicable(in *DiscountInput) bool {
	coupon, ok := in.Coupon.Get()
	return ok && ValidCoupon(coupon) == nil
}

func couponDiscount(in *DiscountInput) Discount {
	coupon := in.Coupon.OrZero()
	return Discount{
		Amount: in.Order.Subtotal().Percentage(coupon.Percent),
		Reason: "coupon:" + string(coupon.Code),
	}
}

func premiumLoyaltyDiscount(in *DiscountInput) Discount {
	return Discount{
		Amount: in.Order.Subtotal().Percentage(premiumLoyaltyPercent),
		Reason: reasonPremiumLoyalty,
	}
}

func noDiscount(in *DiscountInput) Discount {
	return Discount{
		Amount: ZeroMoney(in.Order.Subtotal().Currency()),
		Reason: reasonNone,
	}
}

// finalDiscount selects the coupon when valid, otherwise premium loyalty for
// eligible customers, otherwise a zero discount.
func finalDiscount() policy.Policy[*DiscountInput, Discount] {
	couponBranch := policy.When(couponApplicable, couponDiscount)
	loyaltySpec := specification.Contramap(LoyaltyEligibility, func(in *DiscountInput) CustomerProfile {
		return in.Customer
	})
	loyaltyBranch := policy.When(predicate.Predicate[*DiscountInput](loyaltySpec), premiumLoyaltyDiscount)
	return policy.Fallback(policy.FirstApplicable(couponBranch, loyaltyBranch), noDiscount)
}

// carriesLoyaltyPromotion reports whether the discount earns the recorded
// promotion. Zero-amount loyalty discounts record nothing: state that no
// event carries would break the decide/replay law.
func carriesLoyaltyPromotion(d Discount) bool {
	return d.Reason == reasonPremiumLoyalty && !d.Amount.IsZero()
}

// Decide is the pure draft-order checkout transition. The order is never
// mutated; failures return the zero decision and the error unchanged.
func Decide(o Order, input *DecideInput) (modeling.Decision[Order, Event], error) {
	if err := checkoutRequirements(o, input.Customer.ID); err != nil {
		return modeling.Decision[Order, Event]{}, err
	}
	if coupon, ok := input.Coupon.Get(); ok {
		if err := ValidCoupon(coupon); err != nil {
			return modeling.Decision[Order, Event]{}, err
		}
	}
	discount := finalDiscount()(&DiscountInput{Order: o, Customer: input.Customer, Coupon: input.Coupon})
	total, err := result.ToPair(o.Subtotal().Sub(discount.Amount))
	if err != nil {
		return modeling.Decision[Order, Event]{}, err
	}
	changes := []change.Change[Order]{
		markCheckedOutChange(),
		discountChange(discount.Amount),
		totalChange(total),
	}
	if carriesLoyaltyPromotion(discount) {
		changes = append(changes, promotionChange(premiumLoyaltyPromotionID))
	}
	changes = append(changes, versionChange())
	post := change.ApplyAll(o, changes...)
	if err := checkoutInvariants(post); err != nil {
		return modeling.Decision[Order, Event]{}, err
	}
	return modeling.Decide[Event](post).Emit(decideEvents(post, discount, input.Payment)...), nil
}

// decideEvents builds the sealed event list in emission order: the checkout
// fact first, then a discount fact when the amount is non-zero.
func decideEvents(post Order, discount Discount, payment PaymentChoice) []Event {
	events := []Event{
		&OrderCheckedOut{
			OrderID:    post.ID(),
			CustomerID: post.CustomerID(),
			Subtotal:   post.Subtotal(),
			Discount:   post.Discount(),
			Total:      post.Total(),
			Payment:    payment,
			Version:    post.Version(),
		},
	}
	if discount.Amount.IsZero() {
		return events
	}
	promotion := optional.None[PromotionID]()
	if carriesLoyaltyPromotion(discount) {
		promotion = optional.Some(premiumLoyaltyPromotionID)
	}
	return append(events, DiscountApplied{
		OrderID:   post.ID(),
		Amount:    discount.Amount,
		Reason:    discount.Reason,
		Promotion: promotion,
	})
}

// ReduceOrder applies one sealed domain event to order state. Nil and unknown
// values are identity: the Event interface is sealed, so the default branch
// exists only for impossible zero interface values.
func ReduceOrder(state Order, event Event) Order {
	switch e := event.(type) {
	case *OrderCheckedOut:
		return state.ApplyCheckedOutState(e.Subtotal, e.Discount, e.Total, e.Version)
	case DiscountApplied:
		promotion, ok := e.Promotion.Get()
		if !ok {
			return state
		}
		return state.WithPromotion(promotion)
	default:
		return state
	}
}

// ReplayOrder folds events onto initial state in encounter order.
func ReplayOrder(initial Order, events iter.Seq[Event]) Order {
	return reducer.Replay(initial, events, ReduceOrder)
}

// Summary is the pure read model folded from checkout events.
type Summary struct {
	OrderID       OrderID
	Total         Money
	EventCount    int
	DiscountCount int
	reasons       *immutable.Map[string, Money]
}

// Reasons exposes accumulated discount reasons; the zero Summary yields an
// empty map.
func (s Summary) Reasons() immutable.Map[string, Money] {
	if s.reasons == nil {
		return immutable.Map[string, Money]{}
	}
	return *s.reasons
}

// ProjectSummary folds events into a Summary without calling Decide or any
// persistence layer.
func ProjectSummary(events iter.Seq[Event]) Summary {
	return projection.Project(Summary{}, events, projectSummary)
}

func projectSummary(summary Summary, event Event) Summary {
	summary.EventCount++
	switch e := event.(type) {
	case *OrderCheckedOut:
		summary.OrderID = e.OrderID
		summary.Total = e.Total
	case DiscountApplied:
		summary.DiscountCount++
		next := summary.Reasons().Set(e.Reason, e.Amount)
		summary.reasons = &next
	}
	return summary
}
