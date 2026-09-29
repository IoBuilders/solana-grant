package event

import (
	"context"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/retry"
)

type RetryableListener struct {
	wrappedListener Listener
	retryer         retry.Retryer
	retryOptions    retry.Options
}

type RetryableListenerOption func(*RetryableListener)

func WithRetryOptions(opts retry.Options) RetryableListenerOption {
	return func(rl *RetryableListener) { rl.retryOptions = opts }
}

func NewRetryableListener(listener Listener, retryer retry.Retryer, o ...RetryableListenerOption) *RetryableListener {
	rl := &RetryableListener{
		wrappedListener: listener,
		retryer:         retryer,
		retryOptions:    retry.NewOptions(),
	}
	for _, opt := range o {
		opt(rl)
	}
	return rl
}

func (rl *RetryableListener) Listen(ctx context.Context, event Event) error {
	_, err := rl.retryer.Execute(ctx, rl.retryOptions, func(ctx context.Context) (any, error) {
		return nil, rl.wrappedListener.Listen(ctx, event)
	})
	return err
}
