package getfailedevents

import (
	"time"

	"github.com/google/uuid"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
)

type Query struct {
	PaginationParams pagination.PaginationParams
}

type Response = pagination.PaginatedQueryResponse[*EventConsumerQueryResponse]

type EventConsumerQueryResponse struct {
	Id           uuid.UUID
	Type         string
	Consumer     string
	Payload      string
	CreatedAt    time.Time
	ErrorDetails *string
}
