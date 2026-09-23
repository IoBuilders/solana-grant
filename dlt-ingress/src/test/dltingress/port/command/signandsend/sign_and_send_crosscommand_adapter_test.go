package signandsendcross

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/command/signandsend"
	transactionDomain "dlt-ingress/src/main/dltingress/domain/transaction"
	signandsendcross "dlt-ingress/src/main/dltingress/port/command/signandsend"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	commandbusmock "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/test/core/command"
)

const (
	validSenderDltAccountId = "0x123"
	validNetworkId          = "mainnet"
	validTxId               = "0x789"
	validDlt                = "EVM"
)

func TestDltIngressSignAndSendCrossCommandAdapter_Execute_Success(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	nonce, _ := amount.NewFromString("1")
	gasLimit, _ := amount.NewFromString("21000")
	gasPrice, _ := amount.NewFromString("1000000000")

	expectedResp := &signandsend.Response{
		TransactionSentEvent: transactionDomain.TransactionSentEvent{
			TxId: validTxId,
			Dlt:  validDlt,
			EvmTransactionEventModel: &transactionDomain.EvmTransactionEventModel{
				FromAddress: validSenderDltAccountId,
				ToAddress:   "0x456",
				Nonce:       *nonce,
				GasLimit:    *gasLimit,
				GasPrice:    *gasPrice,
			},
		},
	}

	commandBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(cmd *signandsend.Command) bool {
		return cmd.SenderDltAccountId == validSenderDltAccountId && cmd.NetworkId == validNetworkId
	})).Return(expectedResp, nil)

	adapter := signandsendcross.NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendcross.CrossCommand{
		SenderDltAccountId: validSenderDltAccountId,
		NetworkId:          validNetworkId,
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.TxId)
	assert.Equal(t, validDlt, resp.Dlt)
	assert.Equal(t, expectedResp.EvmTransactionEventModel.FromAddress, resp.FromAddress)
	assert.Equal(t, expectedResp.EvmTransactionEventModel.ToAddress, resp.ToAddress)
	commandBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendCrossCommandAdapter_Execute_DltIngressError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	commandBus.On("Dispatch", mock.Anything, mock.Anything).
		Return((*signandsend.Response)(nil), assert.AnError)

	adapter := signandsendcross.NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendcross.CrossCommand{
		NetworkId: validNetworkId,
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	commandBus.AssertExpectations(t)
}
