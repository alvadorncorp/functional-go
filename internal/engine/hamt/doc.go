// Package hamt is the Hash Array Mapped Trie engine behind persistent Map
// and Set.
//
// Pure immutable Set/Delete path-copy with branching factor 32 (5 hash bits
// per level) and always publish new nodes with owner == nil. TransientSet/
// TransientDelete take a *ownership.Owner and mutate exclusively owned nodes
// in place, shallow-cloning foreign or published nodes. Engines do not import
// transient — callers pass the owner from a transient.Window.
//
// Publish after a builder batch is O(1): Invalidate the owner token. Nodes
// may keep stale inactive stamps; IsOwnedBy fails for them, so no Detach walk
// is required. Hash collisions land in multi-entry collision nodes. Each Root
// carries a maphash.Seed; derived versions preserve that seed so structural
// sharing remains valid.
//
// Replacing an existing key is not required to be a structural no-op even when
// the new value equals the old one (V is not comparable; no reflect equality).
//
// Allowed non-stdlib imports (when needed): ownership, cow, sharing, keyhash.
// Must not import transient, eq, optional, frozen, persistent, valueobject,
// or vector.
package hamt
