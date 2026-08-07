// Package persistent provides immutable collections — List, Map, and Set —
// that write functionally: every update method returns a new collection and
// leaves the receiver unchanged. There is no in-place mutation anywhere in
// this API; "write" always means "derive a new value from an old one."
//
// # Constructors
//
//   - ListOf[T](items ...T) List[T]
//   - MapOf[K,V](m map[K]V) Map[K,V]
//   - SetOf[T](items ...T) Set[T]
//
// CollectList / CollectMap / CollectSet drain an iter.Seq into a new
// collection. There are no NewList / NewMap / NewSet constructors.
//
// # Semantic immutability
//
// Collections are semantically immutable: the structure never mutates in
// place. Values (and map keys) are stored as-is; the package does not deep-
// copy them. If T or V is a pointer, slice, map, or other shared reference,
// a caller that mutates the referent can still observe that change through
// every collection version that holds the same reference. Prefer immutable
// element types (or frozen / value-object wrappers) when deep immutability
// of values is required.
//
// # Engines and structural sharing
//
// List is backed by an internal 32-way persistent vector
// (functional/internal/engine/vector) with path-copy structural sharing.
// Map and Set are backed by an internal HAMT
// (functional/internal/engine/hamt); each empty Root carries a maphash.Seed
// (hashing via maphash.Comparable) preserved across derived versions.
// Single updates allocate along a tree path (depth curve), not a full
// collection copy. Core (ownership/cow/sharing/transient) is laws only;
// engines own representation.
//
// # Builders
//
// Builders open an internal transient.Window (see
// functional/internal/core/transient), edit under the window's exclusive
// owner, and Freeze on Build (O(1) Owner.Invalidate — no tree walk).
// ListBuilder exposes At; MapBuilder exposes Get/Has; SetBuilder exposes
// Has. Builders are mutable and not safe for concurrent use. After Build,
// every builder method panics with "persistent: builder used after Build".
//
// # Zero values and seed materialization
//
// The zero List, Map, and Set are valid empty collections (Len 0, IsEmpty
// true; reads miss / At panics on empty List). Map and Set first write
// (Set/Delete/Add/Remove) or Builder() materializes a proper hamt.Empty
// seed so the lineage never freezes a zero maphash.Seed. Independent zero
// Maps/Sets each get their own seed on first materialization.
//
// # Concurrency
//
// Published (immutable) collections are safe for concurrent read, including
// while another goroutine derives new versions from the same base. Builders
// are single-owner and not thread-safe.
//
// # Equality and reflection
//
// Production code in this package does not import reflect. List.Equal and
// Map.Equal delegate to functional/internal/eq.Equal, which never uses
// reflect.DeepEqual: it prefers Equal(T) bool, then interface == for
// comparable dynamic values, and returns false for non-comparable types
// without Equal. Set.Equal uses membership alone (T is comparable).
// List.EqualFunc and Map.EqualFunc accept a caller-supplied comparator.
//
// Public panic strings remain "persistent: ..." — engine panics such as
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
//     on it. List iterates in index order; that order is specified and part
//     of the contract.
//
//  3. Documented complexity is an upper bound the public API reserves room
//     for, not a description of today's constants:
//
//     Len, IsEmpty           O(1)
//     At, Get, Has           O(log n)
//     Set, Delete            O(log n)
//     Append                 O(log n) amortized
//
//     List.Delete is currently O(n) (vector rebuild by iteration); a caller
//     MUST NOT depend on a tighter bound. Only the table above is guaranteed
//     as API headroom for the methods that meet it, and only the bound is
//     ever documented on a method — never today's actual constant.
//
//  4. List, Map, and Set are deliberately NOT ==-comparable. Use Equal or
//     EqualFunc. Comparing by == would mean comparing identity of the
//     underlying representation.
//
//  5. Calling Append/Set/Delete n times in a loop still allocates n
//     intermediate collections. Builder exists to give a real linear path for
//     that case: <Type>.Builder() seeds a mutable builder from a snapshot
//     (and opens a transient window), batch writes go through it, and Build()
//     freezes the window, hands back an immutable collection, and permanently
//     disables the builder. Building in a loop MUST go through Builder.
//
//  6. Every test in this package is package persistent_test (external) and
//     asserts only publicly observable behavior, never the internal
//     representation. That suite is the acceptance criterion for any future
//     engine change: it must keep passing unchanged.
package persistent
