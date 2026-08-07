// Package vector is the 32-way persistent vector engine behind persistent List.
//
// Pure immutable At/Set/Append path-copy with branching factor 32 (5 index
// bits per level). New path nodes for pure immutable ops always publish with
// owner == nil. TransientAppend/TransientSet/TransientDelete take an
// *ownership.Owner and mutate exclusively owned nodes in place; this package
// does not import transient — callers pass an owner from a transient.Window.
//
// The representation is the classic Clojure-style persistent vector: a dense
// trie of 32-way nodes plus a tail array of up to 32 elements. Structural
// sharing retains untouched sibling subtrees across Set/Append. Delete shifts
// subsequent elements left and may be O(n) in this version (rebuild via
// iteration); it is correct, not log-n path-aware.
//
// This package does not implement RRB, concat, or split.
//
// Allowed non-stdlib imports (when needed): ownership, cow, sharing.
// Must not import transient, eq, optional, frozen, persistent, valueobject,
// keyhash, or hamt.
package vector
