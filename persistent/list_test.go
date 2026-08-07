package persistent_test

import (
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
)

func TestListOf(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)
	if got, want := l.Len(), 3; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	// The caller's backing array must not be retained: mutating the source
	// slice after construction must not reach the List.
	source := []int{1, 2, 3}
	fromSlice := persistent.ListOf(source...)
	source[0] = 999
	if got := fromSlice.At(0); got != 1 {
		t.Fatalf("At(0) = %d after external mutation of source, want 1 (unaffected)", got)
	}
}

func TestCollectList(t *testing.T) {
	t.Parallel()

	seq := slices.Values([]string{"a", "b", "c"})
	l := persistent.CollectList(seq)

	if got, want := l.Items(), []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Fatalf("Items() = %v, want %v", got, want)
	}
}

func TestListAt(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf("x", "y", "z")

	tests := []struct {
		name string
		i    int
		want string
	}{
		{"first", 0, "x"},
		{"middle", 1, "y"},
		{"last", 2, "z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := l.At(tt.i); got != tt.want {
				t.Errorf("At(%d) = %q, want %q", tt.i, got, tt.want)
			}
		})
	}
}

func TestListAtPanicsOutOfRange(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)

	tests := []struct {
		name string
		i    int
	}{
		{"negative", -1},
		{"equal to len", 3},
		{"far beyond len", 100},
		{"empty list index 0", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			target := l
			if tt.name == "empty list index 0" {
				target = persistent.ListOf[int]()
			}

			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("At did not panic")
				}
				if r != "persistent: index out of range" {
					t.Fatalf("panic value = %v, want %q", r, "persistent: index out of range")
				}
			}()
			_ = target.At(tt.i)
		})
	}
}

func TestListAppend(t *testing.T) {
	t.Parallel()

	base := persistent.ListOf(1, 2)
	before := slices.Collect(base.Values())

	appended := base.Append(3, 4)

	if got, want := slices.Collect(appended.Values()), []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Append result = %v, want %v", got, want)
	}
	// The receiver must be unchanged: same snapshot before and after.
	after := slices.Collect(base.Values())
	if !slices.Equal(before, after) {
		t.Fatalf("receiver mutated by Append: before %v, after %v", before, after)
	}
}

func TestListAppendNoArgs(t *testing.T) {
	t.Parallel()

	base := persistent.ListOf(1, 2)
	result := base.Append()
	if !result.Equal(base) {
		t.Fatalf("Append() with no args = %v, want equal to base %v", result.Items(), base.Items())
	}
}

func TestListSet(t *testing.T) {
	t.Parallel()

	base := persistent.ListOf("a", "b", "c")
	before := slices.Collect(base.Values())

	updated := base.Set(1, "B")

	if got, want := slices.Collect(updated.Values()), []string{"a", "B", "c"}; !slices.Equal(got, want) {
		t.Fatalf("Set result = %v, want %v", got, want)
	}
	after := slices.Collect(base.Values())
	if !slices.Equal(before, after) {
		t.Fatalf("receiver mutated by Set: before %v, after %v", before, after)
	}
}

func TestListSetPanicsOutOfRange(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)

	defer func() {
		r := recover()
		if r != "persistent: index out of range" {
			t.Fatalf("panic value = %v, want %q", r, "persistent: index out of range")
		}
	}()
	_ = l.Set(-1, 0)
}

func TestListDelete(t *testing.T) {
	t.Parallel()

	base := persistent.ListOf(1, 2, 3, 4)
	before := slices.Collect(base.Values())

	deleted := base.Delete(1)

	if got, want := slices.Collect(deleted.Values()), []int{1, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Delete result = %v, want %v", got, want)
	}
	after := slices.Collect(base.Values())
	if !slices.Equal(before, after) {
		t.Fatalf("receiver mutated by Delete: before %v, after %v", before, after)
	}
}

func TestListDeletePanicsOutOfRange(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)

	defer func() {
		r := recover()
		if r != "persistent: index out of range" {
			t.Fatalf("panic value = %v, want %q", r, "persistent: index out of range")
		}
	}()
	_ = l.Delete(3)
}

func TestListValues(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)
	var got []int
	for v := range l.Values() {
		got = append(got, v)
	}
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Values() yielded %v, want %v", got, want)
	}
}

func TestListValuesStopsEarly(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3, 4, 5)
	var got []int
	for v := range l.Values() {
		got = append(got, v)
		if v == 2 {
			break
		}
	}
	if want := []int{1, 2}; !slices.Equal(got, want) {
		t.Fatalf("Values() early break yielded %v, want %v", got, want)
	}
}

func TestListAll(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf("a", "b", "c")
	var gotIdx []int
	var gotVal []string
	for i, v := range l.All() {
		gotIdx = append(gotIdx, i)
		gotVal = append(gotVal, v)
	}
	if want := []int{0, 1, 2}; !slices.Equal(gotIdx, want) {
		t.Fatalf("All() indices = %v, want %v", gotIdx, want)
	}
	if want := []string{"a", "b", "c"}; !slices.Equal(gotVal, want) {
		t.Fatalf("All() values = %v, want %v", gotVal, want)
	}
}

