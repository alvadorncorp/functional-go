package optics_test

import "testing"

var opticsBenchmarkPerson person

func BenchmarkNestedUpdate(b *testing.B) {
	source := person{Name: "Ana", Address: address{Location: location{City: "Recife"}}}
	lens := personCityLens()
	b.Run("Manual", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			updated := source
			updated.Address.Location.City = "Curitiba"
			opticsBenchmarkPerson = updated
		}
	})
	b.Run("ComposedLens", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			opticsBenchmarkPerson = lens.Set(source, "Curitiba")
		}
	})
}
