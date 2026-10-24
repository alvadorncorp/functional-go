package lazy_test

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alvadorncorp/gopackages/functional/structural/lazy"
)

func TestLazyDefersAndMemoizes(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	l := lazy.New(func() int {
		calls.Add(1)
		return 42
	})

	if got := calls.Load(); got != 0 {
		t.Fatalf("initializer ran before Get: calls=%d", got)
	}

	if got := l.Get(); got != 42 {
		t.Fatalf("first Get: got %d, want 42", got)
	}
	if got := l.Get(); got != 42 {
		t.Fatalf("second Get: got %d, want 42", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls: got %d, want 1", got)
	}
}

func TestLazyConcurrentGetRunsOnce(t *testing.T) {
	t.Parallel()

	const goroutines = 256
	var calls atomic.Int32
	l := lazy.New(func() int {
		calls.Add(1)
		return 7
	})

	var wg sync.WaitGroup
	wg.Add(goroutines)
	results := make([]int, goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			results[i] = l.Get()
		}(i)
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls: got %d, want 1", got)
	}
	for i, got := range results {
		if got != 7 {
			t.Fatalf("result[%d]: got %d, want 7", i, got)
		}
	}
}

func TestLazyMemoizesPanic(t *testing.T) {
	t.Parallel()

	const goroutines = 256
	payload := "boom"
	var calls atomic.Int32
	l := lazy.New(func() int {
		calls.Add(1)
		panic(payload)
	})

	var wg sync.WaitGroup
	wg.Add(goroutines)
	panics := make([]any, goroutines)
	for i := range goroutines {
		go func(i int) {
			defer wg.Done()
			defer func() {
				panics[i] = recover()
			}()
			_ = l.Get()
		}(i)
	}
	wg.Wait()

	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls: got %d, want 1", got)
	}
	for i, got := range panics {
		if got != payload {
			t.Fatalf("panic[%d]: got %#v, want %#v", i, got, payload)
		}
	}
}

func TestLazyMemoizesGoexitPanic(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	l := lazy.New(func() int {
		calls.Add(1)
		runtime.Goexit()
		return 42
	})

	firstPanic := make(chan any, 1)
	go func() {
		defer func() {
			firstPanic <- recover()
		}()
		_ = l.Get()
	}()
	payload := <-firstPanic
	if _, ok := payload.(*runtime.PanicNilError); !ok {
		t.Fatalf("first Get panic: got %T, want *runtime.PanicNilError", payload)
	}

	for i := range 2 {
		func() {
			defer func() {
				got := recover()
				if _, ok := got.(*runtime.PanicNilError); !ok {
					t.Fatalf("later Get %d panic: got %T, want *runtime.PanicNilError", i, got)
				}
			}()
			_ = l.Get()
			t.Fatalf("later Get %d returned after initializer called Goexit", i)
		}()
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls: got %d, want 1", got)
	}
}

func TestNewRejectsNilInitializer(t *testing.T) {
	t.Parallel()

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("New(nil) did not panic")
		}
		got, ok := recovered.(string)
		if !ok {
			t.Fatalf("panic type: got %T, want string", recovered)
		}
		const want = "functional/lazy: nil initializer"
		if got != want {
			t.Fatalf("panic message: got %q, want %q", got, want)
		}
	}()
	_ = lazy.New[int](nil)
}
