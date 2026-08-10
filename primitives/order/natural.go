package order

import "cmp"

// Compare compares two naturally ordered values.
func Compare[T cmp.Ordered](a, b T) int { return cmp.Compare(a, b) }

// Natural returns the natural total order for T.
func Natural[T cmp.Ordered]() Order[T] { return Compare[T] }
