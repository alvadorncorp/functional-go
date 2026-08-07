// Package persistent provides immutable collections — List, Map, and Set —
// that write functionally: every update method returns a new collection and
// leaves the receiver unchanged. There is no in-place mutation anywhere in
// this API; "write" always means "derive a new value from an old one."
//
// The engine underneath is copy-on-write over a stdlib slice or map today.
// The contract below is written so a future swap to a structure-sharing
// engine (a HAMT for Map/Set, an RRB tree for List) changes no caller. Every
// rule exists to make that swap invisible:
//
//  1. No method accepts or returns the internal representation. Items always
//     copies rather than handing out the backing slice or map, so nothing
//     outside this package can reach into a collection's storage.
//
//  2. Map and Set iteration order is UNSPECIFIED and may change between
//     releases, or even between two calls on the same engine — do not depend
//     on it. List iterates in index order; that order is specified and part
//     of the contract.
//
//  3. Documented complexity is an upper bound the future engine must honor,
//     not a description of what copy-on-write does today:
//
//     Len                    O(1)
//     At, Get, Has           O(log n)
//     Set, Delete            O(log n)
//     Append                 O(log n) amortized
//
//     The current engine beats several of these (At is O(1) today, for
//     instance); a caller MUST NOT depend on that. Only the bound above is
//     guaranteed, and only the bound is ever documented on a method — never
//     today's actual constant.
//
//  4. List, Map, and Set are deliberately NOT ==-comparable. Use Equal.
//     Comparing by == would mean comparing identity of the underlying
//     representation, which is exactly what structural sharing would change;
//     forbidding == is what stops a caller from ever depending on it.
//
//  5. A single update is a full copy: O(n) with today's engine, so calling
//     Append/Set/Delete n times in a loop is O(n^2). Builder exists to give a
//     real linear (or better) path for that case: <Type>.Builder() seeds a
//     mutable builder from a snapshot, batch writes go through it, and
//     Build() hands back an immutable collection while permanently disabling
//     the builder. This is the same shape a HAMT/RRB transient would expose,
//     which is why building-in-a-loop code does not change across the swap.
//     Building in a loop MUST go through Builder.
//
//  6. Every test in this package is package persistent_test (external) and
//     asserts only publicly observable behavior, never the internal
//     representation. That is deliberate: the suite is required to keep
//     passing unchanged the day the engine becomes a HAMT/RRB tree, and an
//     unchanged pass is the acceptance criterion for that swap.
package persistent
