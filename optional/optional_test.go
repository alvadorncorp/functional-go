package optional_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/optional"
)

func TestOptionalAccessors(t *testing.T) {
	t.Parallel()

	const heldValue = 7
	const fallbackValue = -1

	tests := []struct {
		name        string
		o           optional.Optional[int]
		wantGet     int
		wantGetOK   bool
		wantOrElse  int
		wantOrZero  int
		wantString  string
		wantIsSome  bool
		wantIsNone  bool
		wantIsZero  bool
		filterKeep  func(int) bool
		wantFilter  optional.Optional[int]
		wantPointer bool
	}{
		{
			name:        "Some",
			o:           optional.Some(heldValue),
			wantGet:     heldValue,
			wantGetOK:   true,
			wantOrElse:  heldValue,
			wantOrZero:  heldValue,
			wantString:  "Some(7)",
			wantIsSome:  true,
			wantIsNone:  false,
			wantIsZero:  false,
			filterKeep:  func(v int) bool { return v == heldValue },
			wantFilter:  optional.Some(heldValue),
			wantPointer: true,
		},
		{
			name:        "None",
			o:           optional.None[int](),
			wantGet:     0,
			wantGetOK:   false,
			wantOrElse:  fallbackValue,
			wantOrZero:  0,
			wantString:  "None",
			wantIsSome:  false,
			wantIsNone:  true,
			wantIsZero:  true,
			filterKeep:  func(int) bool { t.Fatal("keep must not be called on an absent Optional"); return false },
			wantFilter:  optional.None[int](),
			wantPointer: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got, ok := tc.o.Get(); got != tc.wantGet || ok != tc.wantGetOK {
				t.Errorf("Get() = (%v, %v), want (%v, %v)", got, ok, tc.wantGet, tc.wantGetOK)
			}
			if got := tc.o.OrElse(fallbackValue); got != tc.wantOrElse {
				t.Errorf("OrElse(%v) = %v, want %v", fallbackValue, got, tc.wantOrElse)
			}
			if got := tc.o.OrElseGet(func() int { return fallbackValue }); got != tc.wantOrElse {
				t.Errorf("OrElseGet(...) = %v, want %v", got, tc.wantOrElse)
			}
			if got := tc.o.OrZero(); got != tc.wantOrZero {
				t.Errorf("OrZero() = %v, want %v", got, tc.wantOrZero)
			}
			if got := tc.o.String(); got != tc.wantString {
				t.Errorf("String() = %q, want %q", got, tc.wantString)
			}
			if got := tc.o.IsSome(); got != tc.wantIsSome {
				t.Errorf("IsSome() = %v, want %v", got, tc.wantIsSome)
			}
			if got := tc.o.IsNone(); got != tc.wantIsNone {
				t.Errorf("IsNone() = %v, want %v", got, tc.wantIsNone)
			}
			if got := tc.o.IsZero(); got != tc.wantIsZero {
				t.Errorf("IsZero() = %v, want %v", got, tc.wantIsZero)
			}
			if got := tc.o.Filter(tc.filterKeep); !got.Equal(tc.wantFilter) {
				t.Errorf("Filter(...) = %v, want %v", got, tc.wantFilter)
			}
			if p := tc.o.Pointer(); (p != nil) != tc.wantPointer {
				t.Errorf("Pointer() = %v, want non-nil = %v", p, tc.wantPointer)
			} else if p != nil && *p != tc.wantGet {
				t.Errorf("*Pointer() = %v, want %v", *p, tc.wantGet)
			}
		})
	}
}

func TestOptionalOrElseGetSkipsFallbackWhenPresent(t *testing.T) {
	t.Parallel()

	const heldValue = 3

	called := false
	got := optional.Some(heldValue).OrElseGet(func() int {
		called = true
		return -1
	})
	if called {
		t.Fatal("OrElseGet called fallback on a present Optional")
	}
	if got != heldValue {
		t.Fatalf("OrElseGet() = %v, want %v", got, heldValue)
	}
}

