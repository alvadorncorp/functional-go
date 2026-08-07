package frozen_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/frozen"
)

func TestSliceAccessors(t *testing.T) {
	t.Parallel()

	data := []int{1, 2, 3}

	tests := []struct {
		name      string
		s         frozen.Slice[int]
		wantLen   int
		wantItems []int
	}{
		{
			name:      "empty",
			s:         frozen.NewSlice[int](),
			wantLen:   0,
			wantItems: nil,
		},
		{
			name:      "several",
			s:         frozen.NewSlice(data...),
			wantLen:   len(data),
			wantItems: data,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.s.Len(); got != tc.wantLen {
				t.Errorf("Len() = %d, want %d", got, tc.wantLen)
			}
			if got := tc.s.Items(); !slices.Equal(got, tc.wantItems) {
				t.Errorf("Items() = %v, want %v", got, tc.wantItems)
			}
			if got := slices.Collect(tc.s.Values()); !slices.Equal(got, tc.wantItems) {
				t.Errorf("Values() collected = %v, want %v", got, tc.wantItems)
			}
			for i, want := range tc.wantItems {
				if got := tc.s.At(i); got != want {
					t.Errorf("At(%d) = %v, want %v", i, got, want)
				}
			}
		})
	}
}

func TestSliceAll(t *testing.T) {
	t.Parallel()

	s := frozen.NewSlice("a", "b", "c")

	var gotIdx []int
	var gotVal []string
	for i, v := range s.All() {
		gotIdx = append(gotIdx, i)
		gotVal = append(gotVal, v)
	}

	wantIdx := []int{0, 1, 2}
	wantVal := []string{"a", "b", "c"}
	if !slices.Equal(gotIdx, wantIdx) {
		t.Errorf("All() indices = %v, want %v", gotIdx, wantIdx)
	}
	if !slices.Equal(gotVal, wantVal) {
		t.Errorf("All() values = %v, want %v", gotVal, wantVal)
	}
}

func TestSliceAtPanicsOutOfRange(t *testing.T) {
	t.Parallel()

	s := frozen.NewSlice("a", "b", "c")

	tests := []struct {
		name string
		i    int
	}{
		{name: "negative", i: -1},
		{name: "equal to Len", i: s.Len()},
		{name: "well past Len", i: s.Len() + s.Len() + 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if r != "frozen: index out of range" {
					t.Fatalf("At(%d) panic = %v, want the documented message", tc.i, r)
				}
			}()
			s.At(tc.i)
			t.Fatalf("At(%d) did not panic", tc.i)
		})
	}
}

func TestSliceEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    frozen.Slice[int]
		b    frozen.Slice[int]
		want bool
	}{
		{
			name: "equal",
			a:    frozen.NewSlice(1, 2, 3),
			b:    frozen.NewSlice(1, 2, 3),
			want: true,
		},
		{
			name: "different length",
			a:    frozen.NewSlice(1, 2, 3),
			b:    frozen.NewSlice(1, 2),
			want: false,
		},
		{
			name: "different order",
			a:    frozen.NewSlice(1, 2, 3),
			b:    frozen.NewSlice(3, 2, 1),
			want: false,
		},
		{
			name: "both zero value",
			a:    frozen.Slice[int]{},
			b:    frozen.Slice[int]{},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.a.Equal(tc.b); got != tc.want {
				t.Errorf("Equal() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSliceEqualFunc(t *testing.T) {
	t.Parallel()

	ints := frozen.NewSlice(1, 2, 3)
	strs := frozen.NewSlice("1", "2", "3")
	sameText := func(x int, y string) bool { return strconv.Itoa(x) == y }

	if !frozen.SliceEqualFunc(ints, strs, sameText) {
		t.Error("SliceEqualFunc(ints, strs, sameText) = false, want true")
	}
	if frozen.SliceEqualFunc(ints, frozen.NewSlice("1", "2"), sameText) {
		t.Error("SliceEqualFunc(...) = true for different lengths, want false")
	}
}

func TestSliceCopiesOnConstruction(t *testing.T) {
	t.Parallel()

	xs := []int{1, 2, 3}
	s := frozen.NewSlice(xs...)

	xs[0] = -1

	if got := s.At(0); got != 1 {
		t.Errorf("mutating the source slice leaked into Slice: At(0) = %v, want 1", got)
	}
}

func TestSliceCopiesOnItems(t *testing.T) {
	t.Parallel()

	s := frozen.NewSlice(1, 2, 3)

	first := s.Items()
	first[0] = -1
	second := s.Items()

	if second[0] != 1 {
		t.Errorf("mutating an earlier Items() result leaked into a later call: got %v, want 1", second[0])
	}
}

func TestSliceZeroValue(t *testing.T) {
	t.Parallel()

	var s frozen.Slice[int]

	if got := s.Len(); got != 0 {
		t.Errorf("zero Slice Len() = %d, want 0", got)
	}
	for range s.Values() {
		t.Fatal("zero Slice Values() iterated an element")
	}
	if got := s.Items(); got != nil {
		t.Errorf("zero Slice Items() = %v, want nil", got)
	}
}

func TestCollectSlice(t *testing.T) {
	t.Parallel()

	src := []int{1, 2, 3}
	s := frozen.CollectSlice(slices.Values(src))

	if want := frozen.NewSlice(src...); !s.Equal(want) {
		t.Errorf("CollectSlice(...) = %v, want %v", s.Items(), want.Items())
	}
}
