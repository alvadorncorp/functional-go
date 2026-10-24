// Package vector is the 32-way persistent vector engine behind immutable.List.
//
// Pure immutable At/Set/Append/Pop path-copy with branching factor 32 (5 index
// bits per level). New path nodes for pure immutable ops always publish with
// owner == nil. TransientAppend/TransientSet/TransientDelete/TransientPop take
// an *ownership.Owner and mutate exclusively owned nodes in place; this package
// does not import transient — callers pass an owner from a transient.Window.
//
// The representation is the classic Clojure-style persistent vector: a dense
// trie of 32-way nodes plus a tail array of up to 32 elements. Structural
// sharing retains untouched sibling subtrees across Set/Append/Pop. Pop is O(1)
// when the tail has more than one element and O(log32 n) when the last full
// leaf must be pulled from the trie into the tail. Deleting the last element
// uses the same path as Pop (TransientPop for owner-scoped edits). Delete at
// other indices shifts subsequent elements left and is O(n) in this version
// (rebuild via sequential walk). All is O(n) via the same sequential trie+tail
// walk (not per-index At).
//
// This package does not implement RRB, concat, or split.
//
// Allowed functional-module imports: internal/core/ownership, cow and sharing.
// Any other functional package is forbidden; in particular transient, eq,
// immutable, state, values/optional, keyhash and the peer engines hamt and
// wbtree. The authority is internal/archtest (vectorAllowedFunctional,
// engineForbiddenTargets).
package vector
