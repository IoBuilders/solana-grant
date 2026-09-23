package interactor

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
)

// Type discriminates the kind of interaction performed against a Node.
//
// Reserved for future support: non-block interaction kinds.
type Type string

const (
	TypeBlock Type = "BLOCK"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeBlock:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}

// Interactor is the base contract for on-demand interactions with a Node.
type Interactor interface {
	Type() Type
}

// BlockInteractor fetches chain data from a Node on demand. Implementations
// return TypeBlock from Type().
type BlockInteractor interface {
	Interactor

	// GetBlock returns the block at slot. If the leader skipped slot, the
	// returned error wraps block.ErrSlotSkipped. If the block exists but
	// is not yet available on the node, it wraps block.ErrBlockNotAvailable.
	// If the block was pruned from the node's storage, it wraps
	// block.ErrBlockCleanedUp.
	GetBlock(ctx context.Context, slot uint64) (*block.SolanaBlock, error)

	// GetSlot returns the node's current slot, at the same implicit
	// commitment level GetBlock resolves against, so callers can compare
	// a requested slot with it directly.
	GetSlot(ctx context.Context) (uint64, error)
}
