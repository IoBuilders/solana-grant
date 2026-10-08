package buildtransactionadapter

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/app/command/buildtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	buildtransactioncross "dlt-ingress/src/main/dltingress/port/crosscommand/buildtransaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	commandbusmock "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/command"
)

const (
	validSenderDltAccountId = "0x123"
	validNetworkId          = "mainnet"
	validPayload            = `{"data":"0x","to":"0x456"}`
)

func TestDltIngressBuildTransactionCrossCommandAdapter_Execute_Success(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	expectedResp := &buildtransaction.Response{
		TransactionBuiltEvent: transaction.TransactionBuiltEvent{
			DltAccountId: validSenderDltAccountId,
			Payload:      validPayload,
		},
	}

	commandBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(cmd *buildtransaction.Command) bool {
		return cmd.SenderDltAccountId == validSenderDltAccountId && cmd.NetworkId == validNetworkId
	})).Return(expectedResp, nil)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &buildtransactioncross.CrossCommand{
		SenderDltAccountId: validSenderDltAccountId,
		NetworkId:          validNetworkId,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedResp.DltAccountId, resp.DltAccountId)
	assert.Equal(t, expectedResp.Payload, resp.Payload)
	commandBus.AssertExpectations(t)
}

func TestDltIngressBuildTransactionCrossCommandAdapter_Execute_DltIngressError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	commandBus.On("Dispatch", mock.Anything, mock.Anything).
		Return((*buildtransaction.Response)(nil), assert.AnError)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &buildtransactioncross.CrossCommand{
		NetworkId: validNetworkId,
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	commandBus.AssertExpectations(t)
}
