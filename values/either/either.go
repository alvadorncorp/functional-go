package either

// Either is a neutral disjoint union whose zero value is Left(zero(L)).
type Either[L any, R any] struct {
	left    L
	right   R
	isRight bool
}

// Left returns an Either containing value in its left branch.
func Left[L, R any](value L) Either[L, R] { return Either[L, R]{left: value} }

// Right returns an Either containing value in its right branch.
func Right[L, R any](value R) Either[L, R] { return Either[L, R]{right: value, isRight: true} }

// IsLeft reports whether e contains a left value.
func (e Either[L, R]) IsLeft() bool { return !e.isRight }

// IsRight reports whether e contains a right value.
func (e Either[L, R]) IsRight() bool { return e.isRight }

// LeftValue returns the left value and whether e is in its left branch.
func (e Either[L, R]) LeftValue() (L, bool) { return e.left, !e.isRight }

// RightValue returns the right value and whether e is in its right branch.
func (e Either[L, R]) RightValue() (R, bool) { return e.right, e.isRight }

// MapRight transforms a right value or preserves a left value.
func MapRight[L, R, U any](e Either[L, R], transform func(R) U) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return Right[L](transform(e.right))
}

// MapLeft transforms a left value or preserves a right value.
func MapLeft[L, R, U any](e Either[L, R], transform func(L) U) Either[U, R] {
	if e.isRight {
		return Right[U](e.right)
	}
	return Left[U, R](transform(e.left))
}

// Bimap transforms the value in the branch e contains.
func Bimap[L, R, L2, R2 any](e Either[L, R], leftFn func(L) L2, rightFn func(R) R2) Either[L2, R2] {
	if e.isRight {
		return Right[L2](rightFn(e.right))
	}
	return Left[L2, R2](leftFn(e.left))
}

// FlatMapRight transforms a right value or preserves a left value.
func FlatMapRight[L, R, U any](e Either[L, R], transform func(R) Either[L, U]) Either[L, U] {
	if !e.isRight {
		return Left[L, U](e.left)
	}
	return transform(e.right)
}

// Match calls the function for the branch e contains and returns its result.
func Match[L, R, T any](e Either[L, R], leftFn func(L) T, rightFn func(R) T) T {
	if e.isRight {
		return rightFn(e.right)
	}
	return leftFn(e.left)
}
