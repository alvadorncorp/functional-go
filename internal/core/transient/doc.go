// Package transient provides a single-threaded mutation window that holds
// an exclusive ownership.Owner for a batch of edits.
//
// Freeze is O(1): it Invalidates the owner token and clears the window.
// It does not walk collection nodes and does not require Detach. Nodes may
// keep stale inactive owner stamps; IsOwnedBy fails for those stamps, so a
// later window treats them as published and copy-on-writes.
package transient
