// Package sharing provides publish-side primitives for structural sharing:
// detaching an exclusive owner stamp so a node may be held by multiple
// parents without further in-place mutation.
//
// This is the dual of package cow: cow enters exclusive mutation; sharing
// leaves exclusive mutation. Neither package clones structure-specific nodes.
package sharing
