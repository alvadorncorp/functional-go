package immutable

// panicIndexOutOfRange is what List.At, List.Set, List.Delete,
// ListBuilder.At, ListBuilder.Set, and ListBuilder.Delete panic with when
// the index is outside [0, Len).
const panicIndexOutOfRange = "immutable: index out of range"

// panicBuilderUsedAfterBuild is what any method on a ListBuilder, MapBuilder,
// SetBuilder, or SortedMapBuilder panics with once Build has already run on
// that builder. Build freezes the transient window and clears builder state
// so no method can honor a call afterward. Public builders never leak
// "transient: ..." panics.
const panicBuilderUsedAfterBuild = "immutable: builder used after Build"

// panicNilComparator is what EmptySortedMap and CollectSortedMap panic with
// when the comparator is nil.
const panicNilComparator = "immutable: nil comparator"

// panicZeroSortedMap is what SortedMap.Set and SortedMap.Builder panic with
// on the zero SortedMap, which has no comparator. It replaces the engine's
// "wbtree: ..." panic so engine strings never reach callers.
const panicZeroSortedMap = "immutable: zero SortedMap has no comparator; " +
	"construct with EmptySortedMap or CollectSortedMap"
