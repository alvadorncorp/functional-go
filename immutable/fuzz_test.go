package immutable_test

import (
	"cmp"
	"maps"
	"slices"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
)

// maxFuzzOps bounds total instructions, including builder headers and edits.
// Every published version is retained and checked after every edit/build;
// bounding instructions, not sampling history, keeps that work feasible.
const maxFuzzOps = 200

// A builder header's last byte selects 1..16 edits from subsequent instructions.
// Edits also count against maxFuzzOps; truncated batches still Build.
const maxFuzzBuilderOps = 16

// Instructions consume an opcode and two operands. A builder header uses
// operand one to select a retained version and operand two for its edit count.
const listOpWidth = 3

// FuzzListOps checks pure and retained-version builder branches against copied
// native slices. Every historical List must survive each later edit and Build.
func FuzzListOps(f *testing.F) {
	f.Add([]byte{0x00, 5, 0, 0x00, 7, 0, 0x01, 0, 9, 0x02, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x02, 0, 0, 0x01, 0, 0})
	f.Add([]byte{0x00, 1, 0, 0x00, 2, 0, 0x00, 3, 0, 0x02, 0, 0, 0x02, 0, 0})
	// Empty builder, mixed edits, then a branch from an earlier published Build.
	f.Add([]byte{3, 0, 3, 0, 0, 7, 0, 0, 8, 1, 0, 9, 2, 1, 0, 3, 1, 1, 0, 0, 6, 1, 0, 5})
	// Cross the vector tail32 boundary, edit/delete it, and branch from before Build.
	var tail32 []byte
	for i := range 32 {
		tail32 = append(tail32, 0, 0, byte(i))
	}
	tail32 = append(tail32, 3, 32, 3, 0, 0, 99, 1, 32, 98, 2, 32, 0, 1, 0, 97,
		3, 32, 1, 2, 31, 0, 0, 0, 96, 0, 0, 95)
	f.Add(tail32)

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			list  immutable.List[int]
			model []int
		}

		var model []int
		current := immutable.ListOf[int]()
		versions := []version{{list: current, model: slices.Clone(model)}}
		checkVersions := func() {
			for i, v := range versions {
				if got := v.list.Items(); !slices.Equal(got, v.model) || v.list.Len() != len(v.model) {
					t.Fatalf("retained List[%d] = %v (Len %d), want %v", i, got, v.list.Len(), v.model)
				}
			}
		}

		ops := 0
		for cur := 0; cur+listOpWidth <= len(data) && ops < maxFuzzOps; cur += listOpWidth {
			ops++
			op, a, v := data[cur]%4, int(data[cur+1]), int(data[cur+2])

			switch op {
			case 0: // append
				current = current.Append(v)
				model = append(model, v)
			case 1: // set
				if len(model) > 0 {
					idx := a % len(model)
					current = current.Set(idx, v)
					model[idx] = v
				}
			case 2: // delete
				if len(model) > 0 {
					idx := a % len(model)
					current = current.Delete(idx)
					model = slices.Delete(model, idx, idx+1)
				}
			case 3: // builder branch
				base := versions[a%len(versions)]
				builder := base.list.Builder()
				model = slices.Clone(base.model)
				for range 1 + v%maxFuzzBuilderOps {
					if cur+2*listOpWidth > len(data) || ops >= maxFuzzOps {
						break
					}
					cur += listOpWidth
					ops++
					edit, index, value := data[cur]%3, int(data[cur+1]), int(data[cur+2])
					switch edit {
					case 0:
						builder.Append(value)
						model = append(model, value)
					case 1:
						if len(model) > 0 {
							idx := index % len(model)
							builder.Set(idx, value)
							model[idx] = value
						}
					case 2:
						if len(model) > 0 {
							idx := index % len(model)
							builder.Delete(idx)
							model = slices.Delete(model, idx, idx+1)
						}
					}
					if builder.Len() != len(model) {
						t.Fatalf("List builder Len = %d, want %d", builder.Len(), len(model))
					}
					for i, want := range model {
						if got := builder.At(i); got != want {
							t.Fatalf("List builder At(%d) = %d, want %d", i, got, want)
						}
					}
					checkVersions()
				}
				current = builder.Build()
			}

			versions = append(versions, version{list: current, model: slices.Clone(model)})
			checkVersions()
		}
		checkVersions()
	})
}

// Map instructions use the same width/header format as List instructions.
// Pure opcodes are set/delete/update; builder edits are set/delete.
const mapOpWidth = 3

