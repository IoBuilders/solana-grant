package evm

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type GoEthClient struct {
	ethClient *ethclient.Client
}

func NewGoEthClient(ctx context.Context, url string, httpClient *http.Client) (*GoEthClient, error) {
	rpcClient, err := rpc.DialOptions(ctx, url, rpc.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("error creating ethereum client for url %s: %w", url, err)
	}

	return &GoEthClient{
		ethclient.NewClient(rpcClient),
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

func (c *GoEthClient) GetBalance(ctx context.Context, account string) (*amount.Amount, error) {
	balance, err := c.ethClient.BalanceAt(ctx, common.HexToAddress(account), nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving balance for account %s: %w", account, err)
	}

	result, err := amount.New(balance, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid balance amount: %w", err)
	}
	return result, nil
}

func (c *GoEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return c.ethClient.HeaderByNumber(ctx, number)
}

func (c *GoEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return c.ethClient.BlockByNumber(ctx, number)
}

func (c *GoEthClient) NetworkID(ctx context.Context) (*big.Int, error) {
	return c.ethClient.NetworkID(ctx)
}

func (c *GoEthClient) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return c.ethClient.CallContract(ctx, msg, blockNumber)
}

func (c *GoEthClient) Close() {
	c.ethClient.Close()
}
