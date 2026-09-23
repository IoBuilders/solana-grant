package store

import (
	"github.com/google/uuid"
)

// Configuration describes how one Node's ingested data is persisted. Callers
// branch on State: only an ActiveConfiguration names a backend and features.
type Configuration interface {
	NodeID() uuid.UUID
	State() State
	Validate() error
}
