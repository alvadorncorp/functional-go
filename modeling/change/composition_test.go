package change_test

import (
	"testing"

	"go.alvadorn.org/functional/modeling/change"
)

type compositionBuilder[S any] func(...change.Change[S]) change.Change[S]

var (
	_ compositionBuilder[int] = change.Compose[int]
	_ compositionBuilder[int] = change.Pipe[int]
)

func addOne(value int) int {
	return value + 1
}

func double(value int) int {
	return value * 2
}

func subtractThree(value int) int {
	return value - 3
}

func TestPipeAppliesChangesLeftToRight(t *testing.T) {
	if got, want := change.Pipe[int](addOne, double)(3), 8; got != want {
		t.Fatalf("Pipe(addOne, double)(3) = %d, want %d", got, want)
	}
}

func TestComposeAppliesChangesRightToLeft(t *testing.T) {
	if got, want := change.Compose[int](addOne, double)(3), 7; got != want {
		t.Fatalf("Compose(addOne, double)(3) = %d, want %d", got, want)
	}
}

func TestEmptyCompositionIsIdentity(t *testing.T) {
	for _, value := range []int{-2, 0, 4} {
		if got := change.Pipe[int]()(value); got != value {
			t.Fatalf("Pipe()(%d) = %d, want %d", value, got, value)
		}
		if got := change.Compose[int]()(value); got != value {
			t.Fatalf("Compose()(%d) = %d, want %d", value, got, value)
		}
	}
}

func TestCompositionIdentityLaws(t *testing.T) {
	identity := change.Identity[int]()
	for _, value := range []int{-2, 0, 4} {
		for name, composed := range map[string]change.Change[int]{
			"pipe left":     change.Pipe(identity, change.Change[int](addOne)),
			"pipe right":    change.Pipe(change.Change[int](addOne), identity),
			"compose left":  change.Compose(identity, change.Change[int](addOne)),
			"compose right": change.Compose(change.Change[int](addOne), identity),
		} {
			if got, want := composed(value), addOne(value); got != want {
				t.Fatalf("%s at %d = %d, want %d", name, value, got, want)
			}
		}
	}
}

func TestCompositionAssociativityLaws(t *testing.T) {
	for _, value := range []int{-2, 0, 4} {
		pipeLeft := change.Pipe(change.Pipe[int](addOne, double), subtractThree)
		pipeRight := change.Pipe(change.Change[int](addOne), change.Pipe[int](double, subtractThree))
		if got, want := pipeLeft(value), pipeRight(value); got != want {
			t.Fatalf("Pipe associativity at %d: %d != %d", value, got, want)
		}

		composeLeft := change.Compose(change.Compose[int](addOne, double), subtractThree)
		composeRight := change.Compose(change.Change[int](addOne), change.Compose[int](double, subtractThree))
		if got, want := composeLeft(value), composeRight(value); got != want {
			t.Fatalf("Compose associativity at %d: %d != %d", value, got, want)
		}
	}
}
