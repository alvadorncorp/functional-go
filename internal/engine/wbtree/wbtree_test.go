package wbtree

import (
	"cmp"
	"slices"
	"strings"
	"testing"
)

func TestBasicSetGetDeleteLen(t *testing.T) {
	t.Parallel()

	r := New[string, int](cmp.Compare[string])
	if r.Len() != 0 {
		t.Fatalf("empty Len = %d", r.Len())
	}
	if _, ok := r.Get("missing"); ok {
		t.Fatal("Get on empty must miss")
	}

	r1 := r.Set("b", 2)
	r2 := r1.Set("a", 1)
	r3 := r2.Set("c", 3)
	if r.Len() != 0 || r1.Len() != 1 || r2.Len() != 2 || r3.Len() != 3 {
		t.Fatalf("lens: empty=%d r1=%d r2=%d r3=%d", r.Len(), r1.Len(), r2.Len(), r3.Len())
	}
	assertGet(t, r3, "a", 1)
	assertGet(t, r3, "b", 2)
	assertGet(t, r3, "c", 3)
	if _, ok := r1.Get("a"); ok {
		t.Fatal("r1 should not contain a")
	}

	d := r3.Delete("b")
	if d.Len() != 2 {
		t.Fatalf("after delete Len = %d", d.Len())
	}
	if _, ok := d.Get("b"); ok {
		t.Fatal("deleted key still present")
	}
	assertGet(t, d, "a", 1)
	assertGet(t, d, "c", 3)
	assertGet(t, r3, "b", 2) // prior version intact
}

func TestReplaceDoesNotGrow(t *testing.T) {
	t.Parallel()

	r := New[int, string](cmp.Compare[int]).Set(1, "one").Set(2, "two")
	r2 := r.Set(1, "ONE")
	if r2.Len() != 2 {
		t.Fatalf("replace grew size: %d", r2.Len())
	}
	assertGet(t, r2, 1, "ONE")
	assertGet(t, r, 1, "one")
}

// TestSetReplacesKeyAndValue pins the key-identity rule: under a comparator
// that treats "Go" and "GO" as equivalent, the latest key and value win.
func TestSetReplacesKeyAndValue(t *testing.T) {
	t.Parallel()

	fold := func(a, b string) int { return cmp.Compare(strings.ToLower(a), strings.ToLower(b)) }
	base := New[string, int](fold).Set("Go", 1).Set("Rust", 2)

	pure := base.Set("GO", 10)
	if k := firstKey(pure); k != "GO" {
		t.Fatalf("pure Set first key = %q, want GO", k)
	}
	assertGet(t, pure, "go", 10)
	assertGet(t, base, "go", 1)
	if k := firstKey(base); k != "Go" {
		t.Fatalf("base first key = %q, want Go", k)
	}

	owner := newOwner(t)
	transient := base.TransientSet(owner, "gO", 20)
	if k := firstKey(transient); k != "gO" {
		t.Fatalf("TransientSet first key = %q, want gO", k)
	}
	assertGet(t, transient, "GO", 20)
	if transient.Len() != 2 {
		t.Fatalf("Len = %d, want 2", transient.Len())
	}
}

func TestAllAscendingAndEarlyStop(t *testing.T) {
	t.Parallel()

	r := New[int, int](cmp.Compare[int])
	for _, k := range []int{5, 3, 9, 1, 7, 2, 8} {
		r = r.Set(k, k*10)
	}
	var keys []int
	for k, v := range r.All() {
		if v != k*10 {
			t.Fatalf("value for %d = %d", k, v)
		}
		keys = append(keys, k)
	}
	if want := []int{1, 2, 3, 5, 7, 8, 9}; !slices.Equal(keys, want) {
		t.Fatalf("All keys = %v, want %v", keys, want)
	}

	var seen []int
	for k := range r.All() {
		seen = append(seen, k)
		if len(seen) == 3 {
			break
		}
	}
	if want := []int{1, 2, 3}; !slices.Equal(seen, want) {
		t.Fatalf("early stop saw %v, want %v", seen, want)
	}
}

func TestReversedComparatorOrdersDescending(t *testing.T) {
	t.Parallel()

	r := New[int, int](func(a, b int) int { return cmp.Compare(b, a) })
	for k := range 10 {
		r = r.Set(k, k)
	}
	checkInvariants(t, r)
	var got []int
	for k := range r.All() {
		got = append(got, k)
	}
	if want := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}; !slices.Equal(got, want) {
		t.Fatalf("All = %v, want %v", got, want)
	}
}

func TestNewNilComparatorPanics(t *testing.T) {
	t.Parallel()

	assertPanics(t, "wbtree: nil comparator", func() { New[int, int](nil) })
}

func TestZeroRoot(t *testing.T) {
	t.Parallel()

	var r Root[string, int]
	if r.HasComparator() || r.Len() != 0 {
		t.Fatalf("zero root: HasComparator=%v Len=%d", r.HasComparator(), r.Len())
	}
	if _, ok := r.Get("k"); ok {
		t.Fatal("Get on zero root must miss")
	}
	for k := range r.All() {
		t.Fatalf("zero root yielded %q", k)
	}
	if d := r.Delete("k"); d.node != nil || d.HasComparator() {
		t.Fatal("Delete on zero root must be a no-op")
	}
	owner := newOwner(t)
	if d := r.TransientDelete(owner, "k"); d.node != nil || d.HasComparator() {
		t.Fatal("TransientDelete on zero root must be a no-op")
	}
	assertPanics(t, "wbtree: no comparator", func() { r.Set("k", 1) })
	assertPanics(t, "wbtree: no comparator", func() { r.TransientSet(owner, "k", 1) })
}

func firstKey[V any](r Root[string, V]) string {
	for k := range r.All() {
		return k
	}
	return ""
}

func assertGet[K any, V comparable](t *testing.T, r Root[K, V], key K, want V) {
	t.Helper()
	got, ok := r.Get(key)
	if !ok {
		t.Fatalf("Get(%v) missing", key)
	}
	if got != want {
		t.Fatalf("Get(%v) = %v, want %v", key, got, want)
	}
}

func assertPanics(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		if got := recover(); got != want {
			t.Fatalf("panic = %v, want %q", got, want)
		}
	}()
	fn()
}
