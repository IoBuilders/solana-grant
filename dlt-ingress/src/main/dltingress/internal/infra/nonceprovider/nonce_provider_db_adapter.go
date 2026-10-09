package nonceprovider

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/nonce"
	"dlt-ingress/src/main/dltingress/internal/infra/evm"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

type DbNonceProvider struct {
	repository        nonce.Repository
	evmClientRegistry evm.ClientRegistry
}

func NewDbNonceProvider(repository nonce.Repository, evmClientRegistry evm.ClientRegistry) *DbNonceProvider {
	return &DbNonceProvider{
		repository,
		evmClientRegistry,
	}
}

func (dnp *DbNonceProvider) GetNonce(ctx context.Context, request *GetNonceRequest) (*amount.Amount, error) {
	currentNonce, err := dnp.repository.FindAndLockByDltAccountIdAndNetworkId(ctx, request.DltAccountId, request.NetworkId)
	if err != nil && !errors.Is(err, coreerror.ErrNotFound) {
		return nil, err
	}

	if currentNonce != nil {
		if err = dnp.waitForBlockTime(ctx, request.NetworkId, currentNonce.UpdatedAt); err != nil {
			return nil, err
		}
	}

	remoteNonceValue, err := dnp.getRemoteNonce(ctx, request.NetworkId, request.DltAccountId)
	if err != nil {
		return nil, err
	}

	var nonce amount.Amount
	if currentNonce == nil || currentNonce.Value.LessThan(*remoteNonceValue) {
		nonce = *remoteNonceValue
	} else {
		nonce = currentNonce.Value.Add(*amount.One())
	}
	return &nonce, nil
}

func (dnp *DbNonceProvider) SetNonce(ctx context.Context, request *SetNonceRequest) error {
	accountNonce, err := dnp.repository.FindAndLockByDltAccountIdAndNetworkId(ctx, request.DltAccountId, request.NetworkId)
	if err != nil && !errors.Is(err, coreerror.ErrNotFound) {
		return err
	}

	if accountNonce == nil {
		if accountNonce, err = nonce.NewNonce(request.DltAccountId, request.NetworkId, request.Value); err != nil {
			return err
		}

		if err = dnp.repository.Create(ctx, accountNonce); err != nil {
			return err
		}
	} else {
		if err = accountNonce.UpdateValue(request.Value); err != nil {
			return err
		}

		if err = dnp.repository.Save(ctx, accountNonce); err != nil {
			return err
		}
	}

	return nil
}

// waitForBlockTime blocks until the network's configured block time has elapsed since lastUsedAt, so that
// the next nonce for this wallet is not handed out before the previous one had time to be mined.
//
// This is a temporary, heuristic stopgap: it assumes the previous transaction gets mined within one
// average block time, but it doesn't verify that it actually did. A slow block still lets the next nonce
// through too early, a fast one wastes time waiting, and a transaction that never gets mined isn't detected
// at all.
func (dnp *DbNonceProvider) waitForBlockTime(ctx context.Context, networkId string, lastUsedAt time.Time) error {
	networkConfig, err := config.DltIngressConfig.DltIngress.GetDltIngressNetwork(networkId)
	if err != nil || networkConfig.BlockTime <= 0 {
		return nil
	}

	remaining := networkConfig.BlockTime - time.Since(lastUsedAt)
	if remaining <= 0 {
		return nil
	}

	timer := time.NewTimer(remaining)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (dnp *DbNonceProvider) getRemoteNonce(ctx context.Context, networkId string, dltAccountId string) (*amount.Amount, error) {
	evmClient, err := dnp.evmClientRegistry.GetClientForNetworkId(ctx, networkId)
	if err != nil {
		return nil, err
	}

	rawNonce, err := evmClient.PendingNonceAt(ctx, dltAccountId)
	if err != nil {
		return nil, fmt.Errorf("error getting nonce for address %s under network %s: %w", dltAccountId, networkId, err)
	}

	return amount.New(new(big.Int).SetUint64(rawNonce), 0)
}
