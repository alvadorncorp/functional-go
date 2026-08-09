// Package functional is an umbrella module for small, dependency-free
// functional-programming building blocks; this file is the entire root
// package, which exports nothing.
//
// Each concept lives in its own subpackage: optional (Optional[T] and its
// combinators), persistent (immutable collections with functional writes:
// HAMT Map/Set and a 32-way vector List; Builders use an exclusive-owner
// transient window with O(1) Freeze), state (pure semantic evolution over T:
// reverse-linked plans, Version, single-use Transaction, structural Diff),
// and valueobject (reflective verification of structural immutability, a
// test helper).
//
// The nested values family contains independent algebraic value packages:
// values/optional, values/result, values/either, values/nonempty, and
// values/validated. The legacy top-level optional package remains a separate
// public API. Values packages do not use persistent/state/core/engine runtime
// machinery; only values/validated imports values/nonempty.
//
// The module has zero dependencies outside the standard library, permanently.
// No public subpackage imports another public subpackage, except for the
// ADR-007 values/validated -> values/nonempty edge. Shared internals
// include functional/internal/eq (Equal resolution: method, then ==;
// non-comparable types without Equal return false — no reflect),
// functional/internal/core (thin ownership / cow /
// sharing / transient substrate — laws, not engines), and
// functional/internal/engine (hamt, vector, keyhash representations used by
// persistent). Production state does not import core or engines.
//
// The exported type is always Optional, never Option: Option already names
// the functional-options pattern elsewhere in this repository (eventbus,
// workopher).
package functional
