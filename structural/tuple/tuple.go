package tuple

type Tuple2[A, B any] struct {
	First  A
	Second B
}

func Of2[A, B any](a A, b B) Tuple2[A, B] { return Tuple2[A, B]{First: a, Second: b} }

// FromValues2 constructs a Tuple2 from two values.
//
// Deprecated: Use Of2 instead.
func FromValues2[A, B any](a A, b B) Tuple2[A, B] { return Of2(a, b) }

func MapFirst[A, B, C any](t Tuple2[A, B], fn func(A) C) Tuple2[C, B] {
	return Of2(fn(t.First), t.Second)
}

func MapSecond[A, B, C any](t Tuple2[A, B], fn func(B) C) Tuple2[A, C] {
	return Of2(t.First, fn(t.Second))
}

func Swap[A, B any](t Tuple2[A, B]) Tuple2[B, A] { return Of2(t.Second, t.First) }

type Tuple3[A, B, C any] struct {
	First  A
	Second B
	Third  C
}

func Of3[A, B, C any](a A, b B, c C) Tuple3[A, B, C] {
	return Tuple3[A, B, C]{First: a, Second: b, Third: c}
}

func MapFirst3[A, B, C, X any](t Tuple3[A, B, C], fn func(A) X) Tuple3[X, B, C] {
	return Of3(fn(t.First), t.Second, t.Third)
}

func MapSecond3[A, B, C, X any](t Tuple3[A, B, C], fn func(B) X) Tuple3[A, X, C] {
	return Of3(t.First, fn(t.Second), t.Third)
}

func MapThird3[A, B, C, X any](t Tuple3[A, B, C], fn func(C) X) Tuple3[A, B, X] {
	return Of3(t.First, t.Second, fn(t.Third))
}
