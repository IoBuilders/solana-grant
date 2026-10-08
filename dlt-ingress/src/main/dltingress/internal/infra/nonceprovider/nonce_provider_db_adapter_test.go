package nonceprovider

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/nonce"
	testnonce "dlt-ingress/src/main/dltingress/internal/domain/nonce"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/nonce/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var networkId = "dummyNetworkId"

func init() {
	config.DltIngressConfig = &config.Config{}
	config.DltIngressConfig.DltIngress.Networks = []*config.NetworkConfig{
		{
			Id:        networkId,
			BlockTime: 0,
		},
	}
}

func TestDltIngressDbNonceProviderGetNonce_RepositoryError(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	ctx := context.Background()
	expectedError := fmt.Errorf("dummy")

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, expectedError)

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_NoEthClient(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
	}
	expectedNonce := testnonce.NewNonceTestFactory().CreateEntity()
	ctx := context.Background()
	errorToThrow := fmt.Errorf("dummy")

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(expectedNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(nil, errorToThrow)

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, errorToThrow, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_ErrorGettingRemoteNonce(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)

	request := &GetNonceRequest{
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

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_RepositoryValueReturned(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	request := &GetNonceRequest{
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

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, resp, expectedNonce)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_RemoteValueReturned(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	request := &GetNonceRequest{
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

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, err)
	assert.Equal(t, resp, expectedNonce)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonceUpdating_Ok(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &SetNonceRequest{
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

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.Nil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonceCreating_Ok(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.One(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, coredomainerrors.NewEntityNotFoundDomainError("Account", uuid.Nil))
	repo.On("Create", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(nil)

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.Nil(t, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoGetError(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.One(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, assert.AnError)

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_InvalidNonceValue(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	invalidValue, _ := amount.NewFromString("-1")
	request := &SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        invalidValue,
	}
	existingNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.Zero())
	})
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(existingNonce, nil)

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoSaveError(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &SetNonceRequest{
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

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.Equal(t, expectedError, err)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_WaitsForBlockTime(t *testing.T) {
	blockTimeNetworkId := "blockTimeNetworkId"
	blockTime := 100 * time.Millisecond
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{
		Id:        blockTimeNetworkId,
		BlockTime: blockTime,
	})
	defer func() {
		config.DltIngressConfig.DltIngress.Networks = config.DltIngressConfig.DltIngress.Networks[:len(config.DltIngressConfig.DltIngress.Networks)-1]
	}()

	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	request := &GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    blockTimeNetworkId,
	}
	dbNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.One())
		n.UpdatedAt = time.Now()
	})
	remoteNonce := dbNonce.Value.RawValue().Uint64() - 1
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(dbNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(ethClient, nil)
	ethClient.On("PendingNonceAt", ctx, request.DltAccountId).Return(remoteNonce, nil)

	provider := NewDbNonceProvider(repo, registry)
	start := time.Now()
	resp, err := provider.GetNonce(ctx, request)
	elapsed := time.Since(start)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.GreaterOrEqual(t, elapsed, blockTime)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_BlockTimeAlreadyElapsed(t *testing.T) {
	blockTimeNetworkId := "blockTimeElapsedNetworkId"
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{
		Id:        blockTimeNetworkId,
		BlockTime: time.Millisecond,
	})
	defer func() {
		config.DltIngressConfig.DltIngress.Networks = config.DltIngressConfig.DltIngress.Networks[:len(config.DltIngressConfig.DltIngress.Networks)-1]
	}()

	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	ethClient := new(mock2.EvmClientMock)
	request := &GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    blockTimeNetworkId,
	}
	dbNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.One())
		n.UpdatedAt = time.Now().Add(-time.Hour)
	})
	remoteNonce := dbNonce.Value.RawValue().Uint64() - 1
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(dbNonce, nil)
	registry.On("GetClientForNetworkId", ctx, request.NetworkId).Return(ethClient, nil)
	ethClient.On("PendingNonceAt", ctx, request.DltAccountId).Return(remoteNonce, nil)

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	repo.AssertExpectations(t)
	registry.AssertExpectations(t)
	ethClient.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderGetNonce_BlockTimeContextCancelled(t *testing.T) {
	blockTimeNetworkId := "blockTimeCancelledNetworkId"
	config.DltIngressConfig.DltIngress.Networks = append(config.DltIngressConfig.DltIngress.Networks, &config.NetworkConfig{
		Id:        blockTimeNetworkId,
		BlockTime: time.Hour,
	})
	defer func() {
		config.DltIngressConfig.DltIngress.Networks = config.DltIngressConfig.DltIngress.Networks[:len(config.DltIngressConfig.DltIngress.Networks)-1]
	}()

	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &GetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    blockTimeNetworkId,
	}
	dbNonce := testnonce.NewNonceTestFactory().CreateEntity(func(n *nonce.Nonce) {
		_ = n.UpdateValue(amount.One())
		n.UpdatedAt = time.Now()
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(dbNonce, nil)

	provider := NewDbNonceProvider(repo, registry)
	resp, err := provider.GetNonce(ctx, request)

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, context.Canceled)
	repo.AssertExpectations(t)
}

func TestDltIngressDbNonceProviderSetNonce_RepoCreateError(t *testing.T) {
	repo := new(mocknoncerepo.Postgres)
	registry := new(mock2.EvmClientRegistryMock)
	request := &SetNonceRequest{
		DltAccountId: "dltAccountId",
		NetworkId:    networkId,
		Value:        amount.Zero(),
	}
	ctx := context.Background()

	repo.On("FindAndLockByDltAccountIdAndNetworkId", ctx, request.DltAccountId, request.NetworkId).Return(nil, coredomainerrors.NewEntityNotFoundDomainError("Transaction", uuid.Nil))
	repo.On("Create", ctx, mock.MatchedBy(func(n *nonce.Nonce) bool {
		return n.Value == request.Value
	})).Return(assert.AnError)

	provider := NewDbNonceProvider(repo, registry)
	err := provider.SetNonce(ctx, request)

	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}
