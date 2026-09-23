// Package blockprocessor defines BlockProcessorPermanentTrigger and its
// Solana implementation. It lives in its own subpackage, rather than
// directly in trigger, for the same reason as slotprocessor: it depends on
// dispatch.Dispatcher to publish the TransactionEvents it derives from a
// BlockEvent — and dispatch itself depends on trigger, so this dependency
// cannot live inside trigger without an import cycle.
package blockprocessor
