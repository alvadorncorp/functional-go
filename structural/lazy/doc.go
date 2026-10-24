// Package lazy provides concurrent-safe deferred, memoized computations.
// A Lazy evaluates at most once and memoizes either its value or panic.
// Construct values with New or Of; zero values and nil *Lazy pointers are invalid.
// An initializer calling runtime.Goexit causes Get and later calls to panic;
// under Go's default panic(nil) conversion (main module go >= 1.21, GODEBUG
// panicnil unset) the payload is *runtime.PanicNilError.
package lazy