func TestListItemsCopies(t *testing.T) {
	t.Parallel()

	l := persistent.ListOf(1, 2, 3)
	items := l.Items()
	items[0] = 999

	if got := l.At(0); got != 1 {
		t.Fatalf("At(0) = %d after mutating Items() result, want 1 (List unaffected)", got)
	}
}

func TestListEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    persistent.List[int]
		b    persistent.List[int]
		want bool
	}{
		{"both empty", persistent.ListOf[int](), persistent.ListOf[int](), true},
		{"same elements same order", persistent.ListOf(1, 2, 3), persistent.ListOf(1, 2, 3), true},
		{"different order", persistent.ListOf(1, 2, 3), persistent.ListOf(3, 2, 1), false},
		{"different length", persistent.ListOf(1, 2), persistent.ListOf(1, 2, 3), false},
		{"different values", persistent.ListOf(1, 2, 3), persistent.ListOf(1, 9, 3), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListIsEmpty(t *testing.T) {
	t.Parallel()

	var zero persistent.List[int]
	if !zero.IsEmpty() {
		t.Fatal("zero List.IsEmpty() = false, want true")
	}
	if !persistent.ListOf[int]().IsEmpty() {
		t.Fatal("empty ListOf.IsEmpty() = false, want true")
	}
	if persistent.ListOf(1).IsEmpty() {
		t.Fatal("non-empty List.IsEmpty() = true, want false")
	}
}

func TestListEqualFunc(t *testing.T) {
	t.Parallel()

	eq := func(a, b int) bool { return a == b }
	alwaysFalse := func(a, b int) bool { return false }

	tests := []struct {
		name string
		a    persistent.List[int]
		b    persistent.List[int]
		eq   func(int, int) bool
		want bool
	}{
		{"both empty", persistent.ListOf[int](), persistent.ListOf[int](), eq, true},
		{"same elements", persistent.ListOf(1, 2, 3), persistent.ListOf(1, 2, 3), eq, true},
		{"different order", persistent.ListOf(1, 2, 3), persistent.ListOf(3, 2, 1), eq, false},
		{"different length", persistent.ListOf(1, 2), persistent.ListOf(1, 2, 3), eq, false},
		{"eq rejects pair", persistent.ListOf(1, 2, 3), persistent.ListOf(1, 2, 3), alwaysFalse, false},
		{
			"custom abs equality",
			persistent.ListOf(1, -2, 3),
			persistent.ListOf(-1, 2, -3),
			func(a, b int) bool {
				if a < 0 {
					a = -a
				}
				if b < 0 {
					b = -b
				}
				return a == b
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.EqualFunc(tt.b, tt.eq); got != tt.want {
				t.Errorf("EqualFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListAppendPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := persistent.ListOf(1)
	v2 := v1.Append(2)
	v3 := v2.Append(3, 4)

	if got := v1.Items(); !slices.Equal(got, []int{1}) {
		t.Fatalf("v1.Items() = %v, want [1]", got)
	}
	if got := v2.Items(); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("v2.Items() = %v, want [1 2]", got)
	}
	if got := v3.Items(); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("v3.Items() = %v, want [1 2 3 4]", got)
	}
}

func TestListSetPreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := persistent.ListOf(10, 20, 30)
	v2 := v1.Set(1, 200)
	v3 := v2.Set(0, 100).Set(2, 300)

	if got := v1.Items(); !slices.Equal(got, []int{10, 20, 30}) {
		t.Fatalf("v1.Items() = %v, want [10 20 30]", got)
	}
	if got := v2.Items(); !slices.Equal(got, []int{10, 200, 30}) {
		t.Fatalf("v2.Items() = %v, want [10 200 30]", got)
	}
	if got := v3.Items(); !slices.Equal(got, []int{100, 200, 300}) {
		t.Fatalf("v3.Items() = %v, want [100 200 300]", got)
	}
}

func TestListDeletePreservesOlderVersions(t *testing.T) {
	t.Parallel()

	v1 := persistent.ListOf(1, 2, 3, 4)
	v2 := v1.Delete(1) // drop 2 → [1, 3, 4]
	v3 := v2.Delete(0) // drop 1 → [3, 4]
	v4 := v3.Append(5) // [3, 4, 5]

	if got := v1.Items(); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("v1.Items() = %v, want [1 2 3 4]", got)
	}
	if got := v2.Items(); !slices.Equal(got, []int{1, 3, 4}) {
		t.Fatalf("v2.Items() = %v, want [1 3 4]", got)
	}
	if got := v3.Items(); !slices.Equal(got, []int{3, 4}) {
		t.Fatalf("v3.Items() = %v, want [3 4]", got)
	}
	if got := v4.Items(); !slices.Equal(got, []int{3, 4, 5}) {
		t.Fatalf("v4.Items() = %v, want [3 4 5]", got)
	}
}
