package model

import (
	"encoding/json"

	"github.com/google/uuid"
)

// CreateBatchResponse is the generic envelope every batch endpoint returns.
type CreateBatchResponse struct {
	ProcessId uuid.UUID         `json:"processId" format:"uuid" example:"f4585827-0aa5-444c-a995-de62b5f5ac34"`
	Status    string            `json:"status" example:"STARTED"`
	BatchSize int               `json:"batchSize" example:"50"`
	Items     []json.RawMessage `json:"items"`
	// Items rejected by validation, present only under FAIL_SAFE when some item failed
	InvalidItems []InvalidBatchItem `json:"invalidItems,omitempty"`
} // @name CreateBatchResponse

// InvalidBatchItem reports one item dropped by the per-endpoint invariant validation service.
type InvalidBatchItem struct {
	Index    int        `json:"index" example:"3"`
	EntityId *uuid.UUID `json:"entityId,omitempty" format:"uuid" example:"f4585827-0aa5-444c-a995-de62b5f5ac34"`
	Reason   string     `json:"reason" example:"asset is not ISSUED"`
} // @name InvalidBatchItem
