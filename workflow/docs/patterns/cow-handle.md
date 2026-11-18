# Handle COW for large aggregates

`hugeParam` (gocritic, 80 bytes) is a guardrail over values the **application**
copies through signatures, not a reason to pointerize `Decision`, `Reducer` or
`Workflow` (`adr:modeling:streaming-replay-project-seq`).

Pattern (checkout's `Order`, withdraw's `Account`): a struct
`Order { p *orderPayload }` with a by-value API; the payload is immutable
once published.

Getters on the zero value (`p == nil`) return zeros; edits are rejected.
Mutable collections live in `functional/immutable` **inside** the payload.

Don't extract `Handle[T]`. See `functional/docs/patterns/cow-handles.md`.
