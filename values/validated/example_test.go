package validated_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/values/validated"
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
