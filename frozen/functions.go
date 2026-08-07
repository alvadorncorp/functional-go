package frozen

// SliceEqualFunc reports whether a and b have the same length and eq reports
// true for every pair of elements at the same index.
func SliceEqualFunc[T, U any](a Slice[T], b Slice[U], eq func(T, U) bool) bool {
	if len(a.items) != len(b.items) {
		return false
	}
	for i := range a.items {
		if !eq(a.items[i], b.items[i]) {
			return false
		}
	}
	return true
}

// MapEqualFunc reports whether a and b have the same length and, for every
// key in a, b holds the same key with a value eq considers equal.
func MapEqualFunc[K comparable, V1, V2 any](a Map[K, V1], b Map[K, V2], eq func(V1, V2) bool) bool {
	if len(a.items) != len(b.items) {
		return false
	}
	for k, v1 := range a.items {
		v2, ok := b.items[k]
		if !ok || !eq(v1, v2) {
			return false
		}
	}
	return true
}
