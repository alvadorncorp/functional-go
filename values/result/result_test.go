package result_test

import (
	"errors"
	"strconv"
	"testing"

	"go.alvadorn.org/functional/values/result"
)

func TestResult(t *testing.T) {
	t.Parallel()
	var zero result.Result[int, string]
	if !zero.IsErr() || zero.IsOk() {
		t.Fatal("zero Result must be Err")
	}
	if got, ok := zero.Get(); ok || got != 0 {
		t.Fatalf("zero.Get() = (%d, %v)", got, ok)
	}
	if got, ok := zero.ErrValue(); !ok || got != "" {
		t.Fatalf("zero.ErrValue() = (%q, %v)", got, ok)
	}

	okResult := result.Ok[int, string](4)
	if got, ok := okResult.Get(); !ok || got != 4 {
		t.Fatalf("Ok.Get() = (%d, %v)", got, ok)
	}
	if got, ok := okResult.ErrValue(); ok || got != "" {
		t.Fatalf("Ok.ErrValue() = (%q, %v)", got, ok)
	}
	if okResult.OrElse(9) != 4 || okResult.OrZero() != 4 || okResult.MustGet() != 4 {
		t.Fatal("Ok extraction mismatch")
	}

	errResult := result.Err[int]("bad")
	if got, ok := errResult.ErrValue(); !ok || got != "bad" {
		t.Fatalf("Err.ErrValue() = (%q, %v)", got, ok)
	}
	if errResult.OrElse(9) != 9 || errResult.OrZero() != 0 {
		t.Fatal("Err fallback mismatch")
	}
	if !result.Equal(errResult, result.Err[int]("bad")) || result.Equal(errResult, result.Err[int]("other")) {
		t.Fatal("Equal mismatch")
	}
}

func TestResultMustGetPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if got := recover(); got != "functional/values/result: Result is Err" {
			t.Fatalf("panic = %v", got)
		}
	}()
	result.Err[int]("bad").MustGet()
}

func TestResultTransformsAreLazy(t *testing.T) {
	t.Parallel()
	calls := 0
	errResult := result.Err[int]("bad")
	if !result.Map(errResult, func(v int) int { calls++; return v + 1 }).IsErr() {
		t.Fatal("Map(Err) must preserve Err")
	}
	flatErr := result.FlatMap(errResult, func(v int) result.Result[int, string] {
		calls++
		return result.Ok[int, string](v)
	})
	if !flatErr.IsErr() {
		t.Fatal("FlatMap(Err) must preserve Err")
	}
	if calls != 0 {
		t.Fatalf("success callbacks called %d times", calls)
	}

	mappedErr := result.MapErr(errResult, func(e string) int { calls++; return len(e) })
	if got, ok := mappedErr.ErrValue(); !ok || got != 3 || calls != 1 {
		t.Fatalf("MapErr = (%d, %v), calls=%d", got, ok, calls)
	}

	okResult := result.Ok[int, string](2)
	if got := result.Map(okResult, func(v int) int { calls++; return v * 2 }).MustGet(); got != 4 {
		t.Fatalf("Map(Ok) = %d", got)
	}
	flatOK := result.FlatMap(okResult, func(v int) result.Result[int, string] {
		calls++
		return result.Ok[int, string](v + 1)
	})
	if got := flatOK.MustGet(); got != 3 {
		t.Fatalf("FlatMap(Ok) = %d", got)
	}
	if got := result.MapErr(okResult, func(string) int { calls++; return 0 }); !got.IsOk() {
		t.Fatal("MapErr(Ok) must preserve Ok")
	}
	if calls != 3 {
		t.Fatalf("callbacks called %d times, want 3", calls)
	}
}

type myErr struct{}

func (*myErr) Error() string { return "my error" }

// concreteNilErr returns a nil *myErr, which FromPair receives as a non-nil error.
func concreteNilErr() (int, *myErr) { return 1, nil }

func TestFromPairToPair(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("sentinel")
	var typedNil *myErr
	tests := map[string]struct {
		got  result.Result[int, error]
		want result.Result[int, error]
	}{
		"value with nil error":             {result.FromPair(7, nil), result.Ok[int, error](7)},
		"zero value with nil error":        {result.FromPair(0, nil), result.Ok[int, error](0)},
		"error drops value":                {result.FromPair(7, sentinel), result.Err[int](sentinel)},
		"typed nil error interface":        {result.FromPair(7, error(typedNil)), result.Err[int](error(typedNil))},
		"concrete pointer error type trap": {result.FromPair(concreteNilErr()), result.Err[int](error(typedNil))},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !result.Equal(tt.got, tt.want) {
				t.Fatalf("FromPair = %v, want %v", tt.got, tt.want)
			}
		})
	}

	if v, err := result.ToPair(result.FromPair(7, nil)); v != 7 || err != nil {
		t.Fatalf("ToPair(FromPair(7, nil)) = (%d, %v)", v, err)
	}
	if v, err := result.ToPair(result.FromPair(7, sentinel)); v != 0 || err != sentinel {
		t.Fatalf("ToPair(FromPair(7, sentinel)) = (%d, %v)", v, err)
	}
	for _, r := range []result.Result[int, error]{
		result.Ok[int, error](7), result.Ok[int, error](0), result.Err[int](sentinel),
	} {
		if got := result.FromPair(result.ToPair(r)); !result.Equal(got, r) {
			t.Fatalf("FromPair(ToPair(%v)) = %v", r, got)
		}
	}
}

func TestToPairPanicsOnNilErr(t *testing.T) {
	t.Parallel()
	var zero result.Result[int, error]
	tests := map[string]result.Result[int, error]{
		"zero Result":        zero,
		"Err with nil error": result.Err[int](error(nil)),
	}
	for name, r := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if got := recover(); got != "functional/values/result: ToPair on Err with nil error" {
					t.Fatalf("panic = %v", got)
				}
			}()
			_, _ = result.ToPair(r)
		})
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	var zero result.Result[int, string]
	tests := map[string]struct {
		input             result.Result[int, string]
		want              string
		errCalls, okCalls int
	}{
		"Ok":          {result.Ok[int, string](4), "ok:4", 0, 1},
		"Err":         {result.Err[int]("bad"), "err:bad", 1, 0},
		"zero Result": {zero, "err:", 1, 0},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			errCalls, okCalls := 0, 0
			got := result.Match(tt.input,
				func(e string) string { errCalls++; return "err:" + e },
				func(v int) string { okCalls++; return "ok:" + strconv.Itoa(v) },
			)
			if got != tt.want || errCalls != tt.errCalls || okCalls != tt.okCalls {
				t.Fatalf("Match = %q, errCalls=%d, okCalls=%d", got, errCalls, okCalls)
			}
			identity := result.Match(tt.input, result.Err[int, string], result.Ok[int, string])
			if !result.Equal(identity, tt.input) {
				t.Fatalf("Match(r, Err, Ok) = %v, want %v", identity, tt.input)
			}
		})
	}
}
