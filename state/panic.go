package state

// panicTransactionUsedAfterCommit is what a transaction method panics with
// once Commit has already run on that transaction. Commit freezes the
// transaction so no method can honor a call afterward.
const panicTransactionUsedAfterCommit = "state: transaction used after Commit"

// panicUnrelatedLineages is what comparison or join operations panic with
// when two states do not share a common lineage root.
const panicUnrelatedLineages = "state: unrelated lineages"
