package persistent

// panicIndexOutOfRange is what List.At, List.Set, List.Delete, and their
// ListBuilder equivalents panic with when the index is outside [0, Len).
const panicIndexOutOfRange = "persistent: index out of range"

// panicBuilderUsedAfterBuild is what any method on a ListBuilder, MapBuilder,
// or SetBuilder panics with once Build has already run on that builder.
// Build transfers ownership of the builder's internal buffer to the
// collection it returns, so no method can honor a call afterward.
const panicBuilderUsedAfterBuild = "persistent: builder used after Build"
