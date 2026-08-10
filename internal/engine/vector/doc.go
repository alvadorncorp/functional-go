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
// leaf must be pulled from the trie into the tail. Delete mid-index shifts
// subsequent elements left and is O(n) in this version (rebuild via sequential
// walk); it is correct, not log-n path-aware. All is O(n) via the same
// sequential trie+tail walk (not per-index At).
//
// This package does not implement RRB, concat, or split.
//
// Allowed non-stdlib imports (when needed): ownership, cow, sharing.
// Must not import transient, eq, optional, immutable, valueobject,
// keyhash, or hamt.
package vector
