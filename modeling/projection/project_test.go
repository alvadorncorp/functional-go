package projection_test

import (
	"fmt"
	"slices"
	"strconv"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/immutable"
	"github.com/alvadorncorp/gopackages/modeling/projection"
)

type projectFunc[P, E any] func(
	P,
	immutable.List[E],
	projection.Projection[P, E],
) P

var _ projectFunc[int, int] = projection.Project[int, int]

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

	got := projection.Project(0, immutable.ListOf(1, 2, 3), project)

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
	if got := projection.Project(42, immutable.List[int]{}, project); got != 42 {
		t.Fatalf("Project empty result = %d, want 42", got)
	}
	if calls != 0 {
		t.Fatalf("Project empty calls = %d, want 0", calls)
	}

	var nilProjection projection.Projection[int, int]
	if got := projection.Project(42, immutable.List[int]{}, nilProjection); got != 42 {
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
	_ = projection.Project(0, immutable.ListOf(1), project)
}

func TestProjectOneEventAndManualSequenceAgree(t *testing.T) {
	t.Parallel()

	one := projection.Project(7, immutable.ListOf(5), decimalProjection)
	if one != 75 {
		t.Fatalf("Project one event = %d, want 75", one)
	}

	events := immutable.ListOf(2, 4, 6)
	manual := decimalProjection(decimalProjection(decimalProjection(1, 2), 4), 6)
	if got := projection.Project(1, events, decimalProjection); got != manual {
		t.Fatalf("Project sequence = %d, manual = %d", got, manual)
	}
}

func TestProjectConcatenationLaw(t *testing.T) {
	t.Parallel()

	left := immutable.ListOf(1, 2)
	right := immutable.ListOf(3, 4)
	joined := left.Append(right.Items()...)

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
	events := immutable.ListOf(
		indexEvent{key: 1, value: 10},
		indexEvent{key: 2, value: 20},
	)
	before := events.Items()

	first := projection.Project(initial, events, projectIndex)
	second := projection.Project(initial, events, projectIndex)

	if !first.Equal(second) {
		t.Fatalf("deterministic Project results differ: %v and %v", first.Items(), second.Items())
	}
	if initial.Len() != 0 {
		t.Fatalf("initial index length = %d, want unchanged 0", initial.Len())
	}
	if !slices.Equal(events.Items(), before) {
		t.Fatalf("Project changed events: got %v, want %v", events.Items(), before)
	}
}

func TestProjectBatchMatchesIncrementalApplication(t *testing.T) {
	t.Parallel()

	events := immutable.ListOf(3, 1, 4)
	batch := projection.Project(0, events, decimalProjection)
	incremental := 0
	for event := range events.Values() {
		incremental = decimalProjection(incremental, event)
	}

	if batch != incremental {
		t.Fatalf("batch Project = %d, incremental = %d", batch, incremental)
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
	events := immutable.ListOf(
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

var projectBenchmarkSizes = [...]int{0, 1, 8, 32, 128, 1024, 8192}

func benchmarkIntEvents(size int) immutable.List[int] {
	events := make([]int, size)
	for index := range events {
		events[index] = index
	}
	return immutable.ListOf(events...)
}

func benchmarkIndexEvents(size int) immutable.List[indexEvent] {
	events := make([]indexEvent, size)
	for index := range events {
		events[index] = indexEvent{key: index, value: index * 10}
	}
	return immutable.ListOf(events...)
}

func sumProjection(total, event int) int {
	return total + event
}

func projectDirectValues(
	initial int,
	events immutable.List[int],
	project projection.Projection[int, int],
) int {
	projected := initial
	for event := range events.Values() {
		projected = project(projected, event)
	}
	return projected
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

func BenchmarkProjectDirectValues(b *testing.B) {
	for _, size := range projectBenchmarkSizes {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			events := benchmarkIntEvents(size)
			b.ReportAllocs()
			for b.Loop() {
				projectSumSink = projectDirectValues(0, events, sumProjection)
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
