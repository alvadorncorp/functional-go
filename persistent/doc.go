// Package persistent provides immutable collections — List, Map, and Set —
// that write functionally: every update method returns a new collection and
// leaves the receiver unchanged. There is no in-place mutation anywhere in
// this API; "write" always means "derive a new value from an old one."
//
// The engine underneath is full-copy copy-on-write over a stdlib slice or
// map today. Builders open an internal transient.Window (see
// functional/internal/core/transient) and Freeze it on Build; public panic
// strings remain "persistent: ...". Documented complexity bounds below are
// upper-bound API headroom for a possible future structure-sharing engine,
// not a description of today's constants or of the thin core packages.
//
//  1. No method accepts or returns the internal representation. Items always
//     copies rather than handing out the backing slice or map, so nothing
//     outside this package can reach into a collection's storage.
//
//  2. Map and Set iteration order is UNSPECIFIED and may change between
//     releases, or even between two calls on the same value — do not depend
//     on it. List iterates in index order; that order is specified and part
//     of the contract.
//
//  3. Documented complexity is an upper bound the public API reserves room
//     for, not a description of what full-copy COW does today:
//
//     Len                    O(1)
//     At, Get, Has           O(log n)
//     Set, Delete            O(log n)
//     Append                 O(log n) amortized
//
//     The current engine beats several of these (At is O(1) today, for
//     instance); a caller MUST NOT depend on that. Only the bound above is
//     guaranteed as API headroom, and only the bound is ever documented on a
//     method — never today's actual constant.
//
//  4. List, Map, and Set are deliberately NOT ==-comparable. Use Equal.
//     Comparing by == would mean comparing identity of the underlying
//     representation.
//
//  5. A single update is a full copy: O(n) with today's engine, so calling
//     Append/Set/Delete n times in a loop is O(n^2). Builder exists to give a
//     real linear path for that case: <Type>.Builder() seeds a mutable
//     builder from a snapshot (and opens a transient window), batch writes
//     go through it, and Build() freezes the window, hands back an immutable
//     collection, and permanently disables the builder. Building in a loop
//     MUST go through Builder.
//
//  6. Every test in this package is package persistent_test (external) and
//     asserts only publicly observable behavior, never the internal
//     representation. That suite is the acceptance criterion for any future
//     engine change: it must keep passing unchanged.
package persistent
