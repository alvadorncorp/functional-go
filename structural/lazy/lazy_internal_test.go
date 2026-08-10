package lazy

import "testing"

func TestInitializerReleasedAfterSuccessAndPanic(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		l := New(func() int { return 1 })
		if l.initializer == nil {
			t.Fatal("initializer is nil before Get")
		}
		if got := l.Get(); got != 1 {
			t.Fatalf("Get: got %d, want 1", got)
		}
		if l.initializer != nil {
			t.Fatal("initializer retained after successful Get")
		}
	})

	t.Run("panic", func(t *testing.T) {
		t.Parallel()

		l := New(func() int { panic("fail") })
		if l.initializer == nil {
			t.Fatal("initializer is nil before Get")
		}
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("Get did not panic")
				}
			}()
			_ = l.Get()
		}()
		if l.initializer != nil {
			t.Fatal("initializer retained after panicking Get")
		}
	})
}
