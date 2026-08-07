package persistent_test

import (
	"maps"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
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
// persistent.List[int] and a native []int model, and keeps every
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
			list  persistent.List[int]
			model []int
		}

		var model []int
		current := persistent.NewList[int]()
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
// by delete).
const mapOpWidth = 3

// FuzzMapOps is FuzzListOps' Map analogue: a small program of set/delete
// operations run in lockstep against a persistent.Map[int, int] and a
// native map[int]int model, keeping every intermediate Map alongside the
// model snapshot from the step that produced it.
func FuzzMapOps(f *testing.F) {
	f.Add([]byte{0x00, 1, 10, 0x00, 2, 20, 0x01, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x01, 5, 0, 0x00, 5, 99})

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			m     persistent.Map[int, int]
			model map[int]int
		}

		model := map[int]int{}
		current := persistent.NewMap[int, int](nil)
		versions := []version{{m: current, model: maps.Clone(model)}}

		ops := 0
		for cur := 0; cur+mapOpWidth <= len(data) && ops < maxFuzzOps; cur += mapOpWidth {
			ops++
			op, k, v := data[cur]%2, int(data[cur+1]), int(data[cur+2])

			switch op {
			case 0: // set
				current = current.Set(k, v)
				model[k] = v
			case 1: // delete
				current = current.Delete(k)
				delete(model, k)
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
