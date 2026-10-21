package immutable_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// maxFuzzOps bounds how many instructions a single fuzz input can decode
// into, so a large corpus entry cannot turn one fuzz iteration into an
// unbounded amount of work.
const maxFuzzOps = 200

// listOpWidth is how many bytes each List instruction consumes from the
// fuzz input: one opcode byte plus two operand bytes (index/value, whose
// meaning depends on the opcode).
const listOpWidth = 3

// FuzzListOps interprets the fuzz input as a small program of List
// operations (append/set/delete), applies it in lockstep to both a
// immutable.List[int] and a native []int model, and keeps every
// intermediate List value alongside the model snapshot taken at that same
// step. The final List must match the final model, and — this is the part
// that actually proves old versions survive deriving new ones — every
// retained List must still match the model snapshot recorded when it was
// produced, even after later operations changed the model further.
func FuzzListOps(f *testing.F) {
	f.Add([]byte{0x00, 5, 0, 0x00, 7, 0, 0x01, 0, 9, 0x02, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x02, 0, 0, 0x01, 0, 0})
	f.Add([]byte{0x00, 1, 0, 0x00, 2, 0, 0x00, 3, 0, 0x02, 0, 0, 0x02, 0, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			list  immutable.List[int]
			model []int
		}

		var model []int
		current := immutable.ListOf[int]()
		versions := []version{{list: current, model: slices.Clone(model)}}

		ops := 0
		for cur := 0; cur+listOpWidth <= len(data) && ops < maxFuzzOps; cur += listOpWidth {
			ops++
			op, a, v := data[cur]%3, int(data[cur+1]), int(data[cur+2])

			switch op {
			case 0: // append
				current = current.Append(v)
				model = append(model, v)
			case 1: // set
				if len(model) == 0 {
					continue
				}
				idx := a % len(model)
				current = current.Set(idx, v)
				model[idx] = v
			case 2: // delete
				if len(model) == 0 {
					continue
				}
				idx := a % len(model)
				current = current.Delete(idx)
				model = append(model[:idx], model[idx+1:]...)
			}

			versions = append(versions, version{list: current, model: slices.Clone(model)})
		}

		if got := current.Items(); !slices.Equal(got, model) {
			t.Fatalf("final List = %v, want %v (model)", got, model)
		}
		for i, v := range versions {
			if got := v.list.Items(); !slices.Equal(got, v.model) {
				t.Fatalf(
					"retained version[%d] = %v, want %v (model snapshot from the step that produced it); "+
						"a later operation must not have reached back into an earlier version",
					i, got, v.model,
				)
			}
		}
	})
}

// mapOpWidth is how many bytes each Map instruction consumes from the fuzz
// input: one opcode byte plus two operand bytes (key, value — value unused
// by delete; update adds it to the old value).
const mapOpWidth = 3

// FuzzMapOps is FuzzListOps' Map analogue: a small program of
// set/delete/update operations run in lockstep against a
// immutable.Map[int, int] and a native map[int]int model, keeping every
// intermediate Map alongside the model snapshot from the step that produced
// it.
func FuzzMapOps(f *testing.F) {
	f.Add([]byte{0x00, 1, 10, 0x00, 2, 20, 0x01, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x01, 5, 0, 0x00, 5, 99})
	f.Add([]byte{0x02, 3, 4, 0x02, 3, 5, 0x01, 3, 0, 0x02, 3, 6})

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			m     immutable.Map[int, int]
			model map[int]int
		}

		model := map[int]int{}
		current := immutable.MapOf[int, int](nil)
		versions := []version{{m: current, model: maps.Clone(model)}}

		ops := 0
		for cur := 0; cur+mapOpWidth <= len(data) && ops < maxFuzzOps; cur += mapOpWidth {
			ops++
			op, k, v := data[cur]%3, int(data[cur+1]), int(data[cur+2])

			switch op {
			case 0: // set
				current = current.Set(k, v)
				model[k] = v
			case 1: // delete
				current = current.Delete(k)
				delete(model, k)
			case 2: // update
				current = fuzzMapUpdate(t, current, model, k, v)
			}

			versions = append(versions, version{m: current, model: maps.Clone(model)})
		}

		if got := current.Items(); !maps.Equal(got, model) {
			t.Fatalf("final Map = %v, want %v (model)", got, model)
		}
		for i, v := range versions {
			if got := v.m.Items(); !maps.Equal(got, v.model) {
				t.Fatalf(
					"retained version[%d] = %v, want %v (model snapshot from the step that produced it); "+
						"a later operation must not have reached back into an earlier version",
					i, got, v.model,
				)
			}
		}
	})
}

