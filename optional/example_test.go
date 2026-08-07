package optional_test

import (
	"fmt"
	"os"
	"strconv"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

func ExampleSome() {
	o := optional.Some(42)
	fmt.Println(o)
	// Output:
	// Some(42)
}

// ExampleOptional_Or shows the left-to-right precedence chain the method
// exists for: a flag value beats an environment value, which beats a
// hard-coded default.
func ExampleOptional_Or() {
	flagValue := optional.None[string]()
	envValue := optional.Some("from-env")
	defaultValue := optional.Some("from-default")

	result := flagValue.Or(envValue).Or(defaultValue)
	fmt.Println(result)
	// Output:
	// Some(from-env)
}

func ExampleMap() {
	o := optional.Some(3)
	doubled := optional.Map(o, func(x int) int { return x * 2 })
	fmt.Println(doubled)
	// Output:
	// Some(6)
}

// ExampleMapErr parses a port number, a fallible conversion, without an
// intermediate error check when the port itself is absent.
func ExampleMapErr() {
	rawPort := optional.Some("8080")
	port, err := optional.MapErr(rawPort, strconv.Atoi)
	fmt.Println(port, err)
	// Output:
	// Some(8080) <nil>
}

func ExampleFromOK() {
	const key = "FUNCTIONAL_OPTIONAL_EXAMPLE_VAR"
	if err := os.Setenv(key, "example-value"); err != nil {
		panic(err)
	}
	defer func() { _ = os.Unsetenv(key) }()

	found := optional.FromOK(os.LookupEnv(key))
	missing := optional.FromOK(os.LookupEnv(key + "_MISSING"))

	fmt.Println(found)
	fmt.Println(missing)
	// Output:
	// Some(example-value)
	// None
}
