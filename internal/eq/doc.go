// Package eq resolves value equality for type parameters that carry no
// comparability constraint. It is shared by every functional subpackage that
// needs an Equal method on a generic type, so all of them agree on the answer
// for the same pair of values.
package eq
