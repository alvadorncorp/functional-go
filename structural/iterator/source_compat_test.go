package iterator

import (
	"iter"
	"slices"
	"testing"
)

func TestSourceAdaptersPreserveLiveView(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		from func([]int) iter.Seq[int]
	}{
		{name: "FromSlice", from: FromSlice[int]},
		{name: "FromValues", from: func(values []int) iter.Seq[int] { return FromValues(values...) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			values := []int{1, 2, 3}
			seq := test.from(values)
			values[0] = 9
			var got []int
			for value := range seq {
				got = append(got, value)
				values[1] = 8
			}
			if !slices.Equal(got, []int{9, 8, 3}) {
				t.Fatalf("live view after mutations before and during iteration = %v", got)
			}
		})
	}
}
