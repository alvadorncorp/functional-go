package optics

const (
	nilGetterPanic = "functional/optics: nil getter"
	nilSetterPanic = "functional/optics: nil setter"
)

// Lens focuses on a part A of a whole S with immutable updates.
// The zero value is invalid and must not be used; construct lenses with NewLens.
type Lens[S, A any] struct {
	get func(S) A
	set func(S, A) S
}

// NewLens builds a Lens from getter and setter callbacks.
// Lawful callbacks must obey Get-Set, Set-Get, and Set-Set; NewLens does not
// check those laws at runtime.
// NewLens panics with "functional/optics: nil getter" if get is nil, and with
// "functional/optics: nil setter" if set is nil.
func NewLens[S, A any](get func(S) A, set func(S, A) S) Lens[S, A] {
	if get == nil {
		panic(nilGetterPanic)
	}
	if set == nil {
		panic(nilSetterPanic)
	}
	return Lens[S, A]{get: get, set: set}
}

// Get returns the focused value from source.
func (lens Lens[S, A]) Get(source S) A { return lens.get(source) }

// Set returns a copy of source with the focus replaced by value.
func (lens Lens[S, A]) Set(source S, value A) S { return lens.set(source, value) }

// Modify applies fn to the focused value and returns the updated source.
// Modify does not check whether fn is nil.
func (lens Lens[S, A]) Modify(source S, fn func(A) A) S {
	return lens.Set(source, fn(lens.Get(source)))
}
