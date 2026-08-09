package result_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/values/result"
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
