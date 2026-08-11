package plan_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/workflow/plan"
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
		fmt.Println(child.Kind(), child.Name())
	}
	// Output:
	// true
	// true
	// 1 load-customer
	// 3
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
