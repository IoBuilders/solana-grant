package filter

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SyncState is the synchronization progress of a single Filter: the last slot
// processed for it and where it stands in catching up with the chain.
type SyncState struct {
	FilterID          uuid.UUID
	LastProcessedSlot uint64
	Status            SyncStatus
}

func NewSyncState(filterID uuid.UUID, lastProcessedSlot uint64, status SyncStatus) (*SyncState, error) {
	if filterID == uuid.Nil {
		return nil, domainerrors.NewEmptyFieldError("FilterID", "SyncState")
	}
	if !status.IsValid() {
		return nil, domainerrors.NewInvalidFieldError("Status", "SyncState", "unsupported status "+status.String())
	}
	return &SyncState{
		FilterID:          filterID,
		LastProcessedSlot: lastProcessedSlot,
		Status:            status,
	}, nil
}