// FuzzMapOps checks pure and retained-version builder branches against copied
// native maps, checking every historical Map after each edit and Build.
func FuzzMapOps(f *testing.F) {
	f.Add([]byte{0x00, 1, 10, 0x00, 2, 20, 0x01, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x01, 5, 0, 0x00, 5, 99})
	f.Add([]byte{0x02, 3, 4, 0x02, 3, 5, 0x01, 3, 0, 0x02, 3, 6})
	// Replace/delete/reinsert in one builder, then branch from a pre-Build version.
	f.Add([]byte{
		0, 1, 10, 0, 2, 20, 3, 2, 3, 0, 1, 11, 1, 2, 0, 0, 2, 22, 0, 3, 30,
		3, 1, 1, 1, 1, 0, 0, 4, 40, 2, 4, 5,
	})
	// Empty and truncated builder batches still publish a checkable version.
	f.Add([]byte{3, 0, 15, 1, 9, 0, 0, 9, 0, 0, 9, 99})
	f.Add([]byte{3, 0, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		type version struct {
			m     immutable.Map[int, int]
			model map[int]int
		}

		model := map[int]int{}
		current := immutable.MapOf[int, int](nil)
		versions := []version{{m: current, model: maps.Clone(model)}}
		checkVersions := func() {
			for i, v := range versions {
				if got := v.m.Items(); !maps.Equal(got, v.model) || v.m.Len() != len(v.model) {
					t.Fatalf("retained Map[%d] = %v (Len %d), want %v", i, got, v.m.Len(), v.model)
				}
			}
		}

		ops := 0
		for cur := 0; cur+mapOpWidth <= len(data) && ops < maxFuzzOps; cur += mapOpWidth {
			ops++
			op, k, v := data[cur]%4, int(data[cur+1]), int(data[cur+2])

			switch op {
			case 0: // set
				current = current.Set(k, v)
				model[k] = v
			case 1: // delete
				current = current.Delete(k)
				delete(model, k)
			case 2: // update
				current = fuzzMapUpdate(t, current, model, k, v)
			case 3: // builder branch
				base := versions[k%len(versions)]
				builder := base.m.Builder()
				model = maps.Clone(base.model)
				for range 1 + v%maxFuzzBuilderOps {
					if cur+2*mapOpWidth > len(data) || ops >= maxFuzzOps {
						break
					}
					cur += mapOpWidth
					ops++
					key, value := int(data[cur+1]), int(data[cur+2])
					if data[cur]%2 == 0 {
						builder.Set(key, value)
						model[key] = value
					} else {
						builder.Delete(key)
						delete(model, key)
					}
					if builder.Len() != len(model) {
						t.Fatalf("Map builder Len = %d, want %d", builder.Len(), len(model))
					}
					for key, want := range model {
						if got, ok := builder.Get(key); !ok || got != want {
							t.Fatalf("Map builder Get(%d) = (%d, %v), want (%d, true)", key, got, ok, want)
						}
					}
					checkVersions()
				}
				current = builder.Build()
			}

			versions = append(versions, version{m: current, model: maps.Clone(model)})
			checkVersions()
		}

		checkVersions()
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

// FuzzSortedMapOps checks pure and retained-version builder branches against
// copied native maps, including stdlib-sorted iteration in both comparator
// orders. Every historical version is checked after each edit and Build.
func FuzzSortedMapOps(f *testing.F) {
	f.Add([]byte{0x00, 1, 10, 0x00, 2, 20, 0x01, 1, 0})
	f.Add([]byte{})
	f.Add([]byte{0x01, 5, 0, 0x00, 5, 99})
	f.Add([]byte{0x00, 3, 4, 0x00, 1, 5, 0x00, 2, 6, 0x01, 3, 0, 0x00, 3, 7})
	// Mixed builder edits and branching share the Map seed's independent oracle.
	f.Add([]byte{
		0, 1, 10, 0, 2, 20, 2, 2, 3, 0, 1, 11, 1, 2, 0, 0, 2, 22, 0, 3, 30,
		2, 1, 1, 1, 1, 0, 0, 4, 40, 0, 4, 45,
	})
	f.Add([]byte{2, 0, 15, 1, 9, 0, 0, 9, 0, 0, 9, 99})
	f.Add([]byte{2, 0, 0})

	comparators := map[string]func(a, b int) int{"forward": cmp.Compare[int], "reversed": reversed}

	f.Fuzz(func(t *testing.T, data []byte) {
		for name, order := range comparators {
			type version struct {
				m     immutable.SortedMap[int, int]
				model map[int]int
			}

			model := map[int]int{}
			current := immutable.EmptySortedMap[int, int](order)
			versions := []version{{m: current, model: maps.Clone(model)}}
			checkVersions := func() {
				for i, v := range versions {
					wantKeys := slices.SortedFunc(maps.Keys(v.model), order)
					var gotKeys []int
					for k, val := range v.m.All() {
						if want, ok := v.model[k]; !ok || val != want {
							t.Fatalf("%s: version[%d] maps %d to %d, want (%d, %v)", name, i, k, val, want, ok)
						}
						gotKeys = append(gotKeys, k)
					}
					if !slices.Equal(gotKeys, wantKeys) || v.m.Len() != len(wantKeys) {
						t.Fatalf("%s: version[%d] keys = %v (Len %d), want %v", name, i, gotKeys, v.m.Len(), wantKeys)
					}
				}
			}

			ops := 0
			for cur := 0; cur+mapOpWidth <= len(data) && ops < maxFuzzOps; cur += mapOpWidth {
				ops++
				op, k, v := data[cur]%3, int(data[cur+1]), int(data[cur+2])
				switch op {
				case 0:
					current = current.Set(k, v)
					model[k] = v
				case 1:
					current = current.Delete(k)
					delete(model, k)
				case 2: // builder branch
					base := versions[k%len(versions)]
					builder := base.m.Builder()
					model = maps.Clone(base.model)
					for range 1 + v%maxFuzzBuilderOps {
						if cur+2*mapOpWidth > len(data) || ops >= maxFuzzOps {
							break
						}
						cur += mapOpWidth
						ops++
						key, value := int(data[cur+1]), int(data[cur+2])
						if data[cur]%2 == 0 {
							builder.Set(key, value)
							model[key] = value
						} else {
							builder.Delete(key)
							delete(model, key)
						}
						if builder.Len() != len(model) {
							t.Fatalf("%s: SortedMap builder Len = %d, want %d", name, builder.Len(), len(model))
						}
						for key, want := range model {
							if got, ok := builder.Get(key); !ok || got != want {
								t.Fatalf("%s: SortedMap builder Get(%d) = (%d, %v), want (%d, true)", name, key, got, ok, want)
							}
						}
						checkVersions()
					}
					current = builder.Build()
				}
				versions = append(versions, version{m: current, model: maps.Clone(model)})
				checkVersions()
			}

			checkVersions()
		}
	})
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
