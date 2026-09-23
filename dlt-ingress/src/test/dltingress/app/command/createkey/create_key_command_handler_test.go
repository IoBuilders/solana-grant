package createkey

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/common"
	"testing"

	"dlt-ingress/src/main/dltingress/app/command/createkey"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/custody"
	custodymocks "dlt-ingress/src/test/dltingress/port/custody"
	repomocks "dlt-ingress/src/test/dltingress/port/repository"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
	testbus "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/event"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validDltAccountId    = "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"
	validExternalId      = "ext-id-123"
	validDlt             = string(common.EVM)
	validCustodyProvider = string(custodykey.CustodyProviderDFNS)
)

func newHandler() (*createkey.CommandHandler, *repomocks.CustodyKeyRepositoryMock, *custodymocks.CustodyProviderMock, *testbus.EventBusMock) {
	repo := new(repomocks.CustodyKeyRepositoryMock)
	provider := new(custodymocks.CustodyProviderMock)
	eventBus := new(testbus.EventBusMock)
	return createkey.NewCommandHandler(repo, provider, eventBus), repo, provider, eventBus
}

func TestDltIngressCreateKeyCommandHandler_Handle_Success(t *testing.T) {
	handler, repo, provider, eventBus := newHandler()

	provider.On("CreateKey", mock.Anything, mock.AnythingOfType("*custody.CreateKeyRequest")).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(nil)

	cmd := createkey.Command{Dlt: validDlt, CustodyProvider: validCustodyProvider}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validDlt, resp.Dlt)
	assert.Equal(t, validCustodyProvider, resp.CustodyProvider)
	provider.AssertExpectations(t)
	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressCreateKeyCommandHandler_Handle_InvalidDlt(t *testing.T) {
	handler, repo, provider, _ := newHandler()

	cmd := createkey.Command{Dlt: "INVALID_DLT", CustodyProvider: validCustodyProvider}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidDlt)
	provider.AssertNotCalled(t, "CreateKey", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestDltIngressCreateKeyCommandHandler_Handle_InvalidCustodyProvider(t *testing.T) {
	handler, repo, provider, _ := newHandler()

	cmd := createkey.Command{Dlt: validDlt, CustodyProvider: "INVALID_PROVIDER"}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidCustodyProvider)
	provider.AssertNotCalled(t, "CreateKey", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestDltIngressCreateKeyCommandHandler_Handle_CustodyProviderError(t *testing.T) {
	handler, repo, provider, eventBus := newHandler()

	provider.On("CreateKey", mock.Anything, mock.Anything).
		Return((*custody.KeyResponse)(nil), assert.AnError)

	cmd := createkey.Command{Dlt: validDlt, CustodyProvider: validCustodyProvider}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestDltIngressCreateKeyCommandHandler_Handle_EventBusError(t *testing.T) {
	handler, repo, provider, eventBus := newHandler()

	provider.On("CreateKey", mock.Anything, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	repo.On("Save", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, mock.Anything).Return(assert.AnError)

	cmd := createkey.Command{Dlt: validDlt, CustodyProvider: validCustodyProvider}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	provider.AssertExpectations(t)
	repo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestDltIngressCreateKeyCommandHandler_Handle_RepositoryError(t *testing.T) {
	handler, repo, provider, eventBus := newHandler()

	provider.On("CreateKey", mock.Anything, mock.Anything).
		Return(&custody.KeyResponse{DltAccountId: validDltAccountId, ExternalId: validExternalId}, nil)
	repo.On("Save", mock.Anything, mock.Anything).
		Return(assert.AnError)

	cmd := createkey.Command{Dlt: validDlt, CustodyProvider: validCustodyProvider}
	resp, err := handler.Handle(context.Background(), &cmd)

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	eventBus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
	provider.AssertExpectations(t)
	repo.AssertExpectations(t)
}
