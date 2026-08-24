# Pattern: copy-on-write value handles

When an aggregate's payload would blow the `hugeParam` budget (~80 bytes) if
copied through `modeling.Decision[S,E]` or `workflow.Workflow`, keep the
**public type a small value** and hide the payload behind an unexported
pointer. Publish-once: never mutate `*payload` in place after construction.

This is application code. This module does not export `Handle[T]`. The
internal `ownership` / `cow` packages stay internal.

Use `immutable.List`/`Map`/`Set` inside the payload for collection fields
(see [value-objects.md](value-objects.md)).

Evidence: `examples/workflow/checkout` `Order`, `examples/workflow/withdraw` `Account`.
