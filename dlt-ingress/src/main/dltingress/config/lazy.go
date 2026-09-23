package dltingressconfig

import (
	"context"
	"dlt-ingress/src/main/core/shared"
	"sync"
)

type LazyDltIngress struct {
	once sync.Once
	core *shared.CoreCommon
	opts []DltIngressOption
	*DltIngress
	shutdownFn func(context.Context) error
}

func NewLazyDltIngress(core *shared.CoreCommon, opts ...DltIngressOption) *LazyDltIngress {
	return &LazyDltIngress{core: core, opts: opts}
}

func (l *LazyDltIngress) Init(ctx context.Context) {
	l.once.Do(func() {
		l.DltIngress, l.shutdownFn = Setup(ctx, l.core, l.opts...)
	})
}

// ShutdownFn returns the graceful shutdown function registered during Init.
// Must be called after Init.
func (l *LazyDltIngress) ShutdownFn() func(context.Context) error {
	return l.shutdownFn
}
