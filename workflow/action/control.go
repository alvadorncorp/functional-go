package action

import "context"

// Recover replaces a failure with the Action recovery returns, leaving a
// successful result untouched.
//
// recovery runs only when action fails, and it receives that error unchanged —
// including context.Canceled and context.DeadlineExceeded, which this package
// never reclassifies. The Action recovery returns runs with the same context.
// Recover handles errors only: a panic raised inside action is a programming
// defect and propagates untouched, because nothing in this package calls
// recover(). Recover panics with "workflow/action: nil action" or
// "workflow/action: nil recovery" at construction, and with
// "workflow/action: nil action" at execution when recovery returns nil.
func Recover[T any](action Action[T], recovery func(error) Action[T]) Action[T] {
	if action == nil {
		panic(panicNilAction)
	}
	if recovery == nil {
		panic(panicNilRecovery)
	}
	return func(ctx context.Context) (T, error) {
		value, err := action(ctx)
		if err == nil {
			return value, nil
		}
		fallback := recovery(err)
		if fallback == nil {
			panic(panicNilAction)
		}
		return fallback(ctx)
	}
}
