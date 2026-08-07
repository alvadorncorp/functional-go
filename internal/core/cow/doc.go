// Package cow orchestrates the copy-on-write mutability decision without
// performing any structure-specific clone.
//
// Core answers only: may this value be mutated in place under the current
// owner, or must the caller copy first? The concrete structure supplies the
// clone function; this package never inspects node layout and never uses
// reflection.
package cow
