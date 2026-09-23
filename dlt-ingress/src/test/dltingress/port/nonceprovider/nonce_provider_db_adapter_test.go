package nonceprovider

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/domain/nonce"
	"dlt-ingress/src/main/dltingress/port/nonceprovider"
	testnonce "dlt-ingress/src/test/dltingress/domain/nonce"
	ethereummocks "dlt-ingress/src/test/dltingress/port/evm"
	repomocks "dlt-ingress/src/test/dltingress/port/repository"
	"fmt"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var networkId = "dummyNetworkId"

func TestDltIngressDbNonceProviderGetNonce_RepositoryError(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	ctx := context.Background()
	expectedError := fmt.Errorf("dummy")

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, expectedError)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_NoEthClient(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	expectedNonce := testnonce.NewNonceTestFactory().CreateEntity()
	ctx := context.Background()
	errorToThrow := fmt.Errorf("dummy")

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(expectedNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(nil, errorToThrow)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, errorToThrow, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_ErrorGettingRemoteNonce(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	ethClient := new(ethereummocks.EvmClientMock)

	request := &nonceprovider.GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	expectedNonce := testnonce.NewNonceTestFactory().CreateEntity()
	ctx := context.Background()
	errorToThrow := fmt.Errorf("dummy")
	expectedError := fmt.Errorf("error getting nonce for address %s under network %s: %w", request.DltAccountId, networkId, errorToThrow)

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(expectedNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(ethClient, nil)
	ethClient.On("PendingNonceAt", ctx, request.DltAccountId).Return(uint64(0), errorToThrow)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_RepositoryValueReturned(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	ethClient := new(ethereummocks.EvmClientMock)
	request := &nonceprovider.GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	dbNonce := testnonce.NewNonceTestFactory().CreateEntity(func(nonce *nonce.Nonce) {
		_ = nonce.UpdateValue(amount.One())
	})
	remoteNonce := dbNonce.Value.RawValue().Uint64() - 1
	expectedNonce := new(dbNonce.Value.Add(*amount.One()))
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(dbNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(ethClient, nil)
	ethClient.On("PendingNonceAt", ctx, request.DltAccountId).Return(remoteNonce, nil)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, resp, expectedNonce)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_RemoteValueReturned(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	ethClient := new(ethereummocks.EvmClientMock)
	request := &nonceprovider.GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	dbNonce := testnonce.NewNonceTestFactory().CreateEntity(func(nonce *nonce.Nonce) {
		_ = nonce.UpdateValue(amount.One())
	})
	remoteNonce := dbNonce.Value.RawValue().Uint64() + 1
	expectedNonce, _ := amount.New(new(big.Int).SetUint64(remoteNonce), 0)
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(dbNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(ethClient, nil)
	ethClient.On("PendingNonceAt", ctx, request.DltAccountId).Return(remoteNonce, nil)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, resp, expectedNonce)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonceUpdating_Ok(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.One(),
	}
	existingNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.Zero())
	})
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(existingNonce, nil)
	repo.On("Save", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(nil)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.Nil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonceCreating_Ok(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.One(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, coredomainerrors.NewEntityNotFoundDomainError("Account", uuid.Nil))
	repo.On("Create", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(nil)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.Nil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoGetError(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.One(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, assert.AnError)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_InvalidNonceValue(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	invalidValue, _ := amount.NewFromString("-1")
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        invalidValue,
	}
	existingNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.Zero())
	})
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(existingNonce, nil)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoSaveError(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.Zero(),
	}
	existingNonce := testnonce.NewNonceTestFactory().CreateEntity()
	ctx := context.Background()
	expectedError := fmt.Errorf("dummy")

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(existingNonce, nil)
	repo.On("Save", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(expectedError)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoCreateError(t *testing.T) {
	repo := new(repomocks.NonceRepositoryMock)
	registry := new(ethereummocks.EvmClientRegistryMock)
	request := &nonceprovider.SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.Zero(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, coredomainerrors.NewEntityNotFoundDomainError("Transaction", uuid.Nil))
	repo.On("Create", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(assert.AnError)

	provider := nonceprovider.NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}
