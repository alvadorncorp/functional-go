package modeling_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/modeling"
)

var (
	_ func(
		modeling.Decision[int, string],
		func(int) bool,
	) modeling.Decision[bool, string] = modeling.MapState[int, bool, string]
	_ func(
		modeling.Decision[int, string],
		func(string) bool,
	) modeling.Decision[int, bool] = modeling.MapEvents[int, string, bool]
	_ func(
		modeling.Decision[int, string],
		string,
	) modeling.Decision[int, string] = modeling.Emit[int, string]
	_ func(
		modeling.Decision[int, string],
		...string,
	) modeling.Decision[int, string] = modeling.AppendEvents[int, string]
	_ func(
		modeling.Decision[int, string],
		modeling.Decision[int, string],
	) modeling.Decision[int, string] = modeling.Combine[int, string]
)

func intEvents(decision modeling.Decision[string, int]) []int {
	return slices.Collect(decision.Events())
}

func TestDecisionMapStateLaws(t *testing.T) {
	source := modeling.Decide[int](2).Emit(3, 5)
	mapperCalls := 0
	mapped := modeling.MapState(source, func(state int) string {
		mapperCalls++
		return strconv.Itoa(state)
	})
	if mapped.State() != "2" || !slices.Equal(intEvents(mapped), []int{3, 5}) {
		t.Fatalf("mapped = (%q, %v), want (2, [3 5])", mapped.State(), intEvents(mapped))
	}
	if mapperCalls != 1 {
		t.Fatalf("MapState mapper calls = %d, want 1", mapperCalls)
	}
	if source.State() != 2 || !slices.Equal(slices.Collect(source.Events()), []int{3, 5}) {
		t.Fatal("MapState changed its source")
	}

	identity := modeling.MapState(source, func(state int) int { return state })
	composed := modeling.MapState(
		modeling.MapState(source, func(state int) int { return state + 1 }),
		func(state int) int { return state * 2 },
	)
	once := modeling.MapState(source, func(state int) int { return (state + 1) * 2 })
	if identity.State() != source.State() || composed.State() != once.State() ||
		!slices.Equal(slices.Collect(composed.Events()), slices.Collect(once.Events())) {
		t.Fatal("MapState identity or composition law failed")
	}
}

func TestDecisionMapEventsLaws(t *testing.T) {
	source := modeling.Decide[int]("open").Emit(2, 3)
	mapperCalls := 0
	mapped := modeling.MapEvents(source, func(event int) string {
		mapperCalls++
		return strconv.Itoa(event)
	})
	if mapped.State() != "open" || !slices.Equal(slices.Collect(mapped.Events()), []string{"2", "3"}) {
		t.Fatalf("mapped = (%q, %v), want (open, [2 3])", mapped.State(), slices.Collect(mapped.Events()))
	}
	if mapperCalls != source.EventCount() {
		t.Fatalf("MapEvents mapper calls = %d, want %d", mapperCalls, source.EventCount())
	}
	if !slices.Equal(slices.Collect(source.Events()), []int{2, 3}) {
		t.Fatal("MapEvents changed its source")
	}

	identity := modeling.MapEvents(source, func(event int) int { return event })
	composed := modeling.MapEvents(modeling.MapEvents(source, func(event int) int { return event + 1 }), strconv.Itoa)
	once := modeling.MapEvents(source, func(event int) string { return strconv.Itoa(event + 1) })
	if !slices.Equal(slices.Collect(identity.Events()), slices.Collect(source.Events())) ||
		!slices.Equal(slices.Collect(composed.Events()), slices.Collect(once.Events())) {
		t.Fatal("MapEvents identity or composition law failed")
	}

	emptyCalls := 0
	empty := modeling.MapEvents(modeling.Decide[int]("empty"), func(event int) string {
		emptyCalls++
		return strconv.Itoa(event)
	})
	if empty.State() != "empty" || empty.EventCount() != 0 {
		t.Fatal("MapEvents changed an empty decision")
	}
	if emptyCalls != 0 {
		t.Fatalf("empty MapEvents mapper calls = %d, want 0", emptyCalls)
	}
}

func TestDecisionEmitFunction(t *testing.T) {
	source := modeling.Decide[int]("open").Emit(1)
	next := modeling.Emit(source, 2)
	if next.State() != "open" || !slices.Equal(intEvents(next), []int{1, 2}) {
		t.Fatalf("next = (%q, %v), want (open, [1 2])", next.State(), intEvents(next))
	}
	if !slices.Equal(intEvents(source), []int{1}) {
		t.Fatal("Emit changed its source")
	}
}

func TestDecisionAppendEvents(t *testing.T) {
	source := modeling.Decide[int]("open").Emit(1)
	unchanged := modeling.AppendEvents(source)
	next := modeling.AppendEvents(source, 2, 3)
	if !slices.Equal(intEvents(unchanged), []int{1}) {
		t.Fatalf("empty AppendEvents = %v, want [1]", intEvents(unchanged))
	}
	if next.State() != "open" || !slices.Equal(intEvents(next), []int{1, 2, 3}) {
		t.Fatalf("next = (%q, %v), want (open, [1 2 3])", next.State(), intEvents(next))
	}
	if !slices.Equal(intEvents(source), []int{1}) {
		t.Fatal("AppendEvents changed its source")
	}
}

func TestDecisionCombine(t *testing.T) {
	left := modeling.Decide[int]("left").Emit(1, 2)
	right := modeling.Decide[int]("right").Emit(3, 4)
	combined := modeling.Combine(left, right)
	if combined.State() != "right" || !slices.Equal(intEvents(combined), []int{1, 2, 3, 4}) {
		t.Fatalf("combined = (%q, %v), want (right, [1 2 3 4])", combined.State(), intEvents(combined))
	}
	if !slices.Equal(intEvents(left), []int{1, 2}) || !slices.Equal(intEvents(right), []int{3, 4}) {
		t.Fatal("Combine changed an input")
	}

	emptyLeft := modeling.Combine(modeling.Decide[int]("left"), right)
	emptyRight := modeling.Combine(left, modeling.Decide[int]("right"))
	if !slices.Equal(intEvents(emptyLeft), []int{3, 4}) || emptyLeft.State() != "right" {
		t.Fatal("Combine with empty left changed the right decision")
	}
	if !slices.Equal(intEvents(emptyRight), []int{1, 2}) || emptyRight.State() != "right" {
		t.Fatal("Combine with empty right lost left events or right state")
	}

	a := modeling.Decide[int]("a").Emit(1)
	b := modeling.Decide[int]("b").Emit(2)
	c := modeling.Decide[int]("c").Emit(3)
	leftAssociated := modeling.Combine(modeling.Combine(a, b), c)
	rightAssociated := modeling.Combine(a, modeling.Combine(b, c))
	if leftAssociated.State() != rightAssociated.State() ||
		!slices.Equal(intEvents(leftAssociated), intEvents(rightAssociated)) {
		t.Fatal("Combine is not observationally associative")
	}
}
