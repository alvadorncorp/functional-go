package validated_test

import (
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/nonempty"
	"github.com/alvadorncorp/gopackages/functional/values/validated"
)

func TestValidatedConstructionAndAccess(t *testing.T) {
	t.Parallel()
	var zero validated.Validated[int, string]
	if !zero.IsInvalid() || zero.IsValid() {
		t.Fatal("zero Validated must be Invalid")
	}
	if got, ok := zero.Get(); ok || got != 0 {
		t.Fatalf("zero.Get() = (%d, %v)", got, ok)
	}
	if errors, ok := zero.Errors(); !ok || errors.Len() != 1 || errors.Head() != "" {
		t.Fatalf("zero.Errors() = %v, %v", errors.Items(), ok)
	}

	valid := validated.Valid[int, string](7)
	if got, ok := valid.Get(); !ok || got != 7 {
		t.Fatalf("Valid.Get() = (%d, %v)", got, ok)
	}
	if _, ok := valid.Errors(); ok {
		t.Fatal("Valid.Errors() must return false")
	}

	invalid := validated.Invalid[int]("name", "email")
	if errors, ok := invalid.Errors(); !ok || len(errors.Items()) != 2 {
		t.Fatalf("Invalid.Errors() = %v, %v", errors.Items(), ok)
	}
	fromNE := validated.InvalidNonEmpty[int](nonempty.Of("age"))
	if errors, _ := fromNE.Errors(); errors.Head() != "age" {
		t.Fatalf("InvalidNonEmpty = %v", errors.Items())
	}
}

func TestValidatedMapAndMapErrorsAreLazy(t *testing.T) {
	t.Parallel()
	valueCalls, errorCalls := 0, 0
	invalid := validated.Invalid[int]("first", "second")
	mappedInvalid := validated.Map(invalid, func(v int) int { valueCalls++; return v + 1 })
	if !mappedInvalid.IsInvalid() {
		t.Fatal("Map(Invalid) must preserve Invalid")
	}
	if errors, ok := mappedInvalid.Errors(); !ok || !equalStrings(errors.Items(), []string{"first", "second"}) {
		t.Fatalf("Map(Invalid).Errors() = (%v, %v)", errors.Items(), ok)
	}
	mappedErrors := validated.MapErrors(invalid, func(e string) int { errorCalls++; return len(e) })
	if errors, ok := mappedErrors.Errors(); !ok || !equalInts(errors.Items(), []int{5, 6}) {
		t.Fatalf("MapErrors = %v", errors.Items())
	}
	valid := validated.Valid[int, string](2)
	if got, _ := validated.Map(valid, func(v int) int { valueCalls++; return v * 2 }).Get(); got != 4 {
		t.Fatalf("Map(Valid) = %d", got)
	}
	if got, ok := validated.MapErrors(valid, func(string) int { errorCalls++; return 0 }).Get(); !ok || got != 2 {
		t.Fatalf("MapErrors(Valid).Get() = (%d, %v)", got, ok)
	}
	if valueCalls != 1 || errorCalls != 2 {
		t.Fatalf("value calls=%d error calls=%d", valueCalls, errorCalls)
	}
}

func TestMap2Cases(t *testing.T) {
	t.Parallel()
	validA := validated.Valid[int, string](2)
	validB := validated.Valid[int, string](3)
	invalidA := validated.Invalid[int]("a1", "a2")
	invalidB := validated.Invalid[int]("b1", "b2")

	for _, test := range []struct {
		name       string
		a          validated.Validated[int, string]
		b          validated.Validated[int, string]
		wantValid  bool
		wantValue  int
		wantErrors []string
		wantCalls  int
	}{
		{name: "valid plus valid", a: validA, b: validB, wantValid: true, wantValue: 5, wantCalls: 1},
		{name: "invalid plus valid", a: invalidA, b: validB, wantErrors: []string{"a1", "a2"}},
		{name: "valid plus invalid", a: validA, b: invalidB, wantErrors: []string{"b1", "b2"}},
		{name: "invalid plus invalid", a: invalidA, b: invalidB, wantErrors: []string{"a1", "a2", "b1", "b2"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			result := validated.Map2(test.a, test.b, func(a, b int) int { calls++; return a + b })
			if got, ok := result.Get(); ok != test.wantValid || got != test.wantValue {
				t.Fatalf("Get() = (%d, %v), want (%d, %v)", got, ok, test.wantValue, test.wantValid)
			}
			if errors, ok := result.Errors(); ok != !test.wantValid || (ok && !equalStrings(errors.Items(), test.wantErrors)) {
				t.Fatalf("Errors() = (%v, %v), want (%v, %v)", errors.Items(), ok, test.wantErrors, !test.wantValid)
			}
			if calls != test.wantCalls {
				t.Fatalf("combine calls = %d, want %d", calls, test.wantCalls)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	var zero validated.Validated[int, string]
	tests := map[string]struct {
		input                    validated.Validated[int, string]
		want                     string
		wantErrors               []string
		invalidCalls, validCalls int
	}{
		"Valid": {input: validated.Valid[int, string](4), want: "valid:4", validCalls: 1},
		"Invalid": {
			input: validated.Invalid[int]("a", "b"), want: "invalid", wantErrors: []string{"a", "b"}, invalidCalls: 1,
		},
		"zero Validated": {input: zero, want: "invalid", wantErrors: []string{""}, invalidCalls: 1},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			invalidCalls, validCalls := 0, 0
			var gotErrors []string
			got := validated.Match(tt.input,
				func(errors nonempty.NonEmpty[string]) string {
					invalidCalls++
					gotErrors = errors.Items()
					return "invalid"
				},
				func(v int) string { validCalls++; return "valid:" + strconv.Itoa(v) },
			)
			if got != tt.want || invalidCalls != tt.invalidCalls || validCalls != tt.validCalls {
				t.Fatalf("Match = %q, invalidCalls=%d, validCalls=%d", got, invalidCalls, validCalls)
			}
			if !equalStrings(gotErrors, tt.wantErrors) {
				t.Fatalf("invalidFn errors = %q, want %q", gotErrors, tt.wantErrors)
			}

			identity := validated.Match(tt.input, validated.InvalidNonEmpty[int, string], validated.Valid[int, string])
			gotValue, gotValid := identity.Get()
			wantValue, wantValid := tt.input.Get()
			gotErrs, _ := identity.Errors()
			wantErrs, _ := tt.input.Errors()
			if gotValue != wantValue || gotValid != wantValid || !equalStrings(gotErrs.Items(), wantErrs.Items()) {
				t.Fatalf("Match(v, InvalidNonEmpty, Valid) = (%d, %v, %q), want (%d, %v, %q)",
					gotValue, gotValid, gotErrs.Items(), wantValue, wantValid, wantErrs.Items())
			}
		})
	}
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func equalInts(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
