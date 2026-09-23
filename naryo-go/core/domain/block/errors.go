package block

import "errors"

// Sentinels a BlockInteractor wraps when a slot cannot be resolved to a
// block right now, so callers can distinguish the cause with errors.Is.
var (
	// ErrSlotSkipped means the leader never produced a block for the
	// requested slot: no block will ever exist for it.
	ErrSlotSkipped = errors.New("slot skipped")

	// ErrBlockNotAvailable means the block may exist but is not yet
	// available from the node (e.g. not yet rooted, or still being
	// written). Retrying after a delay may succeed.
	ErrBlockNotAvailable = errors.New("block not available")

	// ErrBlockCleanedUp means the block existed but has since been pruned
	// from the node's storage and can no longer be retrieved.
	ErrBlockCleanedUp = errors.New("block cleaned up")
)
