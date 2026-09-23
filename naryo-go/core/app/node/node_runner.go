package node

import (
	"context"
	"sync/atomic"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/dispatch"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/subscription"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/node"
)

type NodeRunner interface {
	Run(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool
	GetNode() *node.Node
	GetSusbcriber() subscription.Subscriber
	GetDispatcher() dispatch.Dispatcher
}

type DefaultNodeRunner struct {
	running    atomic.Bool
	cancelFunc context.CancelFunc
	node       *node.Node
	subscriber subscription.Subscriber
	dispatcher dispatch.Dispatcher
}

func NewDefaultNodeRunner(
	node *node.Node,
	subscriber subscription.Subscriber,
	dispatcher dispatch.Dispatcher,
) *DefaultNodeRunner {
	return &DefaultNodeRunner{node: node, subscriber: subscriber, dispatcher: dispatcher}
}

func (r *DefaultNodeRunner) Run(ctx context.Context) error {
	if r.running.Load() {
		return nil
	}
	nodeCtx, cancelFunc := context.WithCancel(ctx)
	r.cancelFunc = cancelFunc
	r.subscriber.Subscribe(nodeCtx)
	r.running.Store(true)
	return nil
}

func (r *DefaultNodeRunner) Stop(ctx context.Context) error {
	if r.running.Load() {
		r.cancelFunc()
		r.running.Store(false)
	}
	return nil
}

func (r *DefaultNodeRunner) IsRunning() bool {
	return r.running.Load()
}

func (r *DefaultNodeRunner) GetNode() *node.Node {
	return r.node
}

func (r *DefaultNodeRunner) GetSusbcriber() subscription.Subscriber {
	return r.subscriber
}

func (r *DefaultNodeRunner) GetDispatcher() dispatch.Dispatcher {
	return r.dispatcher
}

var _ NodeRunner = (*DefaultNodeRunner)(nil)
