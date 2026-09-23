package subscription

import (
	"context"
)

// Subscriber is the base contract of every Node subscription.
type Subscriber interface {
	Subscribe(ctx context.Context)
}
