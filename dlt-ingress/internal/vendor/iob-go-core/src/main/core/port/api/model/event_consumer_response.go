package model

import (
	"time"

	"github.com/google/uuid"
)

// EventConsumerResponse represents an event that has failed being processed in a consumer.
type EventConsumerResponse struct {
	Id           uuid.UUID `json:"id" format:"uuid" example:"f4585827-0aa5-444c-a995-de62b5f5ac34"`
	Type         string    `json:"type" example:"account.CreatedEvent"`
	Consumer     string    `json:"consumer" example:"usercreated.Listener"`
	Payload      string    `json:"payload" example:"{\"id\":\"af764b3b-1a72-4d4a-a538-f29bf13c4b4f\"}"`
	CreatedAt    time.Time `json:"createdAt" format:"date-time" example:"2026-01-01T10:00:00Z"`
	ErrorDetails *string   `json:"ErrorDetails,omitempty" example:"Account already exists"`
} // @name EventConsumerResponse
