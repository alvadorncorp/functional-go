# ACL between bounded contexts

A pure function translates one context's data into the other's vocabulary.
Neither context needs to know the other's business rules. A bounded context
does not imply a separate transaction boundary
(`adr:workflow:no-distributed-transaction-abstraction`).
Don't extract an `ACL` type into the library.
