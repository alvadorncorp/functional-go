package plan_test

import (
	"fmt"

	"go.alvadorn.org/functional/workflow/plan"
)

func ExamplePlan() {
	var empty plan.Plan
	kind := empty.Kind()
	fmt.Println(kind == plan.KindEmpty)

	created := plan.Sequence(
		plan.Step("load-customer"),
		plan.Transaction(
			plan.Step("save-order"),
			plan.Step("append-outbox"),
		),
	)
	fmt.Println(created.Kind() == plan.KindSequence)
	for _, child := range created.Children() {
		fmt.Printf("%s %q\n", child.Kind(), child.Name())
	}
	// Output:
	// true
	// true
	// step "load-customer"
	// transaction ""
}

func ExampleSequence_identity() {
	step := plan.Step("decide")
	fmt.Println(plan.Sequence().Kind() == plan.KindEmpty)
	fmt.Println(plan.Sequence(step).Kind() == plan.KindStep)
	fmt.Println(plan.Transaction(step).Kind() == plan.KindTransaction)
	// Output:
	// true
	// true
	// true
}

func ExampleKind_String() {
	for _, kind := range []plan.Kind{
		plan.KindEmpty,
		plan.KindStep,
		plan.KindSequence,
		plan.KindTransaction,
		plan.Kind(255),
	} {
		fmt.Println(kind.String())
	}
	// Output:
	// empty
	// step
	// sequence
	// transaction
	// unknown
}

func ExamplePlan_String() {
	var empty plan.Plan
	fmt.Println(empty.String())
	fmt.Println(plan.Step("a").String())
	fmt.Println(plan.Sequence(
		plan.Step("a"),
		plan.Step("b"),
		plan.Transaction(plan.Step("c"), plan.Step("d")),
	).String())
	fmt.Println(plan.Transaction().String())
	// Output:
	// empty
	// a
	// sequence(a, b, transaction(c, d))
	// transaction()
}
