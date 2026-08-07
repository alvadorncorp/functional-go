package persistent_test

import (
	"reflect"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

// Comparable()/TypeOf below go through reflect.TypeOf(zeroValue) rather
// than the equivalent reflect.TypeFor[T](), which is the documented and
// otherwise-preferred spelling: on this toolchain (go1.26.4),
// reflect.TypeFor on a generic struct instantiated only from an external
// _test package trips a linker dead-code-elimination bug (a spurious
// R_USEIFACE panic referencing the type's eqfunc symbol) at `go test` link
// time. reflect.TypeOf on a concrete zero value is unaffected and reports
// the identical answer, so it is used here as a workaround.

func TestListIsNotComparable(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(persistent.List[int]{}).Comparable() {
		t.Error("List[int] must not be ==-comparable: identity of the internal representation must never be observable")
	}
}

func TestMapIsNotComparable(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(persistent.Map[string, int]{}).Comparable() {
		t.Error("Map[string, int] must not be ==-comparable: " +
			"identity of the internal representation must never be observable")
	}
}

func TestSetIsNotComparable(t *testing.T) {
	t.Parallel()

	if reflect.TypeOf(persistent.Set[int]{}).Comparable() {
		t.Error("Set[int] must not be ==-comparable: identity of the internal representation must never be observable")
	}
}
