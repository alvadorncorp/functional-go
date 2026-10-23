package action_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvadorncorp/gopackages/workflow/action"
)

const doubled = answer * 2

func TestMapAppliesTheMapperToTheValue(t *testing.T) {
	t.Parallel()

	mapped := action.Map(action.Pure(answer), func(value int) int { return value * 2 })

	value, err := action.Run(t.Context(), mapped)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != doubled {
		t.Fatalf("Run() value = %d, want %d", value, doubled)
	}
}

func TestMapDoesNotRunDuringConstruction(t *testing.T) {
	t.Parallel()

	mapperCalls := 0
	_ = action.Map(action.Pure(answer), func(value int) int {
		mapperCalls++
		return value
	})

	if mapperCalls != 0 {
		t.Fatalf("mapper calls = %d, want 0", mapperCalls)
	}
}

func TestMapShortCircuitsOnUpstreamFailure(t *testing.T) {
	t.Parallel()

	mapperCalls := 0
	mapped := action.Map(action.Fail[int](errBoom), func(value int) int {
		mapperCalls++
		return value
	})

	_, err := action.Run(t.Context(), mapped)
	if !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want errBoom", err)
	}
	if mapperCalls != 0 {
		t.Fatalf("mapper calls = %d, want 0 after upstream failure", mapperCalls)
	}
}

func TestMapFallibleShortCircuitsAndPreservesItsError(t *testing.T) {
	t.Parallel()

	mapped := action.MapFallible(action.Pure(answer), func(int) (string, error) {
		return "", errBoom
	})

	value, err := action.Run(t.Context(), mapped)
	if !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want errBoom", err)
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
}

func TestMapFallibleDoesNotCallMapperAfterFailure(t *testing.T) {
	t.Parallel()

	mapperCalls := 0
	mapped := action.MapFallible(action.Fail[int](errBoom), func(int) (string, error) {
		mapperCalls++
		return "mapped", nil
	})

	value, err := action.Run(t.Context(), mapped)
	if err != errBoom {
		t.Fatalf("Run() error = %v, want errBoom unchanged", err)
	}
	if value != "" {
		t.Fatalf("Run() value = %q, want the zero value", value)
	}
	if mapperCalls != 0 {
		t.Fatalf("mapper calls = %d, want 0 after upstream failure", mapperCalls)
	}
}

func TestFlatMapSequencesInOrder(t *testing.T) {
	t.Parallel()

	var order []string
	first := action.Do(func(context.Context) (int, error) {
		order = append(order, "first")
		return answer, nil
	})
	chained := action.FlatMap(first, func(value int) action.Action[int] {
		return action.Do(func(context.Context) (int, error) {
			order = append(order, "second")
			return value * 2, nil
		})
	})

	if len(order) != 0 {
		t.Fatalf("order = %v, want empty before Run", order)
	}

	value, err := action.Run(t.Context(), chained)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != doubled {
		t.Fatalf("Run() value = %d, want %d", value, doubled)
	}
	if len(order) != 2 || order[0] != "first" || order[1] != "second" {
		t.Fatalf("order = %v, want [first second]", order)
	}
}

func TestFlatMapDoesNotRunTheBinderAfterFailure(t *testing.T) {
	t.Parallel()

	binderCalls := 0
	chained := action.FlatMap(action.Fail[int](errBoom), func(int) action.Action[int] {
		binderCalls++
		return action.Pure(0)
	})

	if _, err := action.Run(t.Context(), chained); !errors.Is(err, errBoom) {
		t.Fatalf("Run() error = %v, want errBoom", err)
	}
	if binderCalls != 0 {
		t.Fatalf("binder calls = %d, want 0 after upstream failure", binderCalls)
	}
}

func TestFlatMapPropagatesTheRunContext(t *testing.T) {
	t.Parallel()

	chained := action.FlatMap(action.Pure(answer), func(int) action.Action[string] {
		return action.Do(func(ctx context.Context) (string, error) {
			marker, _ := ctx.Value(contextKey{}).(string)
			return marker, nil
		})
	})

	ctx := context.WithValue(t.Context(), contextKey{}, "outer")

	value, err := action.Run(ctx, chained)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}
	if value != "outer" {
		t.Fatalf("Run() value = %q, want %q", value, "outer")
	}
}

func TestTransformCombinatorsPanicOnNilArguments(t *testing.T) {
	t.Parallel()

	assertPanics(t, "workflow/action: nil action", func() {
		_ = action.Map(action.Action[int](nil), func(value int) int { return value })
	})
	assertPanics(t, "workflow/action: nil mapper", func() {
		_ = action.Map[int, int](action.Pure(answer), nil)
	})
	assertPanics(t, "workflow/action: nil mapper", func() {
		_ = action.MapFallible[int, int](action.Pure(answer), nil)
	})
	assertPanics(t, "workflow/action: nil binder", func() {
		_ = action.FlatMap[int, int](action.Pure(answer), nil)
	})
}

func TestFlatMapPanicsWhenBinderReturnsNilAction(t *testing.T) {
	t.Parallel()

	chained := action.FlatMap(action.Pure(answer), func(int) action.Action[int] {
		return nil
	})

	assertPanics(t, "workflow/action: nil action", func() {
		_, _ = action.Run(t.Context(), chained)
	})
}
