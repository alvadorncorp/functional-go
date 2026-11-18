# Pattern: value objects

What a Go value object needs from `functional`, and — just as important —
what it does not. Most of a value object is idiomatic Go with no library
involved at all; `functional` only earns its place at two specific points.

---

## What is NOT library code

A value object in Go is, at its core:

- unexported fields, so nothing outside the package can assign into it
  directly;
- a validating constructor that is the only way to produce a non-zero
  instance;
- `With*` methods that return a modified copy, never mutate the receiver.

None of that needs a helper type from this module, or any module. Go's value
receivers already give you copy-on-write for free: a method with a value
receiver `(m Money) WithAmount(cents int64) Money` operates on a copy of `m`
the moment it is called — there is nothing left for a library to wrap. Do
not reach for a "value object base type," a builder library, or a code
generator to get this shape; write the four or five lines by hand. It is
one of the cheapest idioms in the language.

## What IS library code

Two narrow gaps remain where plain Go does not already give you a batteries-
included answer, and this module fills exactly those two:

1. **A collection field.** A plain `[]string` or `map[string]int` field
   defeats the whole point of a value object: even with a validating
   constructor and unexported fields, `Tags()` returning the raw slice hands
   the backing array to every caller, who can then mutate it and silently
   corrupt every other holder of the "same" value object. Use
   [`immutable.List[T]`](../packages/immutable.md),
   [`immutable.Map[K,V]`](../packages/immutable.md), or
   [`immutable.Set[T]`](../packages/immutable.md) for a collection field
   instead — construction copies into an engine-backed structure, and
   `Items()` is the only egress that materializes a conventional mutable
   collection, so the field is safe to expose for iteration via `Values()`.
2. **Your own tests for the structural half.** Prefer unexported fields and
   constructors that copy into immutable types. There is no runtime helper
   in this module that walks a type reflectively; keep that check in the
   consuming package's tests if you need one.

---

## A canonical value object

The shape below uses every piece this pattern calls for: unexported fields,
a validating constructor, `With*`, `Equal`, and a `immutable.List[string]`
field for the one thing plain Go cannot make safe on its own.

```go
package catalog

import (
	"errors"

	"go.alvadorn.org/functional/immutable"
)

// Money is a value object: an amount in cents plus a set of tags, both
// unreachable for mutation once constructed.
type Money struct {
	amountCents int64
	currency    string
	tags        immutable.List[string]
}

// NewMoney validates and constructs. It is the only way to obtain a
// non-zero Money.
func NewMoney(amountCents int64, currency string, tags ...string) (Money, error) {
	if amountCents < 0 {
		return Money{}, errors.New("catalog: amountCents must be non-negative")
	}
	if currency == "" {
		return Money{}, errors.New("catalog: currency is required")
	}
	return Money{
		amountCents: amountCents,
		currency:    currency,
		tags:        immutable.ListOf(tags...),
	}, nil
}

// AmountCents, Currency, and Tags are read-only accessors. Tags returns the
// immutable.List value itself, not a copied []string — the type is already
// safe to hand out for iteration.
func (m Money) AmountCents() int64                 { return m.amountCents }
func (m Money) Currency() string                   { return m.currency }
func (m Money) Tags() immutable.List[string]      { return m.tags }

// WithAmountCents returns a copy of m with a different amount. Value
// receivers already copy m on entry; there is nothing further to do.
func (m Money) WithAmountCents(amountCents int64) (Money, error) {
	if amountCents < 0 {
		return Money{}, errors.New("catalog: amountCents must be non-negative")
	}
	m.amountCents = amountCents
	return m, nil
}

// Equal compares two Money values field by field. immutable.List is not
// ==-comparable, so its own Equal is used for the tags field.
func (m Money) Equal(other Money) bool {
	return m.amountCents == other.amountCents &&
		m.currency == other.currency &&
		m.tags.Equal(other.tags)
}
```

Notes on this example:

- `amountCents`, `currency`, and `tags` are unexported: `Money{amountCents:
  -1}` from outside the package does not compile.
- `NewMoney` is the only constructor and the only place validation runs.
- `WithAmountCents` re-validates rather than trusting the new value blindly
  — a `With*` method is still a construction path.
- `Equal` is hand-written here, field by field, which is the normal shape
  for a value object with a small, fixed field set. `functional`'s own
  `immutable` types resolve their `Equal` through the shared internal
  package `functional/internal/eq`, but that package is `internal` and
  cannot be imported from outside `functional` — a value object outside
  this module writes its own `Equal` by hand, as `Money` does above.
- `immutable.List[string]` is exposed directly by `Tags()`, not copied into
  a `[]string` first — that is the entire benefit of using `immutable`
  here instead of a plain slice field.

---

## Related documents

- [Package API: immutable](../packages/immutable.md)
- [Examples](../examples.md)
