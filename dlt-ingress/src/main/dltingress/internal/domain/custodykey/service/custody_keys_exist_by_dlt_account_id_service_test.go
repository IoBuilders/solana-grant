//go:build test

package servicecustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey/mock"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	custodykeytest "dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func newExistMultipleService() (*ExistMultiple, *mockcustodykeyrepo.Postgres) {
	repo := new(mockcustodykeyrepo.Postgres)
	return NewExistMultiple(repo), repo
}

func TestDltIngressExistMultiple_Execute_Success(t *testing.T) {
	svc, repo := newExistMultipleService()
	factory := custodykeytest.NewCustodyKeyTestFactory()
	ctx := context.Background()

	keyA := factory.CreateEntity(func(k *custodykey.CustodyKey) { k.DltAccountId = "0xAAA" })
	keyB := factory.CreateEntity(func(k *custodykey.CustodyKey) { k.DltAccountId = "0xBBB" })
	ids := []string{"0xAAA", "0xBBB"}

	repo.On("FindByDltAccountIds", ctx, ids).Return([]*custodykey.CustodyKey{keyA, keyB}, nil)

	result, err := svc.Execute(ctx, ids)

	require.NoError(t, err)
	assert.Equal(t, []*custodykey.CustodyKey{keyA, keyB}, result)
	repo.AssertExpectations(t)
}

func TestDltIngressExistMultiple_Execute_RepositoryError_ReturnsError(t *testing.T) {
	svc, repo := newExistMultipleService()
	ctx := context.Background()
	ids := []string{"0xAAA"}

	repo.On("FindByDltAccountIds", ctx, ids).Return(([]*custodykey.CustodyKey)(nil), assert.AnError)

	result, err := svc.Execute(ctx, ids)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertExpectations(t)
}

func TestDltIngressExistMultiple_Execute_OneIdNotFound_ReturnsNotFoundError(t *testing.T) {
	svc, repo := newExistMultipleService()
	factory := custodykeytest.NewCustodyKeyTestFactory()
	ctx := context.Background()

	keyA := factory.CreateEntity(func(k *custodykey.CustodyKey) { k.DltAccountId = "0xAAA" })
	ids := []string{keyA.DltAccountId, "0xMISSING"}

	repo.On("FindByDltAccountIds", ctx, ids).Return([]*custodykey.CustodyKey{keyA}, nil)

	result, err := svc.Execute(ctx, ids)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Equal(t, domainerrors.ErrorCodeCustodyKeyNotFoundByDltAccountId, err.(coreerror.DomainError).ErrorCode())
	repo.AssertExpectations(t)
}

func TestDltIngressExistMultiple_Execute_AllIdsNotFound_ReturnsNotFoundError(t *testing.T) {
	svc, repo := newExistMultipleService()
	ctx := context.Background()
	ids := []string{"0xMISSING1", "0xMISSING2"}

	repo.On("FindByDltAccountIds", ctx, ids).Return([]*custodykey.CustodyKey{}, nil)

	result, err := svc.Execute(ctx, ids)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Equal(t, domainerrors.ErrorCodeCustodyKeyNotFoundByDltAccountId, err.(coreerror.DomainError).ErrorCode())
	repo.AssertExpectations(t)
}
