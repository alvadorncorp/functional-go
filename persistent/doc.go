// Package persistent provides immutable collections — List, Map, and Set —
// that write functionally: every update method returns a new collection and
// leaves the receiver unchanged. There is no in-place mutation anywhere in
// this API; "write" always means "derive a new value from an old one."
//
// List is backed by an internal 32-way persistent vector
// (functional/internal/engine/vector) with path-copy structural sharing.
// Map and Set are backed by an internal HAMT
// (functional/internal/engine/hamt); each empty Root carries a maphash.Seed
// (hashing via maphash.Comparable) preserved across derived versions.
// Builders open an internal transient.Window (see
// functional/internal/core/transient), edit under the window's exclusive
// owner, and Freeze on Build (O(1) Owner.Invalidate — no tree walk). Core
// (ownership/cow/sharing/transient) is laws only; engines own representation.
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
//     Len                    O(1)
//     At, Get, Has           O(log n)
//     Set, Delete            O(log n)
//     Append                 O(log n) amortized
//
//     List.Delete is currently O(n) (vector rebuild by iteration); a caller
//     MUST NOT depend on a tighter bound. Only the table above is guaranteed
//     as API headroom for the methods that meet it, and only the bound is
//     ever documented on a method — never today's actual constant.
//
//  4. List, Map, and Set are deliberately NOT ==-comparable. Use Equal.
//     Comparing by == would mean comparing identity of the underlying
//     representation.
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
