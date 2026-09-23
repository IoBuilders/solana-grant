package node

import (
	"context"
	"sync/atomic"
)

type Bootstrapper interface {
	Start(ctx context.Context) error
	IsStarted(ctx context.Context) bool
}

type DefaultBootstrapper struct {
	started         atomic.Bool
	nodeInitializer NodeInitializer
	nodeLifecycle   NodeLifecycle
}

func NewDefaultBootstrapper(nodeInitializer NodeInitializer, nodeLifecycle NodeLifecycle) *DefaultBootstrapper {
	return &DefaultBootstrapper{nodeLifecycle: nodeLifecycle, nodeInitializer: nodeInitializer}
}

func (b *DefaultBootstrapper) Start(ctx context.Context) error {
	if b.started.Load() {
		return nil
	}
	nodeRunners, err := b.nodeInitializer.Initialize(ctx)
	if err != nil {
		return err
	}
	if err := b.nodeLifecycle.Launch(ctx, nodeRunners); err != nil {
		return err
	}
	b.started.Store(true)
	return nil
}

func (b *DefaultBootstrapper) IsStarted(ctx context.Context) bool {
	return b.started.Load()
}
