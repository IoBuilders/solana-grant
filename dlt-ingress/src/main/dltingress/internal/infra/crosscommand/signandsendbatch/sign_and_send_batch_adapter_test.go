package signandsendbatchadapter

import (
	"context"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/app/command/signandsend"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	signandsendbatchcross "dlt-ingress/src/main/dltingress/port/crosscommand/signandsendbatch"
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

func validResponse() *signandsend.Response {
	nonce, _ := amount.NewFromString("1")
	gasLimit, _ := amount.NewFromString("21000")
	gasPrice, _ := amount.NewFromString("1000000000")

	return &signandsend.Response{
		TransactionSentEvent: transaction.TransactionSentEvent{
			TxId: validTxId,
			Dlt:  validDlt,
			EvmTransactionEventModel: &transaction.EvmTransactionEventModel{
				FromAddress: validSenderDltAccountId,
				ToAddress:   "0x456",
				Nonce:       *nonce,
				GasLimit:    *gasLimit,
				GasPrice:    *gasPrice,
			},
		},
	}
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_RequestsNestedCallResolution(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	commandBus.On("Dispatch", mock.Anything, mock.MatchedBy(func(cmd *signandsend.Command) bool {
		return cmd.ResolveNestedCalls
	})).Return(validResponse(), nil)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		SenderDltAccountId: validSenderDltAccountId,
		NetworkId:          validNetworkId,
		Calls: []signandsendbatchcross.BatchCall{
			{SmartContractName: "Facets", MethodName: "initializeFixedRate"},
		},
		Dispatch: &signandsendbatchcross.DispatchCall{
			SmartContractId:   "0xFactory",
			SmartContractName: "Factory",
			MethodName:        "batchInitializer",
			MethodArgs:        map[string]any{"_asset": "0xabc"},
			CallDataArgName:   "_initializersCallData",
		},
	})

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, validTxId, resp.TxId)
	assert.Equal(t, validDlt, resp.Dlt)
	commandBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_ComposesTheDispatchCall(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	var forwarded *signandsend.Command
	commandBus.On("Dispatch", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { forwarded = args.Get(1).(*signandsend.Command) }).
		Return(validResponse(), nil)

	adapter := NewCrossCommandAdapter(commandBus)
	_, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		NetworkId: validNetworkId,
		Calls: []signandsendbatchcross.BatchCall{
			{SmartContractName: "Facets", MethodName: "initializeFixedRate", MethodArgs: map[string]any{"a": 1}},
			{SmartContractName: "Facets", MethodName: "initializeCap", MethodArgs: map[string]any{"b": 2}},
		},
		Dispatch: &signandsendbatchcross.DispatchCall{
			SmartContractId:   "0xFactory",
			SmartContractName: "Factory",
			MethodName:        "batchInitializer",
			MethodArgs:        map[string]any{"_asset": "0xabc"},
			CallDataArgName:   "_initializersCallData",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "0xFactory", forwarded.SmartContractId)
	assert.Equal(t, "Factory", forwarded.SmartContractName)
	assert.Equal(t, "batchInitializer", forwarded.MethodName)
	assert.Equal(t, "0xabc", forwarded.MethodArgs["_asset"])
	assert.Equal(t, []portcommon.Invocation{
		{SmartContractName: "Facets", MethodName: "initializeFixedRate", MethodArgs: map[string]any{"a": 1}},
		{SmartContractName: "Facets", MethodName: "initializeCap", MethodArgs: map[string]any{"b": 2}},
	}, forwarded.MethodArgs["_initializersCallData"])
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_WithoutDispatch_ReturnsError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		NetworkId: validNetworkId,
		Calls:     []signandsendbatchcross.BatchCall{{SmartContractName: "Facets", MethodName: "initializeFixedRate"}},
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "dispatch call is required")
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_WithoutCallDataArgName_ReturnsError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		NetworkId: validNetworkId,
		Calls:     []signandsendbatchcross.BatchCall{{SmartContractName: "Facets", MethodName: "initializeFixedRate"}},
		Dispatch: &signandsendbatchcross.DispatchCall{
			SmartContractName: "Factory",
			MethodName:        "batchInitializer",
		},
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "CallDataArgName is required")
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_CallDataArgNameTaken_ReturnsError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		NetworkId: validNetworkId,
		Calls:     []signandsendbatchcross.BatchCall{{SmartContractName: "Facets", MethodName: "initializeFixedRate"}},
		Dispatch: &signandsendbatchcross.DispatchCall{
			SmartContractName: "Factory",
			MethodName:        "batchInitializer",
			MethodArgs:        map[string]any{"_initializersCallData": "already here"},
			CallDataArgName:   "_initializersCallData",
		},
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "reserved for the batched calls")
	commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_DltIngressError(t *testing.T) {
	commandBus := new(commandbusmock.CommandBusMock)

	commandBus.On("Dispatch", mock.Anything, mock.Anything).
		Return((*signandsend.Response)(nil), assert.AnError)

	adapter := NewCrossCommandAdapter(commandBus)
	resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
		NetworkId: validNetworkId,
		Calls: []signandsendbatchcross.BatchCall{
			{SmartContractName: "Facets", MethodName: "initializeFixedRate"},
		},
		Dispatch: &signandsendbatchcross.DispatchCall{
			SmartContractId:   "0xFactory",
			SmartContractName: "Factory",
			MethodName:        "batchInitializer",
			MethodArgs:        map[string]any{"_asset": "0xabc"},
			CallDataArgName:   "_initializersCallData",
		},
	})

	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, assert.AnError)
	commandBus.AssertExpectations(t)
}

func TestDltIngressSignAndSendBatchCrossCommandAdapter_Execute_WithoutCalls_ReturnsError(t *testing.T) {
	for name, calls := range map[string][]signandsendbatchcross.BatchCall{
		"unset": nil,
		"empty": {},
	} {
		t.Run(name, func(t *testing.T) {
			commandBus := new(commandbusmock.CommandBusMock)

			adapter := NewCrossCommandAdapter(commandBus)
			resp, err := adapter.Execute(context.Background(), &signandsendbatchcross.CrossCommand{
				NetworkId: validNetworkId,
				Calls:     calls,
				Dispatch: &signandsendbatchcross.DispatchCall{
					SmartContractName: "Factory",
					MethodName:        "batchInitializer",
					CallDataArgName:   "_initializersCallData",
				},
			})

			assert.Nil(t, resp)
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "no calls to batch")
			commandBus.AssertNotCalled(t, "Dispatch", mock.Anything, mock.Anything)
		})
	}
}
