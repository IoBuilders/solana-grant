package dltingressconfig

import (
	"context"
	"dlt-ingress/src/main/config"
	"sync"
)

type LazyDltIngress struct {
	once             sync.Once
	dltIngressconfig *config.Config
	core             *CoreDependencies
	opts             []DltIngressOption
	*DltIngress
	shutdownFn func(context.Context) error
}

func NewLazyDltIngress(dltIngressconfig *config.Config, core *CoreDependencies, opts ...DltIngressOption) *LazyDltIngress {
	return &LazyDltIngress{dltIngressconfig: dltIngressconfig, core: core, opts: opts}
}

func (l *LazyDltIngress) Init(ctx context.Context) {
	l.once.Do(func() {
		l.DltIngress, l.shutdownFn = Setup(ctx, l.dltIngressconfig, l.core, l.opts...)
	})
}

// ShutdownFn returns the graceful shutdown function registered during Init.
// Must be called after Init.
func (l *LazyDltIngress) ShutdownFn() func(context.Context) error {
	return l.shutdownFn
}
