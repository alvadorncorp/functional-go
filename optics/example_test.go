package optics_test

import (
	"fmt"

	"github.com/alvadorncorp/gopackages/functional/optics"
)

type exampleAddress struct{ City string }

type examplePerson struct {
	Name    string
	Address exampleAddress
}

func ExampleCompose() {
	personAddress := optics.NewLens(
		func(source examplePerson) exampleAddress { return source.Address },
		func(source examplePerson, value exampleAddress) examplePerson {
			source.Address = value
			return source
		},
	)
	addressCity := optics.NewLens(
		func(source exampleAddress) string { return source.City },
		func(source exampleAddress, value string) exampleAddress {
			source.City = value
			return source
		},
	)
	personCity := optics.Compose(personAddress, addressCity)
	person := examplePerson{Name: "Ana", Address: exampleAddress{City: "Curitiba"}}
	updated := personCity.Set(person, "São Paulo")
	fmt.Println(person.Address.City)
	fmt.Println(updated.Address.City)
	// Output:
	// Curitiba
	// São Paulo
}
