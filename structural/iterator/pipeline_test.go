package iterator_test

import (
	"iter"
	"slices"
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/iterator"
	"github.com/alvadorncorp/gopackages/functional/structural/tuple"
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

func TestEmptyAndOnce(t *testing.T) {
	t.Parallel()
	if got := slices.Collect(iterator.Empty[int]()); got != nil {
		t.Fatalf("Empty = %v", got)
	}
	if got := slices.Collect(iterator.Once("x")); !slices.Equal(got, []string{"x"}) {
		t.Fatalf("Once = %v", got)
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
	filtered := iterator.FilterMap(slices.Values([]int{1, 2, 3}), func(v int) (string, bool) {
		return string(rune('a' + v - 1)), v%2 == 1
	})
	if got := slices.Collect(filtered); !slices.Equal(got, []string{"a", "c"}) {
		t.Fatalf("FilterMap = %v", got)
	}
	flat := iterator.FlatMap(slices.Values([]int{1, 2}), func(v int) iter.Seq[int] { return slices.Values([]int{v, -v}) })
	if got := slices.Collect(flat); !slices.Equal(got, []int{1, -1, 2, -2}) {
		t.Fatalf("FlatMap = %v", got)
	}
	nested := slices.Values([]iter.Seq[int]{slices.Values([]int{1, 2}), iterator.Once(3)})
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
	if got := slices.Collect(iterator.Drop(slices.Values([]int{1, 2, 3}), 2)); !slices.Equal(got, []int{3}) {
		t.Fatalf("Drop = %v", got)
	}
	takeWhile := iterator.TakeWhile(slices.Values([]int{1, 2, 3}), func(v int) bool { return v < 3 })
	if got := slices.Collect(takeWhile); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("TakeWhile = %v", got)
	}
	dropWhile := iterator.DropWhile(slices.Values([]int{1, 2, 3, 1}), func(v int) bool { return v < 3 })
	if got := slices.Collect(dropWhile); !slices.Equal(got, []int{3, 1}) {
		t.Fatalf("DropWhile = %v", got)
	}
}

func assertConsumerBreak[T comparable](t *testing.T, seq iter.Seq[T], want []T) {
	t.Helper()
	var got []T
	for value := range seq {
		got = append(got, value)
		if len(got) == len(want) {
			break
		}
	}
	if !slices.Equal(got, want) {
		t.Fatalf("consumer break = %v, want %v", got, want)
	}
}

func countingNumbers(infinite bool, pulls *int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for value := 1; infinite || value <= 8; value++ {
			(*pulls)++
			if !yield(value) {
				return
			}
		}
	}
}

func TestAdaptersConsumerBreak(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		adapt func(iter.Seq[int]) iter.Seq[int]
		want  []int
		pulls []int
	}{
		{
			name: "Map",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.Map(seq, func(v int) int { return v * 10 })
			},
			want: []int{10, 20}, pulls: []int{1, 2},
		},
		{
			name: "Filter",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.Filter(seq, func(v int) bool { return v%2 == 0 })
			},
			want: []int{2, 4}, pulls: []int{2, 4},
		},
		{
			name: "FilterMap",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.FilterMap(seq, func(v int) (int, bool) { return v * 10, v%2 == 0 })
			},
			want: []int{20, 40}, pulls: []int{2, 4},
		},
		{
			name: "Take",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.Take(seq, 4)
			},
			want: []int{1, 2}, pulls: []int{1, 2},
		},
		{
			name: "Drop",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.Drop(seq, 2)
			},
			want: []int{3, 4}, pulls: []int{3, 4},
		},
		{
			name: "TakeWhile",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.TakeWhile(seq, func(v int) bool { return v < 5 })
			},
			want: []int{1, 2}, pulls: []int{1, 2},
		},
		{
			name: "DropWhile",
			adapt: func(seq iter.Seq[int]) iter.Seq[int] {
				return iterator.DropWhile(seq, func(v int) bool { return v < 3 })
			},
			want: []int{3, 4}, pulls: []int{3, 4},
		},
	}
	for _, infinite := range []bool{false, true} {
		for _, k := range []int{1, 2} {
			name := "infinite=" + strconv.FormatBool(infinite) + "/k=" + strconv.Itoa(k)
			t.Run(name, func(t *testing.T) {
				for _, test := range tests {
					t.Run(test.name, func(t *testing.T) {
						pulls := 0
						seq := test.adapt(countingNumbers(infinite, &pulls))
						assertConsumerBreak(t, seq, test.want[:k])
						if pulls != test.pulls[k-1] {
							t.Fatalf("source pulls = %d, want %d", pulls, test.pulls[k-1])
						}
					})
				}
				t.Run("Zip", func(t *testing.T) {
					left, right := 0, 0
					seq := iterator.Zip(countingNumbers(infinite, &left), countingNumbers(infinite, &right))
					want := []tuple.Tuple2[int, int]{tuple.Of2(1, 1), tuple.Of2(2, 2)}
					assertConsumerBreak(t, seq, want[:k])
					if left != k || right != k {
						t.Fatalf("source pulls = %d/%d, want %d/%d", left, right, k, k)
					}
				})
				t.Run("Enumerate", func(t *testing.T) {
					pulls := 0
					seq := iterator.Enumerate(countingNumbers(infinite, &pulls))
					want := []tuple.Tuple2[int, int]{tuple.Of2(0, 1), tuple.Of2(1, 2)}
					assertConsumerBreak(t, seq, want[:k])
					if pulls != k {
						t.Fatalf("source pulls = %d, want %d", pulls, k)
					}
				})
			})
		}
	}
}

