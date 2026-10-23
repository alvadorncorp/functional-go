package immutable_test

import (
	"cmp"
	"math"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

type rejectingValue int

func (rejectingValue) Equal(rejectingValue) bool { return false }

type sliceEqualer []int

func (a sliceEqualer) Equal(b sliceEqualer) bool {
	return len(a) == len(b)
}

type dynamicEqualer int

func (a dynamicEqualer) Equal(b any) bool {
	v, ok := b.(dynamicEqualer)
	return ok && a%10 == v%10
}

func assertCollectionEquality[T any](t *testing.T, a, b T, want bool) {
	t.Helper()
	checks := []struct {
		name string
		got  bool
	}{
		{"List", immutable.ListOf(a).Equal(immutable.ListOf(b))},
		{"Map", immutable.MapOf(map[int]T{1: a}).Equal(immutable.MapOf(map[int]T{1: b}))},
		{"SortedMap", immutable.EmptySortedMap[int, T](cmp.Compare[int]).Set(1, a).
			Equal(immutable.EmptySortedMap[int, T](cmp.Compare[int]).Set(1, b))},
	}
	for _, check := range checks {
		if check.got != want {
			t.Errorf("%s.Equal = %v, want %v", check.name, check.got, want)
		}
	}
}

func TestCollectionEqualDispatch(t *testing.T) {
	t.Parallel()

	t.Run("concrete equaler overrides comparable", func(t *testing.T) {
		assertCollectionEquality(t, rejectingValue(1), rejectingValue(1), false)
	})
	t.Run("noncomparable equaler", func(t *testing.T) {
		assertCollectionEquality(t, sliceEqualer{1}, sliceEqualer{9}, true)
		assertCollectionEquality(t, sliceEqualer{1}, sliceEqualer{1, 2}, false)
	})
	t.Run("interface dynamic dispatch", func(t *testing.T) {
		assertCollectionEquality[any](t, dynamicEqualer(1), dynamicEqualer(11), true)
		assertCollectionEquality[any](t, dynamicEqualer(1), dynamicEqualer(2), false)
		assertCollectionEquality[any](t, nil, nil, true)
		assertCollectionEquality[any](t, nil, 1, false)
		assertCollectionEquality[any](t, 3, 3, true)
		assertCollectionEquality[any](t, 3, "3", false)
		assertCollectionEquality[any](t, []int{1}, []int{1}, false)
	})
}

func TestCollectionEqualSharedVersion(t *testing.T) {
	t.Parallel()

	nan := math.NaN()
	list := immutable.ListOf(nan)
	m := immutable.MapOf(map[int]float64{1: nan})
	set := immutable.SetOf(nan, 1)
	sharedList, sharedMap, sharedSet := list, m, set
	if !list.Equal(sharedList) || !m.Equal(sharedMap) || !set.Equal(sharedSet) {
		t.Error("a shared List, Map or Set version must equal itself, including NaN")
	}
	if list.Equal(immutable.ListOf(nan)) || m.Equal(immutable.MapOf(map[int]float64{1: nan})) ||
		set.Equal(immutable.SetOf(nan, 1)) {
		t.Error("distinct roots must still compare NaN by value or membership")
	}
	if list.Equal(list.Set(0, nan)) || m.Equal(m.Set(1, nan)) || set.Equal(set.Add(1)) {
		t.Error("distinct derived roots must not bypass NaN comparisons")
	}
	nanKey := immutable.MapOf(map[float64]int{nan: 1})
	sharedNaNKey := nanKey
	if !nanKey.Equal(sharedNaNKey) || nanKey.Equal(immutable.MapOf(map[float64]int{nan: 1})) {
		t.Error("NaN keys must be reflexive only for a shared Map version")
	}

	rejectingList := immutable.ListOf(rejectingValue(1))
	rejectingMap := immutable.MapOf(map[int]rejectingValue{1: 1})
	sharedRejectingList, sharedRejectingMap := rejectingList, rejectingMap
	if !rejectingList.Equal(sharedRejectingList) || !rejectingMap.Equal(sharedRejectingMap) {
		t.Error("shared versions must bypass non-reflexive Equal methods")
	}

	// SortedMap has no comparable root identity; it retains value comparison.
	sorted := immutable.EmptySortedMap[int, float64](cmp.Compare[int]).Set(1, nan)
	sharedSorted := sorted
	if sorted.Equal(sharedSorted) || sorted.Equal(sorted.Set(1, nan)) {
		t.Error("SortedMap must retain NaN value comparison")
	}
	for _, check := range []struct {
		name    string
		compare func(func(float64, float64) bool) bool
	}{
		{"List", func(equal func(float64, float64) bool) bool { return list.EqualFunc(sharedList, equal) }},
		{"Map", func(equal func(float64, float64) bool) bool { return m.EqualFunc(sharedMap, equal) }},
		{"SortedMap", func(equal func(float64, float64) bool) bool { return sorted.EqualFunc(sharedSorted, equal) }},
	} {
		calls := 0
		if check.compare(func(float64, float64) bool { calls++; return false }) || calls != 1 {
			t.Errorf("%s.EqualFunc shared root must call the comparator once; calls=%d", check.name, calls)
		}
	}
}
