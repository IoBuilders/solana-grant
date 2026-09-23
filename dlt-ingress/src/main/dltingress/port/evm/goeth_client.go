package evm

import (
	"context"
	"fmt"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type GoEthClient struct {
	ethClient *ethclient.Client
}

func NewGoEthClient(ctx context.Context, url string) (*GoEthClient, error) {
	ethClient, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("error creating ethereum client for url %s: %w", url, err)
	}

	return &GoEthClient{
		ethClient,
	}, nil
}

func (c *GoEthClient) PendingNonceAt(ctx context.Context, account string) (uint64, error) {
	return c.ethClient.PendingNonceAt(ctx, common.HexToAddress(account))
}

func (c *GoEthClient) GetBaseFeePerGas(ctx context.Context) (*amount.Amount, error) {
	latestBlock, err := c.ethClient.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving latest block %w", err)
	}

	baseFeePerGas, err := amount.New(latestBlock.Header().BaseFee, 0)
	if err != nil {
		return nil, fmt.Errorf("error converting to amount %w", err)
	}
	return baseFeePerGas, nil

}

func (c *GoEthClient) CallRpcMethod(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	return c.ethClient.Client().CallContext(ctx, result, method, args...)
}

func (c *GoEthClient) GetGasPrice(ctx context.Context) (*amount.Amount, error) {
	price, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching gas price: %w", err)
	}
	result, err := amount.New(price, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid gas price amount: %w", err)
	}
	return result, nil
}

func (c *GoEthClient) Close() {
	c.ethClient.Close()
}
