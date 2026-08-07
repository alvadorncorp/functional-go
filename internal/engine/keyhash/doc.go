// Package keyhash wraps hash/maphash.Comparable for comparable map and set
// keys used by the HAMT engine.
//
// The seed for hashing is not owned here: it lives on the HAMT Root so that
// an entire tree shares one seed. This package only provides the hashing
// mechanics that consume a seed supplied by the caller.
//
// Dependencies: standard library only. No core packages, no public packages,
// and no functional/internal/eq.
package keyhash
