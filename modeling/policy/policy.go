// Package policy provides pure, composable domain policies.
package policy

const panicNilPolicy = "modeling/policy: nil policy"

// Policy derives an output from one explicit input context.
type Policy[I, O any] func(I) O

// Constant returns a Policy that ignores its input and always returns output.
func Constant[I, O any](output O) Policy[I, O] {
	return func(I) O { return output }
}

func requirePolicy[I, O any](p Policy[I, O]) {
	if p == nil {
		panic(panicNilPolicy)
	}
}
