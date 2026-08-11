// Package workflow provides small, type-safe primitives for building,
// composing, and executing reusable effectful workflows in Go.
//
// A Workflow[I,O] is a reusable operation that, given an input I, produces a
// deferred computation capable of yielding O. Dependencies are captured when
// the workflow is constructed, the input arrives when the workflow is applied,
// and a context.Context arrives only when the computation is executed. Applying
// a workflow to an input performs no work: it returns a value describing what
// would happen.
//
// Execution is explicit and lives in
// github.com/alvadorncorp/gopackages/workflow/action, whose Action[T] is a
// deferred, context-bound computation and whose Run is the single passage from
// description to execution. An Action is re-executable, never memoized, and
// never started in the background.
//
// Transactional composition lives in
// github.com/alvadorncorp/gopackages/workflow/transaction, which models a
// boundary supplied by an adapter and the scoped capabilities it provides. The
// module defines no transaction implementation and no concrete scope type.
//
// Business rules stay in the application domain, and infrastructure stays
// behind application-defined capabilities. The module has no engine, runner,
// scheduler, registry, container, service locator, dispatcher, reflection,
// persistence, transport, or global runtime, and it makes no durability,
// retry, timeout, parallelism, or distributed-atomicity promise. Exported
// signatures remain ordinary Go and standard-library types.
package workflow
