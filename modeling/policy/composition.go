package policy

const panicNilCombiner = "modeling/policy: nil combiner"

// Compose feeds the output of first into second.
func Compose[A, B, C any](first Policy[A, B], second Policy[B, C]) Policy[A, C] {
	requirePolicy(first)
	requirePolicy(second)
	return func(input A) C { return second(first(input)) }
}

// Zip evaluates left and right on the same input, then combines their outputs.
// Evaluation order is left, right, combine.
func Zip[I, A, B, O any](
	left Policy[I, A],
	right Policy[I, B],
	combine func(A, B) O,
) Policy[I, O] {
	requirePolicy(left)
	requirePolicy(right)
	if combine == nil {
		panic(panicNilCombiner)
	}
	return func(input I) O {
		leftOutput := left(input)
		rightOutput := right(input)
		return combine(leftOutput, rightOutput)
	}
}
