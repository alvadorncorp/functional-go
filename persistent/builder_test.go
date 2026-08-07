package persistent_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

const wantBuilderUsedAfterBuildPanic = "persistent: builder used after Build"

// expectPanic calls fn and fails the test unless it panics with exactly
// want.
func expectPanic(t *testing.T, want string, fn func()) {
	t.Helper()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("call did not panic, want panic %q", want)
		}
		if r != want {
			t.Fatalf("panic value = %v, want %q", r, want)
		}
	}()
	fn()
}

func TestListBuilderMethodsPanicAfterBuild(t *testing.T) {
	t.Parallel()

	// newBuilder returns the already-used-up builder together with the List
	// Build() already returned, so a subtest can assert the panicking
	// post-Build call left that List untouched (recover() only swallows the
	// panic; it proves nothing about the collection captured earlier).
	newBuilder := func() (*persistent.ListBuilder[int], persistent.List[int]) {
		b := persistent.NewList(1, 2, 3).Builder()
		built := b.Build()
		return b, built
	}

	assertUnchanged := func(t *testing.T, built persistent.List[int]) {
		t.Helper()
		if got, want := built.Items(), []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("List returned by Build() changed after a post-Build call: got %v, want %v", got, want)
		}
	}

	t.Run("Append", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Append(4) })
		assertUnchanged(t, built)
	})
	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Set(0, 9) })
		assertUnchanged(t, built)
	})
	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Delete(0) })
		assertUnchanged(t, built)
	})
	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Len() })
		assertUnchanged(t, built)
	})
	t.Run("Build again", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Build() })
		assertUnchanged(t, built)
	})
}

func TestMapBuilderMethodsPanicAfterBuild(t *testing.T) {
	t.Parallel()

	// newBuilder returns the already-used-up builder together with the Map
	// Build() already returned, so a subtest can assert the panicking
	// post-Build call left that Map untouched (recover() only swallows the
	// panic; it proves nothing about the collection captured earlier).
	newBuilder := func() (*persistent.MapBuilder[string, int], persistent.Map[string, int]) {
		b := persistent.NewMap(map[string]int{"a": 1}).Builder()
		built := b.Build()
		return b, built
	}

	assertUnchanged := func(t *testing.T, built persistent.Map[string, int]) {
		t.Helper()
		if got, want := built.Items(), map[string]int{"a": 1}; !maps.Equal(got, want) {
			t.Errorf("Map returned by Build() changed after a post-Build call: got %v, want %v", got, want)
		}
	}

	t.Run("Set", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Set("b", 2) })
		assertUnchanged(t, built)
	})
	t.Run("Delete", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Delete("a") })
		assertUnchanged(t, built)
	})
	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Len() })
		assertUnchanged(t, built)
	})
	t.Run("Build again", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Build() })
		assertUnchanged(t, built)
	})
}

func TestSetBuilderMethodsPanicAfterBuild(t *testing.T) {
	t.Parallel()

	// newBuilder returns the already-used-up builder together with the Set
	// Build() already returned, so a subtest can assert the panicking
	// post-Build call left that Set untouched (recover() only swallows the
	// panic; it proves nothing about the collection captured earlier).
	newBuilder := func() (*persistent.SetBuilder[int], persistent.Set[int]) {
		b := persistent.NewSet(1, 2, 3).Builder()
		built := b.Build()
		return b, built
	}

	assertUnchanged := func(t *testing.T, built persistent.Set[int]) {
		t.Helper()
		if got, want := sortedItems(built), []int{1, 2, 3}; !slices.Equal(got, want) {
			t.Errorf("Set returned by Build() changed after a post-Build call: got %v, want %v", got, want)
		}
	}

	t.Run("Add", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Add(4) })
		assertUnchanged(t, built)
	})
	t.Run("Remove", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Remove(1) })
		assertUnchanged(t, built)
	})
	t.Run("Len", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Len() })
		assertUnchanged(t, built)
	})
	t.Run("Build again", func(t *testing.T) {
		t.Parallel()
		b, built := newBuilder()
		expectPanic(t, wantBuilderUsedAfterBuildPanic, func() { b.Build() })
		assertUnchanged(t, built)
	})
}

func TestListBuilderDoesNotAffectSeed(t *testing.T) {
	t.Parallel()

	seed := persistent.NewList(1, 2, 3)
	b := seed.Builder()
	b.Append(4).Set(0, 99).Delete(1)

	if got, want := seed.Items(), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("seed List changed by builder operations: got %v, want %v", got, want)
	}

	built := b.Build()
	if got, want := built.Items(), []int{99, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Build() result = %v, want %v", got, want)
	}
}

