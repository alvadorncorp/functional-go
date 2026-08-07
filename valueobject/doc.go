// Package valueobject reflectively verifies that a Go type follows the
// structural convention of an immutable value object: no field a caller
// outside the package can use to mutate an already-constructed value.
//
// This is a test helper, not a runtime guard. It proves structural
// immutability — nothing external can reach in and mutate a value — never
// semantic correctness: it has no way to know whether a constructor
// validates its input, or whether a With* method actually copies before it
// writes. Go gives no way to embed "unexported fields", "a validating
// constructor", or "copy-on-write With* methods" as a reusable library
// contract; those remain decisions each consumer makes for its own type.
// What Inspect checks mechanically is narrower and still useful: it walks
// T's fields and reports every one an external caller could use to reach in
// and mutate the value — an exported field, or a pointer/slice/map/chan/func
// alias, or an interface whose dynamic type cannot be proven immutable by
// structural inspection alone.
//
// time.Time, time.Duration, and any type declared in functional/persistent
// are treated as immutable leaves by default: those collections already
// guarantee structural immutability of the collection itself. AllowType and
// AllowPackage extend that leaf set for a type this package has no other
// way to know is safe.
package valueobject