func TestNestedAdaptersConsumerBreak(t *testing.T) {
	t.Parallel()
	for _, outerInfinite := range []bool{false, true} {
		for _, innerInfinite := range []bool{false, true} {
			for _, flatten := range []bool{false, true} {
				for _, k := range []int{1, 2} {
					name := "FlatMap"
					if flatten {
						name = "Flatten"
					}
					name += "/outerInfinite=" + strconv.FormatBool(outerInfinite)
					name += "/innerInfinite=" + strconv.FormatBool(innerInfinite) + "/k=" + strconv.Itoa(k)
					t.Run(name, func(t *testing.T) {
						outerPulls, innerPulls := 0, 0
						outer := countingNumbers(outerInfinite, &outerPulls)
						inner := func(v int) iter.Seq[int] {
							return func(yield func(int) bool) {
								// The first empty inner forces traversal to the next outer.
								if v == 1 {
									return
								}
								for offset := 0; innerInfinite || offset < 1; offset++ {
									innerPulls++
									if !yield(v*10 + offset) {
										return
									}
								}
							}
						}
						var seq iter.Seq[int]
						if flatten {
							nested := func(yield func(iter.Seq[int]) bool) {
								for value := range outer {
									if !yield(inner(value)) {
										return
									}
								}
							}
							seq = iterator.Flatten(nested)
						} else {
							seq = iterator.FlatMap(outer, inner)
						}
						want, wantOuter := []int{20, 30}, k+1
						if innerInfinite {
							want, wantOuter = []int{20, 21}, 2
						}
						assertConsumerBreak(t, seq, want[:k])
						if outerPulls != wantOuter || innerPulls != k {
							t.Fatalf("outer/inner pulls = %d/%d, want %d/%d", outerPulls, innerPulls, wantOuter, k)
						}
					})
				}
			}
		}
	}
}

func TestGeneratorsConsumerBreak(t *testing.T) {
	t.Parallel()
	for _, infinite := range []bool{false, true} {
		for _, k := range []int{1, 2} {
			name := "infinite=" + strconv.FormatBool(infinite) + "/k=" + strconv.Itoa(k)
			t.Run(name, func(t *testing.T) {
				t.Run("Iterate", func(t *testing.T) {
					nextCalls := 0
					seq := iterator.Iterate(3, func(v int) int {
						nextCalls++
						return v * 2
					})
					if !infinite {
						seq = iterator.Take(seq, 8)
					}
					assertConsumerBreak(t, seq, []int{3, 6}[:k])
					if nextCalls != k-1 {
						t.Fatalf("next calls = %d, want %d", nextCalls, k-1)
					}
				})
				t.Run("Unfold", func(t *testing.T) {
					stepCalls := 0
					seq := iterator.Unfold(1, func(state int) (int, int, bool) {
						stepCalls++
						return state * 10, state + 1, infinite || state <= 8
					})
					assertConsumerBreak(t, seq, []int{10, 20}[:k])
					if stepCalls != k {
						t.Fatalf("step calls = %d, want %d", stepCalls, k)
					}
				})
				t.Run("Repeat", func(t *testing.T) {
					seq := iterator.Repeat("repeat")
					if !infinite {
						seq = iterator.Take(seq, 8)
					}
					// Repeat has no callback to probe; count only delivered source values.
					delivered := 0
					observed := func(yield func(string) bool) {
						for value := range seq {
							delivered++
							if !yield(value) {
								return
							}
						}
					}
					assertConsumerBreak(t, observed, []string{"repeat", "repeat"}[:k])
					if delivered != k {
						t.Fatalf("delivered values = %d, want %d", delivered, k)
					}
				})
			})
		}
	}
}
