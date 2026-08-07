package persistent

// panicIndexOutOfRange is what List.At, List.Set, List.Delete, and their
// ListBuilder equivalents panic with when the index is outside [0, Len).
const panicIndexOutOfRange = "persistent: index out of range"

// panicBuilderUsedAfterBuild is what any method on a ListBuilder, MapBuilder,
// or SetBuilder panics with once Build has already run on that builder.
// Build freezes the transient window and clears builder state so no method
// can honor a call afterward. Public builders never leak "transient: ..."
// panics.
const panicBuilderUsedAfterBuild = "persistent: builder used after Build"
