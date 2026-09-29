package healthchecks

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
)

type EthereumNodeChecker struct {
	name   string
	client *ethclient.Client
}

func NewEthereumNodeChecker(name string, client *ethclient.Client) *EthereumNodeChecker {
	return &EthereumNodeChecker{
		name:   name,
		client: client,
	}
}

func (e *EthereumNodeChecker) Name() string {
	return e.name
}

func (e *EthereumNodeChecker) Check(ctx context.Context) health.CheckResult {
	start := time.Now()
	if _, err := e.client.BlockNumber(ctx); err != nil {
		return health.NewDownCheckResult(fmt.Errorf("failed to get latest block: %w", err), start)
	}
	return health.NewUpCheckResult(start)
}
