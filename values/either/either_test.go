package either_test

import (
	"testing"

	"go.alvadorn.org/functional/values/either"
)

func TestEitherConstructionAndAccess(t *testing.T) {
	t.Parallel()
	var zero either.Either[int, string]
	if !zero.IsLeft() || zero.IsRight() {
		t.Fatal("zero Either must be Left")
	}
	if got, ok := zero.LeftValue(); !ok || got != 0 {
		t.Fatalf("zero.LeftValue() = (%d, %v)", got, ok)
	}
	if got, ok := zero.RightValue(); ok || got != "" {
		t.Fatalf("zero.RightValue() = (%q, %v)", got, ok)
	}

	left := either.Left[int, string](7)
	if got, ok := left.LeftValue(); !ok || got != 7 {
		t.Fatalf("LeftValue = (%d, %v)", got, ok)
	}
	if got, ok := left.RightValue(); ok || got != "" {
		t.Fatalf("Left.RightValue() = (%q, %v)", got, ok)
	}
	right := either.Right[int, string]("seven")
	if got, ok := right.RightValue(); !ok || got != "seven" {
		t.Fatalf("RightValue = (%q, %v)", got, ok)
	}
	if got, ok := right.LeftValue(); ok || got != 0 {
		t.Fatalf("Right.LeftValue() = (%d, %v)", got, ok)
	}
}

func TestEitherTransformsAndLaziness(t *testing.T) {
	t.Parallel()
	leftCalls, rightCalls := 0, 0
	left := either.Left[int, string](2)

	mappedLeft := either.MapLeft(left, func(v int) string { leftCalls++; return "left-6" })
	if got, ok := mappedLeft.LeftValue(); !ok || got != "left-6" {
		t.Fatalf("MapLeft = (%q, %v)", got, ok)
	}
	mappedRightFromLeft := either.MapRight(left, func(v string) int { rightCalls++; return len(v) })
	if got, ok := mappedRightFromLeft.LeftValue(); !ok || got != 2 {
		t.Fatalf("MapRight(Left).LeftValue() = (%d, %v)", got, ok)
	}
	if got, ok := mappedRightFromLeft.RightValue(); ok || got != 0 {
		t.Fatalf("MapRight(Left).RightValue() = (%d, %v)", got, ok)
	}
	leftBimapCalled, rightBimapCalled := false, false
	biLeft := either.Bimap(
		left,
		func(v int) string { leftCalls++; leftBimapCalled = true; return "left-3" },
		func(v string) int { rightCalls++; rightBimapCalled = true; return len(v) },
	)
	if got, ok := biLeft.LeftValue(); !ok || got != "left-3" {
		t.Fatalf("Bimap(Left) = (%q, %v)", got, ok)
	}
	if !leftBimapCalled || rightBimapCalled {
		t.Fatalf("Bimap(Left) callbacks = (left: %v, right: %v)", leftBimapCalled, rightBimapCalled)
	}
	flatLeft := either.FlatMapRight(left, func(v string) either.Either[int, int] {
		rightCalls++
		return either.Right[int, int](len(v))
	})
	if got, ok := flatLeft.LeftValue(); !ok || got != 2 {
		t.Fatalf("FlatMapRight(Left).LeftValue() = (%d, %v)", got, ok)
	}
	if got, ok := flatLeft.RightValue(); ok || got != 0 {
		t.Fatalf("FlatMapRight(Left).RightValue() = (%d, %v)", got, ok)
	}
	leftMatch := either.Match(
		left,
		func(v int) int { leftCalls++; return v },
		func(v string) int { rightCalls++; return len(v) },
	)
	if leftMatch != 2 {
		t.Fatalf("Match(Left) = %d", leftMatch)
	}
	if leftCalls != 3 || rightCalls != 0 {
		t.Fatalf("Left callbacks=%d Right callbacks=%d", leftCalls, rightCalls)
	}

	right := either.Right[int, string]("go")
	mappedLeftFromRight := either.MapLeft(right, func(v int) int { leftCalls++; return v })
	if got, ok := mappedLeftFromRight.RightValue(); !ok || got != "go" {
		t.Fatalf("MapLeft(Right).RightValue() = (%q, %v)", got, ok)
	}
	if got, ok := mappedLeftFromRight.LeftValue(); ok || got != 0 {
		t.Fatalf("MapLeft(Right).LeftValue() = (%d, %v)", got, ok)
	}
	mappedRight := either.MapRight(right, func(v string) int { rightCalls++; return len(v) })
	if got, ok := mappedRight.RightValue(); !ok || got != 2 {
		t.Fatalf("MapRight = (%d, %v)", got, ok)
	}
	leftBimapCalled, rightBimapCalled = false, false
	biRight := either.Bimap(
		right,
		func(v int) string { leftCalls++; leftBimapCalled = true; return "unused" },
		func(v string) int { rightCalls++; rightBimapCalled = true; return len(v) },
	)
	if got, ok := biRight.RightValue(); !ok || got != 2 {
		t.Fatalf("Bimap(Right) = (%d, %v)", got, ok)
	}
	if leftBimapCalled || !rightBimapCalled {
		t.Fatalf("Bimap(Right) callbacks = (left: %v, right: %v)", leftBimapCalled, rightBimapCalled)
	}
	flat := either.FlatMapRight(right, func(v string) either.Either[int, int] {
		rightCalls++
		return either.Left[int, int](len(v) + 1)
	})
	if got, ok := flat.LeftValue(); !ok || got != 3 {
		t.Fatalf("FlatMapRight callback Left = (%d, %v)", got, ok)
	}
	rightMatch := either.Match(
		right,
		func(v int) int { leftCalls++; return v },
		func(v string) int { rightCalls++; return len(v) },
	)
	if rightMatch != 2 {
		t.Fatalf("Match(Right) = %d", rightMatch)
	}
	if leftCalls != 3 || rightCalls != 4 {
		t.Fatalf("Left callbacks=%d Right callbacks=%d", leftCalls, rightCalls)
	}
}

func TestFlatMapRightRightReturnsCallbackRight(t *testing.T) {
	t.Parallel()

	calls := 0
	got := either.FlatMapRight(either.Right[int, string]("source"), func(value string) either.Either[int, string] {
		calls++
		if value != "source" {
			t.Fatalf("FlatMapRight callback value = %q", value)
		}
		return either.Right[int, string]("mapped")
	})

	if calls != 1 {
		t.Fatalf("FlatMapRight callback calls = %d, want 1", calls)
	}
	if !got.IsRight() || got.IsLeft() {
		t.Fatalf("FlatMapRight result branch: IsRight=%v IsLeft=%v", got.IsRight(), got.IsLeft())
	}
	if value, ok := got.RightValue(); !ok || value != "mapped" {
		t.Fatalf("FlatMapRight result RightValue = (%q, %v)", value, ok)
	}
	if value, ok := got.LeftValue(); ok || value != 0 {
		t.Fatalf("FlatMapRight result LeftValue = (%d, %v)", value, ok)
	}
}
