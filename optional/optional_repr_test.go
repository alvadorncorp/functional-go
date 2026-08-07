package optional_test

import (
	"reflect"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

// numFields is the frozen field count of the Optional[T] representation:
// {value T; present bool}. Any change to this number is a breaking change to
// the comparability guarantee documented on the type.
const numFields = 2

// compileTimeComparabilityAssertion never runs: it exists only so the
// compiler rejects the package if Optional[string] ever stops being usable
// as a map key, i.e. stops being comparable.
var compileTimeComparabilityAssertion = map[optional.Optional[string]]int{}

// Comparable()/NumField() below go through reflect.TypeOf(zeroValue) rather
// than the equivalent reflect.TypeFor[T](), which is the documented and
// otherwise-preferred spelling: on this toolchain (go1.26.4),
// reflect.TypeFor on a generic struct instantiated only from an external
// _test package trips a linker dead-code-elimination bug (a spurious
// R_USEIFACE panic referencing the type's eqfunc symbol) at `go test` link
// time. reflect.TypeOf on a concrete zero value is unaffected and reports
// the identical answer, so it is used here as a workaround.

func TestOptionalComparability(t *testing.T) {
	t.Parallel()

	if !reflect.TypeOf(optional.Optional[string]{}).Comparable() {
		t.Error("Optional[string] must be ==-comparable because string is comparable")
	}
	if reflect.TypeOf(optional.Optional[[]byte]{}).Comparable() {
		t.Error("Optional[[]byte] must not be ==-comparable because []byte is not comparable")
	}
	_ = compileTimeComparabilityAssertion
}

func TestOptionalRepresentationIsFrozen(t *testing.T) {
	t.Parallel()

	if got := reflect.TypeOf(optional.Optional[int]{}).NumField(); got != numFields {
		t.Fatalf(
			"Optional[int] has %d fields, want %d; adding a non-comparable field "+
				"would silently break ==-comparability for every instantiation",
			got, numFields,
		)
	}
}

func TestOptionalZeroValueIsNone(t *testing.T) {
	t.Parallel()

	var zero optional.Optional[int]
	if !zero.IsNone() {
		t.Error("the zero Optional[int] must be None")
	}
}

func TestSomeZeroIsNotNone(t *testing.T) {
	t.Parallel()

	if optional.Some(0).Equal(optional.None[int]()) {
		t.Error("Some(0) must not equal None: presence is tracked apart from the value")
	}
}
