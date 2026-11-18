# Domain DAG and app archtest

Direction always inward: adapters/composition → application → domain.
Domain: no `context`, `workflow`, `postgres`, `net/http`, `database/sql`.

Checkout: `value ← support ← order ← checkout`. The **app** copies an
`internal/archtest` (replace set, purity, DAG). A kit module only in Wave 1,
with a trigger. The library has no official hexagonal layout.
