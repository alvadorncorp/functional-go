package change

import "go.alvadorn.org/functional/primitives/function"

// Compose composes changes right to left.
//
// Compose(f, g, h)(state) is equivalent to f(g(h(state))). With no changes,
// Compose returns the identity change.
func Compose[S any](changes ...Change[S]) Change[S] {
	composed := function.ComposeTransforms(toTransforms(changes)...)
	return Change[S](composed)
}

// Pipe composes changes left to right.
//
// Pipe(f, g, h)(state) is equivalent to h(g(f(state))). With no changes, Pipe
// returns the identity change.
func Pipe[S any](changes ...Change[S]) Change[S] {
	piped := function.PipeTransforms(toTransforms(changes)...)
	return Change[S](piped)
}

func toTransforms[S any](changes []Change[S]) []function.Transform[S] {
	transforms := make([]function.Transform[S], len(changes))
	for index, transformation := range changes {
		transforms[index] = function.Transform[S](transformation)
	}
	return transforms
}
