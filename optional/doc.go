// Package optional provides Optional[T], a value type for an operation that
// may or may not produce a T.
//
// The zero Optional is a valid None, so a struct field of type Optional[T]
// needs no constructor or nil check. Some(zero) and None are distinct:
// presence is tracked apart from the value, never inferred by comparing it
// to the zero value.
//
// Methods keep the result type T: IsSome, Get, OrElse, Filter, Equal, and the
// rest. Every operation that changes the result type — Map, FlatMap, MapErr,
// Zip, EqualFunc — is a package-level function instead, because Go does not
// allow type parameters on methods.
//
// Optional guards T inline rather than behind a pointer: constructing a Some
// never allocates, but the guarded value is a copy, and immutability is only
// as deep as T's own. Equality, both Optional.Equal and the == operator when
// T is comparable, is resolved through the sibling internal package
// functional/internal/eq, so Equal works even when T carries no
// comparability constraint.
package optional
