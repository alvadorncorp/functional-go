package iterator_test

import (
	"iter"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/iterator"
)

func counted(values []int, yielded *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, value := range values {
			(*yielded)++
			if !yield(value) {
				return
			}
		}
	}
}

func TestSourcesAreLazy(t *testing.T) {
	t.Parallel()
	values := []int{1, 2, 3}
	seq := iterator.FromSlice(values)
	values[0] = 9
	if got := slices.Collect(seq); !slices.Equal(got, []int{9, 2, 3}) {
		t.Fatalf("FromSlice = %v", got)
	}
	if got := slices.Collect(iterator.Empty[int]()); got != nil {
		t.Fatalf("Empty = %v", got)
	}
	if got := slices.Collect(iterator.Once("x")); !slices.Equal(got, []string{"x"}) {
		t.Fatalf("Once = %v", got)
	}
	if got := slices.Collect(iterator.FromValues(1, 2)); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("FromValues = %v", got)
	}
}

func TestTransformPipelineIsLazyAndStopsEarly(t *testing.T) {
	t.Parallel()
	yielded, mapped, filtered := 0, 0, 0
	seq := iterator.Map(iterator.Filter(counted([]int{1, 2, 3, 4}, &yielded), func(v int) bool {
		filtered++
		return v%2 == 0
	}), func(v int) int { mapped++; return v * 10 })
	if yielded+mapped+filtered != 0 {
		t.Fatal("pipeline ran during construction")
	}
	got := slices.Collect(iterator.Take(seq, 1))
	if !slices.Equal(got, []int{20}) || yielded != 2 || filtered != 2 || mapped != 1 {
		t.Fatalf("got=%v yielded=%d filtered=%d mapped=%d", got, yielded, filtered, mapped)
	}
}

func TestFilterMapFlatMapAndFlatten(t *testing.T) {
	t.Parallel()
	filtered := iterator.FilterMap(iterator.FromValues(1, 2, 3), func(v int) (string, bool) {
		return string(rune('a' + v - 1)), v%2 == 1
	})
	if got := slices.Collect(filtered); !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("FilterMap = %v", got)
	}
	flat := iterator.FlatMap(iterator.FromValues(1, 2), func(v int) iter.Seq[int] { return iterator.FromValues(v, -v) })
	if got := slices.Collect(flat); !slices.Equal(got, []int{1, -1, 2, -2}) {
		t.Fatalf("FlatMap = %v", got)
	}
	nested := iterator.FromValues(iterator.FromValues(1, 2), iterator.Once(3))
	if got := slices.Collect(iterator.Flatten(nested)); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("Flatten = %v", got)
	}
}

func TestLimits(t *testing.T) {
	t.Parallel()
	started := false
	source := func(yield func(int) bool) { started = true; yield(1) }
	if got := slices.Collect(iterator.Take(source, 0)); got != nil || started {
		t.Fatalf("Take(0) = %v started=%v", got, started)
	}
	if got := slices.Collect(iterator.Drop(source, 0)); !slices.Equal(got, []int{1}) {
		t.Fatalf("Drop(0) = %v", got)
	}
	if got := slices.Collect(iterator.Drop(iterator.FromValues(1, 2, 3), 2)); !slices.Equal(got, []int{3}) {
		t.Fatalf("Drop = %v", got)
	}
	takeWhile := iterator.TakeWhile(iterator.FromValues(1, 2, 3), func(v int) bool { return v < 3 })
	if got := slices.Collect(takeWhile); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("TakeWhile = %v", got)
	}
	dropWhile := iterator.DropWhile(iterator.FromValues(1, 2, 3, 1), func(v int) bool { return v < 3 })
	if got := slices.Collect(dropWhile); !slices.Equal(got, []int{3, 1}) {
		t.Fatalf("DropWhile = %v", got)
	}
}
