package optional_test

import (
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/optional"
)

func TestOptional(t *testing.T) {
	t.Parallel()

	var zero optional.Optional[int]
	if !zero.IsNone() || zero.IsSome() {
		t.Fatal("zero Optional must be None")
	}
	if got, ok := zero.Get(); ok || got != 0 {
		t.Fatalf("zero.Get() = (%d, %v), want (0, false)", got, ok)
	}
	if got := zero.OrElse(7); got != 7 {
		t.Fatalf("zero.OrElse(7) = %d", got)
	}
	if got := zero.OrZero(); got != 0 {
		t.Fatalf("zero.OrZero() = %d", got)
	}

	some := optional.Some(3)
	if !some.IsSome() || some.IsNone() {
		t.Fatal("Some must be present")
	}
	if got, ok := some.Get(); !ok || got != 3 {
		t.Fatalf("Some(3).Get() = (%d, %v)", got, ok)
	}
	if got := some.OrElse(7); got != 3 {
		t.Fatalf("Some(3).OrElse(7) = %d", got)
	}
	if got := some.OrZero(); got != 3 {
		t.Fatalf("Some(3).OrZero() = %d", got)
	}
	if got := some.MustGet(); got != 3 {
		t.Fatalf("Some(3).MustGet() = %d", got)
	}
	if got, ok := optional.From(3, true).Get(); !ok || got != 3 {
		t.Fatalf("From(3, true).Get() = (%d, %v), want (3, true)", got, ok)
	}
	if got, ok := optional.From(3, false).Get(); ok || got != 0 {
		t.Fatalf("From(3, false).Get() = (%d, %v), want (0, false)", got, ok)
	}
	if !optional.Equal(optional.Some(3), some) || optional.Equal(some, optional.None[int]()) {
		t.Fatal("Equal must compare discriminator and value")
	}
	if optional.Equal(optional.Some(3), optional.Some(4)) {
		t.Fatal("Equal(Some(3), Some(4)) must be false")
	}

	someZero := optional.Some(0)
	if !someZero.IsSome() || someZero.IsNone() {
		t.Fatal("Some(0) must be present")
	}
	if got, ok := someZero.Get(); !ok || got != 0 {
		t.Fatalf("Some(0).Get() = (%d, %v), want (0, true)", got, ok)
	}
	if optional.Equal(someZero, optional.None[int]()) {
		t.Fatal("Some(0) must differ from None[int]()")
	}
}

func TestMustGetPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if got := recover(); got != "functional/values/optional: Optional is None" {
			t.Fatalf("panic = %v", got)
		}
	}()
	optional.None[int]().MustGet()
}

func TestOptionalTransformsAreLazy(t *testing.T) {
	t.Parallel()

	calls := 0
	none := optional.None[int]()
	if got := optional.Map(none, func(v int) int { calls++; return v + 1 }); !got.IsNone() {
		t.Fatal("Map(None) must be None")
	}
	flatNone := optional.FlatMap(none, func(v int) optional.Optional[int] {
		calls++
		return optional.Some(v)
	})
	if !flatNone.IsNone() {
		t.Fatal("FlatMap(None) must be None")
	}
	if got := none.Filter(func(int) bool { calls++; return true }); !got.IsNone() {
		t.Fatal("Filter(None) must be None")
	}
	if got := none.Inspect(func(int) { calls++ }); !got.IsNone() {
		t.Fatal("Inspect(None) must preserve None")
	}
	if calls != 0 {
		t.Fatalf("inactive callbacks called %d times", calls)
	}

	some := optional.Some(2)
	if got := optional.Map(some, func(v int) int { calls++; return v * 2 }).MustGet(); got != 4 {
		t.Fatalf("Map(Some(2)) = %d", got)
	}
	mappedString := optional.Map(some, strconv.Itoa)
	if got, ok := mappedString.Get(); !ok || got != "2" {
		t.Fatalf("Map(Some(2), Itoa).Get() = (%q, %v), want (\"2\", true)", got, ok)
	}
	flatSome := optional.FlatMap(some, func(v int) optional.Optional[int] {
		calls++
		return optional.Some(v + 1)
	})
	if got := flatSome.MustGet(); got != 3 {
		t.Fatalf("FlatMap(Some(2)) = %d", got)
	}
	if got := optional.FlatMap(some, func(int) optional.Optional[int] {
		return optional.None[int]()
	}); !got.IsNone() {
		t.Fatal("FlatMap(Some(...), None) must be None")
	}
	flatMappedString := optional.FlatMap(some, func(v int) optional.Optional[string] {
		return optional.Some(strconv.Itoa(v))
	})
	if got, ok := flatMappedString.Get(); !ok || got != "2" {
		t.Fatalf("FlatMap(Some(2), Itoa).Get() = (%q, %v), want (\"2\", true)", got, ok)
	}
	if got := some.Filter(func(v int) bool { calls++; return v == 2 }).MustGet(); got != 2 {
		t.Fatalf("Filter true payload = %d, want 2", got)
	}
	if !some.Filter(func(v int) bool { calls++; return v != 2 }).IsNone() {
		t.Fatal("Filter false must return None")
	}
	if got := some.Inspect(func(v int) { calls += v }).MustGet(); got != 2 {
		t.Fatalf("Inspect payload = %d, want 2", got)
	}
	if calls != 6 {
		t.Fatalf("selected callbacks contributed %d, want 6", calls)
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	var zero optional.Optional[int]
	tests := map[string]struct {
		input                optional.Optional[int]
		want                 string
		noneCalls, someCalls int
	}{
		"Some":          {optional.Some(4), "some:4", 0, 1},
		"Some zero":     {optional.Some(0), "some:0", 0, 1},
		"None":          {optional.None[int](), "none", 1, 0},
		"zero Optional": {zero, "none", 1, 0},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			noneCalls, someCalls := 0, 0
			got := optional.Match(tt.input,
				func() string { noneCalls++; return "none" },
				func(v int) string { someCalls++; return "some:" + strconv.Itoa(v) },
			)
			if got != tt.want || noneCalls != tt.noneCalls || someCalls != tt.someCalls {
				t.Fatalf("Match = %q, noneCalls=%d, someCalls=%d", got, noneCalls, someCalls)
			}
			identity := optional.Match(tt.input, optional.None[int], optional.Some[int])
			if !optional.Equal(identity, tt.input) {
				t.Fatalf("Match(o, None, Some) = %v, want %v", identity, tt.input)
			}
		})
	}
}
