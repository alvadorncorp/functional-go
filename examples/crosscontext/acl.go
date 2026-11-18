// The anti-corruption layer is the only code that speaks both vocabularies,
// and it holds exactly one pure function. Banking never learns what a book
// is, and Ledger never learns what a balance is; each context can evolve on
// its own because the translation lives here, at the seam.
package main

// ToLedgerCommand translates a Banking integration event into a Ledger
// command. It is pure and total: no effects, no failure channel, no context.
func ToLedgerCommand(event TransferOccurred) RecordTransferCommand {
	return RecordTransferCommand{
		Reference: event.Reference,
		DebitOf:   string(event.From),
		CreditOf:  string(event.To),
		Amount:    Amount(event.Amount),
	}
}
