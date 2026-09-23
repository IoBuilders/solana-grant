package getcustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"fmt"
	"testing"

	"dlt-ingress/src/main/dltingress/app/query/getcustodykey"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	testcustodykey "dlt-ingress/src/test/dltingress/domain/custodykey"
	repomocks "dlt-ingress/src/test/dltingress/port/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDltIngressGetCustodyKeyQueryHandler_Execute_Success(t *testing.T) {
	repo := new(repomocks.CustodyKeyRepositoryMock)
	factory := testcustodykey.NewCustodyKeyTestFactory()
	key := factory.CreateEntity()

	repo.On("FindByDltAccountId", mock.Anything, key.DltAccountId).Return(key, nil)

	handler := getcustodykey.NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), getcustodykey.Query{DltAccountId: key.DltAccountId})

	assert.Nil(t, err)
	resp := rawResp.(getcustodykey.Response)
	assert.Equal(t, string(key.KeyType), resp.KeyType)
	assert.Equal(t, string(key.Status), resp.Status)
	assert.Equal(t, key.DltAccountId, resp.DltAccountId)
	assert.Equal(t, string(key.Dlt), resp.Dlt)
	assert.Equal(t, key.ExternalId, resp.ExternalId)
	assert.Equal(t, string(key.CustodyProvider), resp.CustodyProvider)
	repo.AssertExpectations(t)
}

func TestDltIngressGetCustodyKeyQueryHandler_Execute_NotFound(t *testing.T) {
	repo := new(repomocks.CustodyKeyRepositoryMock)
	dltAccountId := "0xUnknown"
	expectedError := domainerrors.NewCustodyKeyNotFoundByDltAccountIdDomainError(dltAccountId)

	repo.On("FindByDltAccountId", mock.Anything, dltAccountId).
		Return((*custodykey.CustodyKey)(nil), expectedError)

	handler := getcustodykey.NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), getcustodykey.Query{DltAccountId: dltAccountId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), expectedError.Error())
	repo.AssertExpectations(t)
}

func TestDltIngressGetCustodyKeyQueryHandler_Execute_RepositoryInternalError(t *testing.T) {
	repo := new(repomocks.CustodyKeyRepositoryMock)
	dltAccountId := "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"

	repo.On("FindByDltAccountId", mock.Anything, dltAccountId).
		Return((*custodykey.CustodyKey)(nil), fmt.Errorf("database connection lost"))

	handler := getcustodykey.NewHandler(repo)
	rawResp, err := handler.Execute(context.Background(), getcustodykey.Query{DltAccountId: dltAccountId})

	assert.Nil(t, rawResp)
	assert.NotNil(t, err)
	repo.AssertExpectations(t)
}