func TestMapBuilderDoesNotAffectSeed(t *testing.T) {
	t.Parallel()

	seed := persistent.NewMap(map[string]int{"a": 1, "b": 2})
	b := seed.Builder()
	b.Set("a", 99).Set("c", 3).Delete("b")

	if got, ok := seed.Get("a"); got != 1 || !ok {
		t.Fatalf(`seed Map changed by builder operations: Get("a") = (%d,%v), want (1,true)`, got, ok)
	}
	if seed.Has("c") {
		t.Fatal(`seed Map changed by builder operations: Has("c") = true, want false`)
	}
	if got, ok := seed.Get("b"); got != 2 || !ok {
		t.Fatalf(`seed Map changed by builder operations: Get("b") = (%d,%v), want (2,true)`, got, ok)
	}

	built := b.Build()
	if got, ok := built.Get("a"); got != 99 || !ok {
		t.Fatalf(`Build() Get("a") = (%d,%v), want (99,true)`, got, ok)
	}
	if got, ok := built.Get("c"); got != 3 || !ok {
		t.Fatalf(`Build() Get("c") = (%d,%v), want (3,true)`, got, ok)
	}
	if built.Has("b") {
		t.Fatal(`Build() Has("b") = true, want false`)
	}
}

func TestSetBuilderDoesNotAffectSeed(t *testing.T) {
	t.Parallel()

	seed := persistent.NewSet(1, 2, 3)
	b := seed.Builder()
	b.Add(4).Remove(2)

	if got, want := sortedItems(seed), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("seed Set changed by builder operations: got %v, want %v", got, want)
	}

	built := b.Build()
	if got, want := sortedItems(built), []int{1, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Build() result = %v, want %v", got, want)
	}
}

func TestListBuilderLen(t *testing.T) {
	t.Parallel()

	b := persistent.NewList(1, 2).Builder()
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	b.Append(3)
	if got, want := b.Len(), 3; got != want {
		t.Fatalf("Len() after Append = %d, want %d", got, want)
	}
}

func TestListBuilderMultiAppend(t *testing.T) {
	t.Parallel()

	const n = 200
	seed := persistent.NewList(0, 1, 2)
	b := seed.Builder()
	for i := 3; i < n; i++ {
		b.Append(i)
	}
	b.Set(0, 999)
	b.Delete(1)

	// Seed remains the original three elements.
	if got, want := seed.Items(), []int{0, 1, 2}; !slices.Equal(got, want) {
		t.Fatalf("seed = %v, want %v", got, want)
	}

	built := b.Build()
	if built.Len() != n-1 { // deleted one element
		t.Fatalf("Len = %d, want %d", built.Len(), n-1)
	}
	if built.At(0) != 999 {
		t.Fatalf("At(0)=%d", built.At(0))
	}
	// After Delete(1) on [999,1,2,3,...] → [999,2,3,...]
	if built.At(1) != 2 {
		t.Fatalf("At(1)=%d want 2", built.At(1))
	}
	if built.At(built.Len()-1) != n-1 {
		t.Fatalf("last=%d want %d", built.At(built.Len()-1), n-1)
	}

	// Seed still intact after Build.
	if got, want := seed.Items(), []int{0, 1, 2}; !slices.Equal(got, want) {
		t.Fatalf("seed after Build = %v, want %v", got, want)
	}
}

func TestListBuilderEmptySeedAppend(t *testing.T) {
	t.Parallel()

	b := persistent.NewList[int]().Builder()
	b.Append(10, 20, 30)
	built := b.Build()
	if got, want := built.Items(), []int{10, 20, 30}; !slices.Equal(got, want) {
		t.Fatalf("Build() = %v, want %v", got, want)
	}
}

func TestMapBuilderLen(t *testing.T) {
	t.Parallel()

	b := persistent.NewMap(map[string]int{"a": 1}).Builder()
	if got, want := b.Len(), 1; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	b.Set("b", 2)
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("Len() after Set = %d, want %d", got, want)
	}
}

func TestSetBuilderLen(t *testing.T) {
	t.Parallel()

	b := persistent.NewSet(1, 2).Builder()
	if got, want := b.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}
	b.Add(3)
	if got, want := b.Len(), 3; got != want {
		t.Fatalf("Len() after Add = %d, want %d", got, want)
	}
}
