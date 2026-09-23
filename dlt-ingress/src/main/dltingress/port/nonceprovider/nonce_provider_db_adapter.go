package nonceprovider

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/nonce"
	"dlt-ingress/src/main/dltingress/port/evm"
	"dlt-ingress/src/main/dltingress/port/repository/nonce"
	"errors"
	"fmt"
	"math/big"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

type DbNonceProvider struct {
	repository        noncerepo.Repository
	evmClientRegistry evm.ClientRegistry
}

func NewDbNonceProvider(repository noncerepo.Repository, evmClientRegistry evm.ClientRegistry) *DbNonceProvider {
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
