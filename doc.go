// Package functional is an umbrella module for small, dependency-free
// functional-programming building blocks; this file is the entire root
// package, which exports nothing.
//
// Each concept lives in its own subpackage: optional (Optional[T] and its
// combinators), frozen (immutable snapshots — Slice and Map — with no write
// operation), persistent (collections with functional writes: HAMT Map/Set
// and a 32-way vector List; Builders use an exclusive-owner transient window
// with O(1) Freeze), and valueobject (reflective verification of structural
// immutability, a test helper).
//
// The module has zero dependencies outside the standard library, permanently.
// No public subpackage imports another public subpackage. Shared internals
// include functional/internal/eq (Equal resolution: method, then ==;
// non-comparable types without Equal return false — no reflect),
// functional/internal/core (thin ownership / cow /
// sharing / transient substrate — laws, not engines), and
// functional/internal/engine (hamt, vector, keyhash representations used by
// persistent).
//
// The exported type is always Optional, never Option: Option already names
// the functional-options pattern elsewhere in this repository (eventbus,
// workopher).
package functional
