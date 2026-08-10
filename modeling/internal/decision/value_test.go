package decision

import (
	"slices"
	"strconv"
	"strings"
	"testing"
)

func TestValueZeroAndNew(t *testing.T) {
	var zero Value[string, int]
	if got := zero.State(); got != "" {
		t.Fatalf("zero State() = %q, want empty", got)
	}
	if got := zero.EventCount(); got != 0 {
		t.Fatalf("zero EventCount() = %d, want 0", got)
	}
	if got := slices.Collect(zero.Events()); len(got) != 0 {
		t.Fatalf("zero Events() = %v, want empty", got)
	}

	value := New[int]("open")
	if got := value.State(); got != "open" {
		t.Fatalf("State() = %q, want open", got)
	}
	if got := value.EventCount(); got != 0 {
		t.Fatalf("EventCount() = %d, want 0", got)
	}
}

func TestValueAppendPreservesReceiverAndOrder(t *testing.T) {
	original := New[int]("open")
	next := original.Append(1, 2, 3)

	if got := original.EventCount(); got != 0 {
		t.Fatalf("original EventCount() = %d, want 0", got)
	}
	if got := next.State(); got != "open" {
		t.Fatalf("next State() = %q, want open", got)
	}
	if got := slices.Collect(next.Events()); !slices.Equal(got, []int{1, 2, 3}) {
		t.Fatalf("next Events() = %v, want [1 2 3]", got)
	}
	if got := next.Append().EventCount(); got != 3 {
		t.Fatalf("empty Append EventCount() = %d, want 3", got)
	}
}

func TestMapStateTransformsStateAndPreservesEvents(t *testing.T) {
	source := New[int](2).Append(3, 5)
	mapperCalls := 0
	mapped := MapState(source, func(state int) string {
		mapperCalls++
		return strings.Repeat("x", state)
	})

	if got := mapped.State(); got != "xx" {
		t.Fatalf("mapped State() = %q, want xx", got)
	}
	if mapperCalls != 1 {
		t.Fatalf("MapState mapper calls = %d, want 1", mapperCalls)
	}
	if got := slices.Collect(mapped.Events()); !slices.Equal(got, []int{3, 5}) {
		t.Fatalf("mapped Events() = %v, want [3 5]", got)
	}
	if got := source.State(); got != 2 {
		t.Fatalf("source State() = %d, want 2", got)
	}

	identity := MapState(source, func(state int) int { return state })
	composed := MapState(
		MapState(source, func(state int) int { return state + 1 }),
		func(state int) int { return state * 2 },
	)
	once := MapState(source, func(state int) int { return (state + 1) * 2 })
	if identity.State() != source.State() || composed.State() != once.State() {
		t.Fatalf(
			"MapState laws failed: identity=%d source=%d composed=%d once=%d",
			identity.State(), source.State(), composed.State(), once.State(),
		)
	}
}

func TestMapEventsTransformsInOrderAndPreservesState(t *testing.T) {
	source := New[int]("open").Append(2, 3)
	mapperCalls := 0
	mapped := MapEvents(source, func(event int) string {
		mapperCalls++
		return strconv.Itoa(event)
	})

	if got := mapped.State(); got != "open" {
		t.Fatalf("mapped State() = %q, want open", got)
	}
	if got := slices.Collect(mapped.Events()); !slices.Equal(got, []string{"2", "3"}) {
		t.Fatalf("mapped Events() = %v, want [2 3]", got)
	}
	if mapperCalls != source.EventCount() {
		t.Fatalf("MapEvents mapper calls = %d, want %d", mapperCalls, source.EventCount())
	}
	if got := slices.Collect(source.Events()); !slices.Equal(got, []int{2, 3}) {
		t.Fatalf("source Events() = %v, want [2 3]", got)
	}

	identity := MapEvents(source, func(event int) int { return event })
	composed := MapEvents(MapEvents(source, func(event int) int { return event + 1 }), strconv.Itoa)
	once := MapEvents(source, func(event int) string { return strconv.Itoa(event + 1) })
	if !slices.Equal(slices.Collect(identity.Events()), slices.Collect(source.Events())) ||
		!slices.Equal(slices.Collect(composed.Events()), slices.Collect(once.Events())) {
		t.Fatal("MapEvents identity or composition law failed")
	}

	emptyCalls := 0
	empty := MapEvents(New[int]("empty"), func(event int) string {
		emptyCalls++
		return strconv.Itoa(event)
	})
	if empty.State() != "empty" || empty.EventCount() != 0 {
		t.Fatalf("empty map = (%q, %d), want (empty, 0)", empty.State(), empty.EventCount())
	}
	if emptyCalls != 0 {
		t.Fatalf("empty MapEvents mapper calls = %d, want 0", emptyCalls)
	}
}

func TestCombineUsesRightStateAndConcatenatesEvents(t *testing.T) {
	left := New[int]("left").Append(1, 2)
	right := New[int]("right").Append(3, 4)
	combined := Combine(left, right)

	if got := combined.State(); got != "right" {
		t.Fatalf("combined State() = %q, want right", got)
	}
	if got := slices.Collect(combined.Events()); !slices.Equal(got, []int{1, 2, 3, 4}) {
		t.Fatalf("combined Events() = %v, want [1 2 3 4]", got)
	}
	if !slices.Equal(slices.Collect(left.Events()), []int{1, 2}) ||
		!slices.Equal(slices.Collect(right.Events()), []int{3, 4}) {
		t.Fatal("Combine changed an input")
	}
}

func TestCombineHandlesEmptySidesAndIsAssociative(t *testing.T) {
	emptyLeft := Combine(New[int]("left"), New[int]("right").Append(1))
	if emptyLeft.State() != "right" || !slices.Equal(slices.Collect(emptyLeft.Events()), []int{1}) {
		t.Fatal("Combine with empty left changed the right value")
	}

	emptyRight := Combine(New[int]("left").Append(1), New[int]("right"))
	if emptyRight.State() != "right" || !slices.Equal(slices.Collect(emptyRight.Events()), []int{1}) {
		t.Fatal("Combine with empty right lost left events or right state")
	}

	a := New[int]("a").Append(1)
	b := New[int]("b").Append(2)
	c := New[int]("c").Append(3)
	leftAssociated := Combine(Combine(a, b), c)
	rightAssociated := Combine(a, Combine(b, c))
	if leftAssociated.State() != rightAssociated.State() ||
		!slices.Equal(slices.Collect(leftAssociated.Events()), slices.Collect(rightAssociated.Events())) {
		t.Fatal("Combine is not observationally associative")
	}
}
