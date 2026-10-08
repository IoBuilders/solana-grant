package model

import "encoding/json"

// CreateBatchRequest is the generic envelope every batch endpoint receives.
type CreateBatchRequest struct {
	SignerDltAccountId string            `json:"signerDltAccountId" binding:"required" example:"0x8be5336c0176ae5c68de296b5a3ff2b99739f4f9"`
	NetworkId          string            `json:"networkId" binding:"required" example:"default"`
	ValidationType     string            `json:"validationType" binding:"required" example:"FAIL_SAFE" enums:"FAIL_FAST,FAIL_SAFE"`
	Items              []json.RawMessage `json:"items" binding:"required"`
} // @name CreateBatchRequest
