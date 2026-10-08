package svm

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

type NodeChecker struct {
	name   string
	client Client
}

func NewNodeChecker(networkId string, client Client) *NodeChecker {
	return &NodeChecker{
		name:   fmt.Sprintf("svm-node-%s", networkId),
		client: client,
	}
}

func (s *NodeChecker) Name() string {
	return s.name
}

// Check reports the node as up when the call is rate limited: either the local limiter has no capacity left
// or the node answered with a 429, and in both cases the node is reachable.
func (s *NodeChecker) Check(ctx context.Context) health.CheckResult {
	start := time.Now()
	if err := s.client.GetHealth(ctx); err != nil && !ratelimit.IsExceededError(err) {
		return health.NewDownCheckResult(fmt.Errorf("node is not healthy: %w", err), start)
	}
	return health.NewUpCheckResult(start)
}

var _ health.Checker = (*NodeChecker)(nil)
