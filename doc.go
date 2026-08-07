// Package functional is an umbrella module for small, dependency-free
// functional-programming building blocks; this file is the entire root
// package, which exports nothing.
//
// Each concept lives in its own subpackage: optional (Optional[T] and its
// combinators), frozen (immutable snapshots — Slice and Map — with no write
// operation), persistent (collections with functional writes, copy-on-write
// today behind a swappable engine), and valueobject (reflective verification
// of structural immutability, a test helper).
//
// The module has zero dependencies outside the standard library, permanently.
// No public subpackage imports another public subpackage, with one
// exception: the internal equality package functional/internal/eq, which
// optional, frozen, and persistent each import to resolve their Equal
// method (an Equal method on the value, then ==, then reflect.DeepEqual) —
// an internal import, not one public subpackage depending on another.
//
// The exported type is always Optional, never Option: Option already names
// the functional-options pattern elsewhere in this repository (eventbus,
// workopher).
package functional
