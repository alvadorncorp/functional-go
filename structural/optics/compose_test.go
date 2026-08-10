package optics_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/optics"
)

type location struct {
	City string
}
type address struct {
	Location location
}
type person struct {
	Name    string
	Address address
}

func personCityLens() optics.Lens[person, string] {
	personAddress := optics.NewLens(
		func(source person) address { return source.Address },
		func(source person, value address) person { source.Address = value; return source },
	)
	addressLocation := optics.NewLens(
		func(source address) location { return source.Location },
		func(source address, value location) address { source.Location = value; return source },
	)
	locationCity := optics.NewLens(
		func(source location) string { return source.City },
		func(source location, value string) location { source.City = value; return source },
	)
	return optics.Compose(optics.Compose(personAddress, addressLocation), locationCity)
}

func TestCompose(t *testing.T) {
	t.Parallel()
	lens := personCityLens()
	source := person{Name: "Ana", Address: address{Location: location{City: "Recife"}}}
	if got := lens.Get(source); got != "Recife" {
		t.Fatalf("Get = %q", got)
	}
	updated := lens.Set(source, "Curitiba")
	if updated.Address.Location.City != "Curitiba" {
		t.Fatalf("Set = %#v", updated)
	}
	if source.Address.Location.City != "Recife" {
		t.Fatalf("source mutated: %#v", source)
	}
	modified := lens.Modify(source, func(city string) string {
		return city + "-PE"
	})
	if modified.Address.Location.City != "Recife-PE" {
		t.Fatalf("Modify = %#v", modified)
	}
}

func TestComposedLensLaws(t *testing.T) {
	t.Parallel()
	lens := personCityLens()
	source := person{Name: "Ana", Address: address{Location: location{City: "Recife"}}}
	if got := lens.Set(source, lens.Get(source)); got != source {
		t.Fatalf("Get-Set: %#v", got)
	}
	if got := lens.Get(lens.Set(source, "Natal")); got != "Natal" {
		t.Fatalf("Set-Get: %q", got)
	}
	got := lens.Set(lens.Set(source, "Natal"), "Belém")
	want := lens.Set(source, "Belém")
	if got != want {
		t.Fatalf("Set-Set: got=%#v want=%#v", got, want)
	}
}
