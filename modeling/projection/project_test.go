package projection_test

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"testing"

	"go.alvadorn.org/functional/immutable"
	"go.alvadorn.org/functional/modeling/projection"
)

type projectFunc[P, E any] func(
	P,
	iter.Seq[E],
	projection.Projection[P, E],
) P

var _ projectFunc[int, int] = projection.Project[int, int]

func seqOf[E any](events ...E) iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, event := range events {
			if !yield(event) {
				return
			}
		}
	}
}

func emptySeq[E any]() iter.Seq[E] {
	return func(func(E) bool) {}
}

func concatSeq[E any](left, right iter.Seq[E]) iter.Seq[E] {
	return func(yield func(E) bool) {
		for event := range left {
			if !yield(event) {
				return
			}
		}
		for event := range right {
			if !yield(event) {
				return
			}
		}
	}
}

func decimalProjection(current, event int) int {
	return current*10 + event
}

func TestProjectThreadsRepresentationInEncounterOrder(t *testing.T) {
	t.Parallel()

	var calls []int
	project := func(current, event int) int {
		calls = append(calls, event)
		return decimalProjection(current, event)
	}

	got := projection.Project(0, seqOf(1, 2, 3), project)

	if got != 123 {
		t.Fatalf("Project result = %d, want 123", got)
	}
	if !slices.Equal(calls, []int{1, 2, 3}) {
		t.Fatalf("Project calls = %v, want [1 2 3]", calls)
	}
}

func TestProjectEmptyReturnsInitialWithoutInvocation(t *testing.T) {
	t.Parallel()

	calls := 0
	project := func(current, event int) int {
		calls++
		return current + event
	}
	if got := projection.Project(42, emptySeq[int](), project); got != 42 {
		t.Fatalf("Project empty result = %d, want 42", got)
	}
	if calls != 0 {
		t.Fatalf("Project empty calls = %d, want 0", calls)
	}

	var nilProjection projection.Projection[int, int]
	if got := projection.Project(42, emptySeq[int](), nilProjection); got != 42 {
		t.Fatalf("Project empty nil result = %d, want 42", got)
	}
}

func TestProjectNonEmptyNilProjectionHasOrdinaryFunctionPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("Project with non-empty events and nil Projection did not panic")
		}
	}()

	var project projection.Projection[int, int]
	_ = projection.Project(0, seqOf(1), project)
}

func TestProjectOneEventAndManualSequenceAgree(t *testing.T) {
	t.Parallel()

	one := projection.Project(7, seqOf(5), decimalProjection)
	if one != 75 {
		t.Fatalf("Project one event = %d, want 75", one)
	}

	events := seqOf(2, 4, 6)
	manual := decimalProjection(decimalProjection(decimalProjection(1, 2), 4), 6)
	if got := projection.Project(1, events, decimalProjection); got != manual {
		t.Fatalf("Project sequence = %d, manual = %d", got, manual)
	}
}

func TestProjectConcatenationLaw(t *testing.T) {
	t.Parallel()

	left := seqOf(1, 2)
	right := seqOf(3, 4)
	joined := concatSeq(left, right)

	direct := projection.Project(0, joined, decimalProjection)
	staged := projection.Project(
		projection.Project(0, left, decimalProjection),
		right,
		decimalProjection,
	)

	if direct != staged {
		t.Fatalf("concatenated Project = %d, staged Project = %d", direct, staged)
	}
}

type indexEvent struct {
	key   int
	value int
}

func projectIndex(
	index immutable.Map[int, int],
	event indexEvent,
) immutable.Map[int, int] {
	return index.Set(event.key, event.value)
}

func TestProjectIsDeterministicAndPreservesImmutableInputs(t *testing.T) {
	t.Parallel()

	initial := immutable.Map[int, int]{}
	source := []indexEvent{
		{key: 1, value: 10},
		{key: 2, value: 20},
	}

	first := projection.Project(initial, seqOf(source...), projectIndex)
	second := projection.Project(initial, seqOf(source...), projectIndex)

	if !first.Equal(second) {
		t.Fatalf("deterministic Project results differ: %v and %v", first.Items(), second.Items())
	}
	if initial.Len() != 0 {
		t.Fatalf("initial index length = %d, want unchanged 0", initial.Len())
	}
	if !slices.Equal(source, []indexEvent{{key: 1, value: 10}, {key: 2, value: 20}}) {
		t.Fatalf("Project changed source events: got %v", source)
	}
}

func TestProjectBatchMatchesIncrementalApplication(t *testing.T) {
	t.Parallel()

	events := []int{3, 1, 4}
	batch := projection.Project(0, seqOf(events...), decimalProjection)
	incremental := 0
	for _, event := range events {
		incremental = decimalProjection(incremental, event)
	}

	if batch != incremental {
		t.Fatalf("batch Project = %d, incremental = %d", batch, incremental)
	}
}

func TestProjectIsSinglePass(t *testing.T) {
	t.Parallel()

	passes := 0
	events := func(yield func(int) bool) {
		passes++
		if passes > 1 {
			t.Fatal("Project ranged over events more than once")
		}
		for _, event := range []int{1, 2} {
			if !yield(event) {
				return
			}
		}
	}

	got := projection.Project(0, events, decimalProjection)
	if got != 12 {
		t.Fatalf("Project single-pass result = %d, want 12", got)
	}
	if passes != 1 {
		t.Fatalf("Project passes = %d, want 1", passes)
	}
}

type salesEventKind uint8

const (
	saleOrderCreated salesEventKind = iota + 1
	saleOrderPaid
)

type salesEvent struct {
	kind   salesEventKind
	amount int
}

type projectedSalesSummary struct {
	orders  int
	revenue int
}

func projectSalesSummary(
	summary projectedSalesSummary,
	event salesEvent,
) projectedSalesSummary {
	switch event.kind {
	case saleOrderCreated:
		summary.orders++
	case saleOrderPaid:
		summary.revenue += event.amount
	}
	return summary
}

func ExampleProject() {
	events := seqOf(
		salesEvent{kind: saleOrderCreated},
		salesEvent{kind: saleOrderPaid, amount: 125},
		salesEvent{kind: saleOrderCreated},
	)

	summary := projection.Project(
		projectedSalesSummary{},
		events,
		projectSalesSummary,
	)
	fmt.Printf("orders=%d revenue=%d\n", summary.orders, summary.revenue)

	// Output:
	// orders=2 revenue=125
}

var (
	projectIntSink immutable.Map[int, int]
	projectSumSink int
)

var projectBenchmarkSizes = [...]int{0, 1, 10, 100, 1000, 10000}

func benchmarkIntEvents(size int) iter.Seq[int] {
	events := make([]int, size)
	for index := range events {
		events[index] = index
	}
	return seqOf(events...)
}

func benchmarkIndexEvents(size int) iter.Seq[indexEvent] {
	events := make([]indexEvent, size)
	for index := range events {
		events[index] = indexEvent{key: index, value: index * 10}
	}
	return seqOf(events...)
}

func sumProjection(total, event int) int {
	return total + event
}

func BenchmarkProjectInt(b *testing.B) {
	for _, size := range projectBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := benchmarkIntEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				projectSumSink = projection.Project(0, events, sumProjection)
			}
		})
	}
}

func BenchmarkProjectImmutableMap(b *testing.B) {
	for _, size := range projectBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := benchmarkIndexEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				projectIntSink = projection.Project(
					immutable.Map[int, int]{},
					events,
					projectIndex,
				)
			}
		})
	}
}
