package specification_test

import (
	"errors"
	"fmt"

	"go.alvadorn.org/functional/modeling"
	"go.alvadorn.org/functional/modeling/policy"
	"go.alvadorn.org/functional/modeling/specification"
)

type customer struct {
	Active        bool
	VerifiedEmail bool
	Debt          bool
	VIP           bool
	Employee      bool
	PromotionCode bool
	Age           int
}

var isActive specification.Specification[customer] = func(input customer) bool {
	return input.Active
}

func ExampleSpecification() {
	fmt.Println(isActive(customer{Active: true}))
	// Output: true
}

func ExampleAll() {
	hasVerifiedEmail := specification.Specification[customer](func(input customer) bool {
		return input.VerifiedEmail
	})
	hasNoOutstandingDebt := specification.Specification[customer](func(input customer) bool {
		return !input.Debt
	})
	eligibleForCredit := specification.All(isActive, hasVerifiedEmail, hasNoOutstandingDebt)

	fmt.Println(eligibleForCredit(customer{Active: true, VerifiedEmail: true}))
	// Output: true
}

func ExampleNot() {
	inactive := specification.Not(isActive)
	fmt.Println(inactive(customer{}))
	// Output: true
}

func ExampleAny() {
	isVIP := specification.Specification[customer](func(input customer) bool { return input.VIP })
	isEmployee := specification.Specification[customer](func(input customer) bool { return input.Employee })
	hasPromotion := specification.Specification[customer](func(input customer) bool {
		return input.PromotionCode
	})
	canReceiveSpecialDiscount := specification.Any(isVIP, isEmployee, hasPromotion)

	fmt.Println(canReceiveSpecialDiscount(customer{Employee: true}))
	// Output: true
}

func ExampleContramap() {
	type money struct{ Cents int64 }
	type order struct{ Total money }
	isHighValueMoney := specification.Specification[money](func(value money) bool {
		return value.Cents >= 100_00
	})
	isHighValueOrder := specification.Contramap(isHighValueMoney, func(input order) money {
		return input.Total
	})

	fmt.Println(isHighValueOrder(order{Total: money{Cents: 120_00}}))
	// Output: true
}

func ExampleSpecification_withPolicy() {
	type pricingContext struct{ Customer customer }
	isVIP := specification.Specification[pricingContext](func(input pricingContext) bool {
		return input.Customer.VIP
	})
	discount := policy.Choose(
		func(input pricingContext) bool { return isVIP(input) },
		policy.Constant[pricingContext](20),
		policy.Constant[pricingContext](0),
	)

	fmt.Println(discount(pricingContext{Customer: customer{VIP: true}}))
	// Output: 20
}

func ExampleSpecification_domainBoundaries() {
	isAdult := specification.Specification[customer](func(input customer) bool { return input.Age >= 18 })
	validAge := modeling.Rule[customer](func(input customer) error {
		if input.Age < 0 {
			return errors.New("age cannot be negative")
		}
		return nil
	})
	type account struct {
		Premium bool
		Balance int
	}
	isPremium := specification.Specification[account](func(input account) bool { return input.Premium })
	canWithdraw := modeling.Requirement[account, int](func(input account, amount int) error {
		if amount > input.Balance {
			return errors.New("insufficient balance")
		}
		return nil
	})
	isOverdrawn := specification.Specification[account](func(input account) bool { return input.Balance < 0 })
	nonNegativeBalance := modeling.Invariant[account](func(input account) error {
		if input.Balance < 0 {
			return errors.New("negative balance")
		}
		return nil
	})

	fmt.Println(isAdult(customer{Age: 20}), validAge(customer{Age: 20}) == nil)
	fmt.Println(isPremium(account{Premium: true}), canWithdraw(account{Balance: 10}, 5) == nil)
	fmt.Println(isOverdrawn(account{Balance: -1}), nonNegativeBalance(account{Balance: -1}) != nil)
	// Output:
	// true true
	// true true
	// true true
}
