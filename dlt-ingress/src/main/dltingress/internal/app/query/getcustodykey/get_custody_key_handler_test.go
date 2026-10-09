package getcustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/repository/custodykey/mock"
	"fmt"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	testcustodykey "dlt-ingress/src/main/dltingress/internal/domain/custodykey"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDltIngressGetCustodyKeyQueryHandler_Execute_Success(t *testing.T) {
	repo := new(mockcustodykeyrepo.Postgres)
	factory := testcustodykey.NewCustodyKeyTestFactory()
	key := factory.CreateEntity()

	repo.On("FindByDltAccountId", mock.Anything, key.DltAccountId).Return(key, nil)

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{DltAccountId: key.DltAccountId})

	assert.Nil(t, err)
	resp := rawResp.(Response)
	assert.Equal(t, string(key.KeyType), resp.KeyType)
	assert.Equal(t, string(key.Status), resp.Status)
	assert.Equal(t, key.DltAccountId, resp.DltAccountId)
	assert.Equal(t, string(key.Dlt), resp.Dlt)
	assert.Equal(t, key.ExternalId, resp.ExternalId)
	assert.Equal(t, string(key.CustodyProvider), resp.CustodyProvider)
	repo.AssertExpectations(t)
}

func TestDltIngressGetCustodyKeyQueryHandler_Execute_NotFound(t *testing.T) {
	repo := new(mockcustodykeyrepo.Postgres)
	dltAccountId := "0xUnknown"
	expectedError := domainerrors.NewCustodyKeyNotFoundByDltAccountIdDomainError(dltAccountId)

	repo.On("FindByDltAccountId", mock.Anything, dltAccountId).
		Return((*custodykey.CustodyKey)(nil), expectedError)

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{DltAccountId: dltAccountId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
	repo.AssertExpectations(t)
}

func TestDltIngressGetCustodyKeyQueryHandler_Execute_RepositoryInternalError(t *testing.T) {
	repo := new(mockcustodykeyrepo.Postgres)
	dltAccountId := "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"

	repo.On("FindByDltAccountId", mock.Anything, dltAccountId).
		Return((*custodykey.CustodyKey)(nil), fmt.Errorf("database connection lost"))

	handler := NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), Query{DltAccountId: dltAccountId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	repo.AssertExpectations(t)
}
