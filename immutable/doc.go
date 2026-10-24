// Package immutable provides immutable collections — List, Map, Set, and
// SortedMap — that write functionally: every update method returns a new
// collection and leaves the receiver unchanged. There is no in-place
// mutation anywhere in this API; "write" always means "derive a new value
// from an old one."
//
// # Constructors
//
//   - ListOf[T](items ...T) List[T]
//   - MapOf[K,V](m map[K]V) Map[K,V]
//   - SetOf[T](items ...T) Set[T]
//   - EmptySortedMap[K,V](cmp func(a, b K) int) SortedMap[K,V]
//
// CollectList / CollectMap / CollectSet drain an iter.Seq into a new
// collection; CollectSortedMap(cmp, seq) does the same for SortedMap. There
// are no NewList / NewMap / NewSet / NewSortedMap constructors. A nil
// comparator panics with "immutable: nil comparator".
//
// # Semantic immutability
//
// Collections are semantically immutable: the structure never mutates in
// place. Values (and map keys) are stored as-is; the package does not deep-
// copy them. If T or V is a pointer, slice, map, or other shared reference,
// a caller that mutates the referent can still observe that change through
// every collection version that holds the same reference. Prefer immutable
// element types (or value-object wrappers) when deep immutability
// of values is required.
//
// # Engines and structural sharing
//
// List is backed by an internal 32-way persistent vector
// (functional/internal/engine/vector) with path-copy structural sharing.
// Map and Set are backed by an internal HAMT
// (functional/internal/engine/hamt); each empty Root carries a maphash.Seed
// (hashing via maphash.Comparable) preserved across derived versions.
// SortedMap is backed by an internal weight-balanced tree
// (functional/internal/engine/wbtree) ordered by its comparator, which every
// derived version keeps. Single updates allocate along a tree path (depth
// curve), not a full collection copy. Core
// (ownership/cow/sharing/transient) is laws only; engines own
// representation.
//
// # Builders
//
// Builders open an internal transient.Window (see
// functional/internal/core/transient), edit under the window's exclusive
// owner, and Freeze on Build (O(1) Owner.Invalidate — no tree walk).
// ListBuilder exposes At; MapBuilder and SortedMapBuilder expose Get/Has;
// SetBuilder exposes Has. A comparator panic during a SortedMapBuilder write
// leaves the builder's prior contents intact. Builders are mutable and not
// safe for concurrent use. After Build, every builder method panics with
// "immutable: builder used after Build".
//
// # Zero values and seed materialization
//
// The zero List, Map, and Set are valid empty collections (Len 0, IsEmpty
// true; reads miss / At panics on empty List). Map and Set first write
// (Set/Update/Delete/Add/Remove) or Builder() materializes a proper hamt.Empty
// seed so the lineage never freezes a zero maphash.Seed. Independent zero
// Maps/Sets each get their own seed on first materialization. MapOf, SetOf,
// and the Collect functions given empty input return the zero value without
// creating a seed; the seed appears on the first write.
//
// SortedMap is the deliberate exception: K has no default order, so the zero
// SortedMap has no comparator. It reads as empty and Delete on it is a no-op,
// but Set and Builder() panic with "immutable: zero SortedMap has no
// comparator; construct with EmptySortedMap or CollectSortedMap". A zero-safe
// getter for a SortedMap field must return EmptySortedMap(cmp), never the
// zero SortedMap.
//
// # Concurrency
//
// Published (immutable) collections are safe for concurrent read, including
// while another goroutine derives new versions from the same base. Builders
// are single-owner and not thread-safe.
//
// # Equality and reflection
//
// Production code in this package does not import reflect. List.Equal,
// Map.Equal, and SortedMap.Equal select functional/internal/eq's value
// comparison strategy once per call. It never uses reflect.DeepEqual: it
// prefers Equal(T) bool, then == for comparable values, and returns false for
// non-comparable types without Equal. Interface element types retain
// per-value dynamic dispatch. Set.Equal uses membership alone (T is comparable).
// List.Equal, Map.Equal, and Set.Equal return true for a shared root, even
// with NaN or non-reflexive Equal methods. Distinct roots still compare values
// or membership. SortedMap.Equal has no root-identity shortcut: it checks the
// entry count and looks up every key by the other map's comparator with an
// equal value. List.EqualFunc, Map.EqualFunc, and SortedMap.EqualFunc always
// use the caller-supplied value comparator, including on a shared version.
//
// Public panic strings remain "immutable: ..." — engine panics such as
// "vector: index out of range" are never leaked to callers. Documented
// complexity bounds below are upper-bound API headroom; a caller MUST NOT
// depend on today's constants.
//
//  1. No method accepts or returns the internal representation. Items always
//     copies rather than handing out the backing storage, so nothing outside
//     this package can reach into a collection's engine.
//
//  2. Map and Set iteration order is UNSPECIFIED and may change between
//     releases, or even between two calls on the same value — do not depend
//     on it. List iterates in index order, and SortedMap iterates in
//     ascending comparator order; those orders are specified and part of the
//     contract.
//
//  3. Documented complexity is an upper bound the public API reserves room
//     for, not a description of today's constants:
//
//     Len, IsEmpty           O(1)
//     At, Get, Has           O(log n)
//     Set, Delete, Update    O(log n)
//     Append                 O(log n) amortized
//     Union, Intersection,
//     Difference             O(min(m,n)·log max(m,n))
//
//     For Set.Union, Set.Intersection, and Set.Difference, m and n are the
//     operand sizes. Both operands are unchanged; which operand's structure
//     and hash-seed lineage the result shares is UNSPECIFIED.
//
//     List.Delete has an O(n) upper bound: deleting the last element uses
//     the vector's tail-pop path; other positions rebuild by iteration. A
//     caller MUST NOT depend on a tighter bound. The table above reserves API
//     headroom only for the methods that meet its bounds.
//
//  4. List, Map, Set, and SortedMap are deliberately NOT ==-comparable. Use
//     Equal or EqualFunc. Comparing by == would mean comparing identity of the
//     underlying representation.
//
//  5. Empty ListOf, MapOf, and SetOf return the zero collection, open no edit
//     window, and allocate nothing. Collect* given an empty sequence also
//     returns the zero, unseeded collection, but that is not an allocation
//     guarantee. A single-item Append, Set.Add, or Set.Remove, a one-entry
//     MapOf, and the first item of a Collect call edit through the pure path
//     with no transient window; that is not zero total allocation, only no
//     builder or edit-window setup. Multi-item List.Append and
//     Set.Add/Remove, and bulk constructors and collectors, batch edits under
//     one transient owner. Calling write methods n times in a loop still
//     creates intermediate collections: use <Type>.Builder() for that case.
//     It seeds a mutable builder from a snapshot, opens an exclusive transient
//     window, and freezes that window on Build(), returning an immutable
//     collection and permanently disabling the builder.
//
//  6. Every test in this package is package immutable_test (external) and
//     asserts only publicly observable behavior, never the internal
//     representation. That suite is the acceptance criterion for any future
//     engine change: it must keep passing unchanged.
package immutable
