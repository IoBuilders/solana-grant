// Package slotprocessor defines SlotProcessorPermanentTrigger. It lives in
// its own subpackage, rather than directly in trigger, because it depends on
// dispatch.Dispatcher to publish the SolanaBlockEvent it derives from a
// SlotEvent — and dispatch itself depends on trigger, so this dependency
// cannot live inside trigger without an import cycle.
package slotprocessor
