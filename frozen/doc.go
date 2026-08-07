// Package frozen provides read-only, allocation-copying snapshots of a slice
// or map. Slice and Map copy on construction and copy again on every read
// that could otherwise leak the backing array or map — so no caller can reach
// or mutate what a Slice or Map holds. Neither type has an update operation
// and never will: for a collection that evolves through functional updates,
// use functional/persistent.
//
// The zero Slice and the zero Map are both valid and empty: a struct field of
// either type needs no constructor. Construction (NewSlice, CollectSlice,
// NewMap, CollectMap) copies its input, and Items copies its output; Values,
// All, Keys, and Get read without copying or allocating.
//
// Map.Get returns (V, bool), deliberately NOT functional/optional.Optional[V]:
// it is this choice that keeps frozen independent of optional, since no
// public subpackage of functional imports another. Bridge at the call site
// with optional.FromOK(m.Get(k)).
//
// Equal on both types delegates element-wise to functional/internal/eq.Equal,
// the same equality resolution (an Equal method, then ==; non-comparable
// types without Equal return false) the rest of the module uses.
package frozen
