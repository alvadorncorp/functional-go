package validated_test

import (
	"fmt"

	"go.alvadorn.org/functional/values/nonempty"
	"go.alvadorn.org/functional/values/validated"
)

func ExampleMap2() {
	name := validated.Invalid[string]("name is required")
	age := validated.Invalid[int]("age must be positive")
	person := validated.Map2(name, age, func(name string, age int) string {
		return fmt.Sprintf("%s:%d", name, age)
	})
	errors, _ := person.Errors()
	fmt.Println(errors.Items())
	// Output:
	// [name is required age must be positive]
}

func ExampleMatch() {
	report := func(age validated.Validated[int, string]) string {
		return validated.Match(
			age,
			func(errors nonempty.NonEmpty[string]) string {
				return fmt.Sprintf("%d errors: %v", errors.Len(), errors.Items())
			},
			func(age int) string { return fmt.Sprintf("age %d", age) },
		)
	}
	fmt.Println(report(validated.Valid[int, string](42)))
	fmt.Println(report(validated.Invalid[int]("age is required", "age must be a number")))
	// Output:
	// age 42
	// 2 errors: [age is required age must be a number]
}
