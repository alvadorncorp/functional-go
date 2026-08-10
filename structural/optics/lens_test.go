package optics_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/optics"
)

type profile struct {
	Name string
	Age  int
}

func nameLens() optics.Lens[profile, string] {
	return optics.NewLens(
		func(source profile) string { return source.Name },
		func(source profile, value string) profile {
			source.Name = value
			return source
		},
	)
}

func TestLensOperationsPreserveSource(t *testing.T) {
	t.Parallel()
	lens := nameLens()
	source := profile{Name: "Ana", Age: 30}
	if got := lens.Get(source); got != "Ana" {
		t.Fatalf("Get = %q", got)
	}
	updated := lens.Set(source, "Bia")
	if updated != (profile{Name: "Bia", Age: 30}) {
		t.Fatalf("Set = %#v", updated)
	}
	if source.Name != "Ana" {
		t.Fatalf("source mutated: %#v", source)
	}
	modified := lens.Modify(source, func(value string) string {
		return value + " Silva"
	})
	if modified.Name != "Ana Silva" || source.Name != "Ana" {
		t.Fatalf("Modify=%#v source=%#v", modified, source)
	}
}

func TestLensLaws(t *testing.T) {
	t.Parallel()
	lens := nameLens()
	source := profile{Name: "Ana", Age: 30}
	if got := lens.Set(source, lens.Get(source)); got != source {
		t.Fatalf("Get-Set: %#v", got)
	}
	if got := lens.Get(lens.Set(source, "Bia")); got != "Bia" {
		t.Fatalf("Set-Get: %q", got)
	}
	left := lens.Set(lens.Set(source, "Bia"), "Cris")
	right := lens.Set(source, "Cris")
	if left != right {
		t.Fatalf("Set-Set: left=%#v right=%#v", left, right)
	}
}

func TestNewLensNilCallbacks(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want any
		call func()
	}{
		{
			"get",
			"functional/optics: nil getter",
			func() {
				optics.NewLens[profile, string](
					nil,
					func(source profile, value string) profile { return source },
				)
			},
		},
		{
			"set",
			"functional/optics: nil setter",
			func() {
				optics.NewLens[profile, string](
					func(source profile) string { return source.Name },
					nil,
				)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if got := recover(); got != test.want {
					t.Fatalf("panic = %#v, want %#v", got, test.want)
				}
			}()
			test.call()
		})
	}
}
