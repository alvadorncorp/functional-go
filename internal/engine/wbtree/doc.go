// Package wbtree is the persistent weight-balanced tree engine behind
// immutable.SortedMap.
//
// Keys are ordered by a caller-supplied comparator. Each Root carries its
// comparator, and every derived version (Set/Delete) preserves it, including
// a version emptied by Delete. A zero Root has no comparator: it reads as
// empty and Delete on it is a no-op, but Set panics because K has no default
// order.
//
// Balance follows Straka's proven parameters for Adams' trees with weight
// size+1: delta = 3 and gamma = 2. Every node satisfies
// isBalanced(left, right) and isBalanced(right, left); one insert or delete
// is repaired by a single or double rotation per level. Rotations look only
// at sizes, never at the comparator.
//
// Pure Set/Delete path-copy and publish new nodes with owner == nil.
// TransientSet/TransientDelete take a *ownership.Owner and mutate exclusively
// owned nodes in place, shallow-cloning foreign or published nodes through
// cow.EnsureEditable. Engines do not import transient — callers pass the
// owner from a transient.Window. Publish is O(1): Invalidate the owner.
//
// Every comparator call on a path runs while descending, before any node is
// edited on the way back up, so a comparator panic leaves owned nodes
// untouched. Set replaces both the stored key and the value of an equivalent
// key. Deleting an absent key returns the same node pointer.
//
// Allowed non-stdlib imports: ownership, cow.
// Must not import transient, eq, sharing, keyhash, hamt, vector, or any
// public package.
package wbtree