// fuzzMapUpdate applies Update(k) to m and to model: the stored value gains
// v, and an absent key is stored as v. It fails t unless fn is called exactly
// once with model's (value, present) pair for k.
func fuzzMapUpdate(t *testing.T, m immutable.Map[int, int], model map[int]int, k, v int) immutable.Map[int, int] {
	t.Helper()
	wantOld, wantPresent := model[k]
	calls := 0
	next := m.Update(k, func(old int, present bool) int {
		calls++
		if old != wantOld || present != wantPresent {
			t.Fatalf("Update(%d) fn got (%d, %v), want (%d, %v)", k, old, present, wantOld, wantPresent)
		}
		return old + v
	})
	if calls != 1 {
		t.Fatalf("Update(%d) called fn %d times, want 1", k, calls)
	}
	model[k] = wantOld + v
	return next
}

// Membership bits of the FuzzSetAlgebra model: which operands hold an element.
const (
	inA    byte = 1
	inB    byte = 2
	inBoth      = inA | inB
)

// setPlacements maps a FuzzSetAlgebra input byte, modulo its length, to the
// operands that hold the byte's element; the quotient is the element itself.
var setPlacements = [...]byte{inA, inB, inBoth, inBoth}

// FuzzSetAlgebra decodes the fuzz input into two Set[int] operands and checks
// Union, Intersection, and Difference in both directions against a
// map-backed model. The first byte selects the operands' lineage: even builds
// a and b independently (distinct hash seeds), odd derives b from a (shared
// seed). Each later byte places one element per setPlacements. Both operands
// must still match their model afterwards.
func FuzzSetAlgebra(f *testing.F) {
	f.Add([]byte{0, 4, 9, 14, 19})
	f.Add([]byte{1, 4, 9, 14, 19})
	f.Add([]byte{})
	f.Add([]byte{1, 2, 6, 10, 14})
	f.Add([]byte{0, 0, 4, 8, 12, 1, 5})

	f.Fuzz(func(t *testing.T, data []byte) {
		derived := len(data) > 0 && data[0]%2 == 1
		model := map[int]byte{}
		for i, c := range data {
			if i == 0 || i > maxFuzzOps {
				continue
			}
			n := len(setPlacements)
			model[int(c)/n] |= setPlacements[int(c)%n]
		}
		modelA := modelSetItems(model, func(m byte) bool { return m&inA != 0 })
		modelB := modelSetItems(model, func(m byte) bool { return m&inB != 0 })

		a, b := immutable.SetOf(modelA...), immutable.SetOf(modelB...)
		if derived {
			b = a.Remove(modelA...).Add(modelB...)
		}

		union := modelSetItems(model, func(m byte) bool { return m != 0 })
		inter := modelSetItems(model, func(m byte) bool { return m == inBoth })
		checks := []struct {
			op   string
			got  immutable.Set[int]
			want []int
		}{
			{"a.Union(b)", a.Union(b), union},
			{"b.Union(a)", b.Union(a), union},
			{"a.Intersection(b)", a.Intersection(b), inter},
			{"b.Intersection(a)", b.Intersection(a), inter},
			{"a.Difference(b)", a.Difference(b), modelSetItems(model, func(m byte) bool { return m == inA })},
			{"b.Difference(a)", b.Difference(a), modelSetItems(model, func(m byte) bool { return m == inB })},
		}
		for _, c := range checks {
			if got := sortedItems(c.got); !slices.Equal(got, c.want) {
				t.Fatalf("derived=%v: %s = %v, want %v", derived, c.op, got, c.want)
			}
		}
		if got := sortedItems(a); !slices.Equal(got, modelA) {
			t.Fatalf("derived=%v: operand a changed: got %v, want %v", derived, got, modelA)
		}
		if got := sortedItems(b); !slices.Equal(got, modelB) {
			t.Fatalf("derived=%v: operand b changed: got %v, want %v", derived, got, modelB)
		}
	})
}

// modelSetItems returns, sorted, every element whose membership bits satisfy
// keep.
func modelSetItems(model map[int]byte, keep func(byte) bool) []int {
	var out []int
	for v, m := range model {
		if keep(m) {
			out = append(out, v)
		}
	}
	slices.Sort(out)
	return out
}
