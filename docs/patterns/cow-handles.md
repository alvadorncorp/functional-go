# Pattern: copy-on-write value handles

When an aggregate's payload would blow the `hugeParam` budget (~80 bytes) if
copied through `modeling.Decision[S,E]` or `workflow.Workflow`, keep the
**public type a small value** and hide the payload behind an unexported
pointer. Publish-once: never mutate `*payload` in place after construction.

This is application code. This module does not export `Handle[T]`. The
internal `ownership` / `cow` packages stay internal.

Use `immutable.List`/`Map`/`Set` inside the payload for collection fields
(see [value-objects.md](value-objects.md)).

## Lighter step: a private `edit` helper

Every infallible edit method repeats the same steps: guard the zero handle,
copy the payload, change a field, wrap the copy in a new handle. A private
`edit` helper on the handle holds those steps once:

```go
// edit returns a new Order whose payload is a copy of o's with fn applied.
// The zero Order stays zero and fn is not called.
func (o Order) edit(fn func(*orderPayload)) Order {
	if o.p == nil {
		return Order{}
	}
	cp := *o.p
	fn(&cp)
	return Order{p: &cp}
}
```

`MarkCheckedOut` before:

```go
func (o Order) MarkCheckedOut() Order {
	next := o.clonePayload()
	if next == nil {
		return Order{}
	}
	next.status = OrderStatusCheckedOut
	return Order{p: next}
}
```

After:

```go
func (o Order) MarkCheckedOut() Order {
	return o.edit(func(p *orderPayload) { p.status = OrderStatusCheckedOut })
}
```

Rules:

- **The zero handle stays zero.** `edit` on the zero handle returns the zero
  handle without calling `fn`, the same as the hand-written methods.
- **Fallible edits still return their error.** `edit` only covers the
  copy-and-assign step. Methods such as `AddItem` or `ApplyCoupon` keep
  their own guard and still return `ErrOrderNotEditable` for a zero or
  checked-out order. They can call `edit` once the guard passes.
- **`fn` must not retain the `*payload`.** The pointer is the unpublished
  copy. Storing it in a field, a closure that outlives the call or a
  goroutine lets someone mutate the payload after the new handle is
  published.
- **The copy is shallow.** Collection fields stay safe because they are
  `immutable` values. A raw slice or map field would be shared between the
  old and new payloads.
- **The helper is unexported application code.** Write it next to the
  handle in your own package. This module provides no `edit` helper.

Evidence: `examples/workflow/checkout` `Order`, `examples/workflow/withdraw` `Account`.

## Generated option: cowgen

[cowgen](../../../cowgen/docs/overview.md) writes this boilerplate for you.
From the payload struct it generates `Valid`, the zero-safe getters,
`clonePayload`, `edit` and opt-in `With` setters into a file such as
`order_cow.go` for `Order`.
The rules above still hold: constructors and fallible edits stay
hand-written, and a reference-typed field needs `cow:"-"` and a hand-written
getter that clones it.
