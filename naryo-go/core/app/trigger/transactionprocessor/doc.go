// Package transactionprocessor defines TransactionProcessorPermanentTrigger
// and its Solana implementation. It lives in its own subpackage, rather than
// directly in trigger, for the same reason as slotprocessor and
// blockprocessor: SolanaTransactionProcessorPermanentTrigger depends on
// dispatch.Dispatcher to publish a matched TransactionEvent — and dispatch
// itself depends on trigger, so this dependency cannot live inside trigger
// without an import cycle.
package transactionprocessor
