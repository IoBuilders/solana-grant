package boundedblockingqueue

import (
	"context"

	"github.com/google/uuid"
)

type Port interface {
	Put(ctx context.Context, networkId string, value uuid.UUID) error
	Take(ctx context.Context, networkId string) error
}
