# DAG de domínio e archtest da app

Direção sempre para dentro: adapters/composition → application → domain.
Domínio: sem `context`, `workflow`, `postgres`, `net/http`, `database/sql`.

Checkout: `value ← support ← order ← checkout`. A **app** copia um
`internal/archtest` (replace set, purity, DAG). Um módulo kit só na Onda 1,
com gatilho. Não há layout hexagonal oficial na biblioteca.
