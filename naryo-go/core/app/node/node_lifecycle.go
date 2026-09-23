package node

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type NodeLifecycle interface {
	Launch(ctx context.Context, nodeRunners []NodeRunner) error
	LaunchOne(ctx context.Context, nodeRunner NodeRunner) error
	IsRunning(nodeID uuid.UUID) bool
	Stop(ctx context.Context, nodeID uuid.UUID) error
	Restart(ctx context.Context, nodeID uuid.UUID, nodeRunner NodeRunner) error
}

type DefaultNodeLifecycle struct {
	nodeRunners map[uuid.UUID]NodeRunner
}

func NewDefaultNodeLifecycle() *DefaultNodeLifecycle {
	return &DefaultNodeLifecycle{nodeRunners: make(map[uuid.UUID]NodeRunner)}
}

func (l *DefaultNodeLifecycle) Launch(ctx context.Context, nodeRunners []NodeRunner) error {
	for _, nodeRunner := range nodeRunners {
		if err := l.LaunchOne(ctx, nodeRunner); err != nil {
			return err
		}
	}
	return nil
}

func (l *DefaultNodeLifecycle) LaunchOne(ctx context.Context, nodeRunner NodeRunner) error {
	if l.IsRunning(nodeRunner.GetNode().ID) {
		return fmt.Errorf("node %s is already running", nodeRunner.GetNode().ID)
	}
	if err := nodeRunner.Run(ctx); err != nil {
		return err
	}
	l.nodeRunners[nodeRunner.GetNode().ID] = nodeRunner
	return nil
}

func (l *DefaultNodeLifecycle) IsRunning(nodeID uuid.UUID) bool {
	nodeRunner, ok := l.nodeRunners[nodeID]
	return ok && nodeRunner.IsRunning()
}

func (l *DefaultNodeLifecycle) Stop(ctx context.Context, nodeID uuid.UUID) error {
	if !l.IsRunning(nodeID) {
		return fmt.Errorf("node %s is not running", nodeID)
	}
	return l.nodeRunners[nodeID].Stop(ctx)
}

func (l *DefaultNodeLifecycle) Restart(ctx context.Context, nodeID uuid.UUID, nodeRunner NodeRunner) error {
	if err := l.Stop(ctx, nodeID); err != nil {
		return err
	}
	return l.LaunchOne(ctx, nodeRunner)
}

var _ NodeLifecycle = (*DefaultNodeLifecycle)(nil)
