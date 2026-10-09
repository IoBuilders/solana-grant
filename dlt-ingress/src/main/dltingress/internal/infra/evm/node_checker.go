package evm

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"
)

const methodBlockNumber = "eth_blockNumber"

type NodeChecker struct {
	name   string
	client Client
}

func NewNodeChecker(networkId string, client Client) *NodeChecker {
	return &NodeChecker{
		name:   fmt.Sprintf("evm-node-%s", networkId),
		client: client,
	}
}

func (e *NodeChecker) Name() string {
	return e.name
}

// Check reports the node as up when the call is rate limited: either the local limiter has no capacity left
// or the node answered with a 429, and in both cases the node is reachable.
func (e *NodeChecker) Check(ctx context.Context) health.CheckResult {
	start := time.Now()
	var blockNumber string
	if err := e.client.CallRpcMethod(ctx, &blockNumber, methodBlockNumber); err != nil && !ratelimit.IsExceededError(err) {
		return health.NewDownCheckResult(fmt.Errorf("failed to get latest block: %w", err), start)
	}
	return health.NewUpCheckResult(start)
}

var _ health.Checker = (*NodeChecker)(nil)
