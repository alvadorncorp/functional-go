// Package transient provides a single-threaded mutation window that holds
// an exclusive ownership.Owner for a batch of edits.
//
// Freeze invalidates the window; it does not walk collection nodes. Stamping
// or detaching owners on nodes is the collection layer's responsibility at
// publish time.
package transient
