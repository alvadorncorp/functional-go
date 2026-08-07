// Package functional is an umbrella module for small, dependency-free
// functional-programming building blocks; this file is the entire root
// package, which exports nothing.
//
// Each concept lives in its own subpackage: optional (Optional[T] and its
// combinators), frozen (immutable snapshots — Slice and Map — with no write
// operation), persistent (collections with functional writes; full-copy
// today, with Builders over an internal transient window), and valueobject
// (reflective verification of structural immutability, a test helper).
//
// The module has zero dependencies outside the standard library, permanently.
// No public subpackage imports another public subpackage. Shared internals
// include functional/internal/eq (Equal resolution: method, then ==, then
// reflect.DeepEqual) and functional/internal/core (thin ownership / cow /
// sharing / transient substrate used by persistent Builders; not a public
// API and not a HAMT/RRB engine).
//
// The exported type is always Optional, never Option: Option already names
// the functional-options pattern elsewhere in this repository (eventbus,
// workopher).
package functional