func TestOptionalMustGet(t *testing.T) {
	t.Parallel()

	const heldValue = 5
	if got := optional.Some(heldValue).MustGet(); got != heldValue {
		t.Fatalf("MustGet() = %v, want %v", got, heldValue)
	}

	defer func() {
		r := recover()
		if r != "functional/optional: Optional has no value" {
			t.Fatalf("MustGet() panic = %v, want the documented message", r)
		}
	}()
	optional.None[int]().MustGet()
	t.Fatal("MustGet() on None did not panic")
}

func TestOptionalOr(t *testing.T) {
	t.Parallel()

	const someValue = 1
	const otherValue = 2

	tests := []struct {
		name  string
		o     optional.Optional[int]
		other optional.Optional[int]
		want  optional.Optional[int]
	}{
		{
			name:  "present Or present keeps the receiver",
			o:     optional.Some(someValue),
			other: optional.Some(otherValue),
			want:  optional.Some(someValue),
		},
		{
			name:  "absent Or present returns other",
			o:     optional.None[int](),
			other: optional.Some(otherValue),
			want:  optional.Some(otherValue),
		},
		{
			name:  "absent Or absent stays absent",
			o:     optional.None[int](),
			other: optional.None[int](),
			want:  optional.None[int](),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.o.Or(tc.other); !got.Equal(tc.want) {
				t.Errorf("Or() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFromPointer(t *testing.T) {
	t.Parallel()

	const heldValue = 9

	v := heldValue
	got := optional.FromPointer(&v)
	if !got.Equal(optional.Some(heldValue)) {
		t.Errorf("FromPointer(&v) = %v, want Some(%v)", got, heldValue)
	}
	v = -1 // later writes must not reach the Optional
	if got.MustGet() != heldValue {
		t.Errorf("FromPointer copy leaked a later write: got %v, want %v", got.MustGet(), heldValue)
	}

	if got := optional.FromPointer[int](nil); !got.Equal(optional.None[int]()) {
		t.Errorf("FromPointer(nil) = %v, want None", got)
	}
}

func TestFromOK(t *testing.T) {
	t.Parallel()

	const heldValue = "value"

	if got := optional.FromOK(heldValue, true); !got.Equal(optional.Some(heldValue)) {
		t.Errorf("FromOK(%q, true) = %v, want Some(%q)", heldValue, got, heldValue)
	}
	if got := optional.FromOK(heldValue, false); !got.Equal(optional.None[string]()) {
		t.Errorf("FromOK(%q, false) = %v, want None", heldValue, got)
	}
}

func TestSomeNilIsPresent(t *testing.T) {
	t.Parallel()

	var p *int
	so := optional.Some(p)
	if !so.IsSome() {
		t.Error("Some(nil pointer) must be present")
	}
	if so.Equal(optional.None[*int]()) {
		t.Error("Some(nil pointer) must not equal None")
	}
	got, ok := so.Get()
	if !ok || got != nil {
		t.Errorf("Get() = (%v, %v), want (nil, true)", got, ok)
	}
}

func TestSomeNilSliceIsPresent(t *testing.T) {
	t.Parallel()

	var s []int
	so := optional.Some(s)
	if !so.IsSome() {
		t.Error("Some(nil slice) must be present")
	}
	if so.Equal(optional.None[[]int]()) {
		t.Error("Some(nil slice) must not equal None")
	}
	got, ok := so.Get()
	if !ok || got != nil {
		t.Errorf("Get() = (%v, %v), want (nil, true)", got, ok)
	}
}

func TestPointerReturnsACopy(t *testing.T) {
	t.Parallel()

	const heldValue = 4

	o := optional.Some(heldValue)
	p := o.Pointer()
	*p = -1
	if got := o.MustGet(); got != heldValue {
		t.Errorf("mutating Pointer() result leaked into the Optional: got %v, want %v", got, heldValue)
	}
}
