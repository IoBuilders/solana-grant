package txservice

import (
	"context"
	"errors"
	"testing"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder"
	ctbmock "dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder/mock"
	mock2 "dlt-ingress/src/main/dltingress/internal/infra/evm/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/nonceprovider"
	npmock "dlt-ingress/src/main/dltingress/internal/infra/nonceprovider/mock"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

const nestedContractName = "Facets"
const nestedCallsArg = "_callData"

func expectEvmPreconditions(ctx context.Context, nonceProvider *npmock.NonceProviderMock, ethClientRegistry *mock2.EvmClientRegistryMock) *amount.Amount {
	nonce, _ := amount.NewFromString("0")
	nonceProvider.On("GetNonce", ctx, &nonceprovider.GetNonceRequest{
		NetworkId:    validNetworkId,
		DltAccountId: validSenderDltAccountId,
	}).Return(nonce, nil)

	networkGasPrice, _ := amount.NewFromString(validGasPrice)
	ethClient := new(mock2.EvmClientMock)
	ethClientRegistry.On("GetClientForNetworkId", ctx, validNetworkId).Return(ethClient, nil)
	ethClient.On("GetGasPrice", ctx).Return(networkGasPrice, nil)

	return nonce
}

func captureOuterRequest(
	registry *contracttransactionbuilder.Registry,
	nonce *amount.Amount,
) (*ctbmock.ContractTransactionBuilderMock, **contracttransactionbuilder.BuildTransactionRequest) {
	var captured *contracttransactionbuilder.BuildTransactionRequest
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.EVM, validSmartContractName, builder)
	builder.On("BuildTransaction", mock.Anything).
		Run(func(args mock.Arguments) {
			captured = args.Get(0).(*contracttransactionbuilder.BuildTransactionRequest)
		}).
		Return(validLegacyTxResponse(nonce), nil)
	return builder, &captured
}

func registerNestedBuilder(registry *contracttransactionbuilder.Registry) *ctbmock.ContractTransactionBuilderMock {
	builder := new(ctbmock.ContractTransactionBuilderMock)
	registry.Register(common.EVM, nestedContractName, builder)
	return builder
}

func TestDltIngressTxService_NestedCalls_EncodedInPlace(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	outerBuilder, captured := captureOuterRequest(txBuilderRegistry, nonce)

	nestedBuilder := registerNestedBuilder(txBuilderRegistry)
	nestedBuilder.On("EncodeCallData", "initializeA", map[string]any{"a": 1}).Return([]byte{0xAA}, nil)
	nestedBuilder.On("EncodeCallData", "initializeB", map[string]any{"b": 2}).Return([]byte{0xBB}, nil)

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA", MethodArgs: map[string]any{"a": 1}},
			{SmartContractName: nestedContractName, MethodName: "initializeB", MethodArgs: map[string]any{"b": 2}},
		},
	}

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	require.NotNil(t, *captured)
	assert.Equal(t, [][]byte{{0xAA}, {0xBB}}, (*captured).MethodArgs[nestedCallsArg])
	nestedBuilder.AssertExpectations(t)
	outerBuilder.AssertExpectations(t)
}

func TestDltIngressTxService_NestedCalls_PlainArgumentsArePreserved(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	_, captured := captureOuterRequest(txBuilderRegistry, nonce)

	nestedBuilder := registerNestedBuilder(txBuilderRegistry)
	nestedBuilder.On("EncodeCallData", mock.Anything, mock.Anything).Return([]byte{0xAA}, nil)

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		"_date":  "2026-09-21",
		"_owner": "0xabc",
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA"},
		},
	}

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	assert.Equal(t, "2026-09-21", (*captured).MethodArgs["_date"])
	assert.Equal(t, "0xabc", (*captured).MethodArgs["_owner"])
	assert.Equal(t, [][]byte{{0xAA}}, (*captured).MethodArgs[nestedCallsArg])
}

func TestDltIngressTxService_NestedCalls_DoesNotMutateCallerArgs(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	captureOuterRequest(txBuilderRegistry, nonce)

	nestedBuilder := registerNestedBuilder(txBuilderRegistry)
	nestedBuilder.On("EncodeCallData", mock.Anything, mock.Anything).Return([]byte{0xAA}, nil)

	callerArgs := map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA"},
		},
	}
	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = callerArgs

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	_, stillCalls := callerArgs[nestedCallsArg].([]portcommon.Invocation)
	assert.True(t, stillCalls, "the caller's map must be left untouched")
}

func TestDltIngressTxService_WithoutNestedCalls_ArgumentsUntouched(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	_, captured := captureOuterRequest(txBuilderRegistry, nonce)

	req := validRequest()
	req.MethodArgs = map[string]any{"_to": "0xabc"}

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	assert.Equal(t, map[string]any{"_to": "0xabc"}, (*captured).MethodArgs)
}

func TestDltIngressTxService_NestedCalls_SupportsSeveralArguments(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	_, captured := captureOuterRequest(txBuilderRegistry, nonce)

	nestedBuilder := registerNestedBuilder(txBuilderRegistry)
	nestedBuilder.On("EncodeCallData", "initializeA", mock.Anything).Return([]byte{0xAA}, nil)
	nestedBuilder.On("EncodeCallData", "initializeB", mock.Anything).Return([]byte{0xBB}, nil)

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		"_first":  []portcommon.Invocation{{SmartContractName: nestedContractName, MethodName: "initializeA"}},
		"_second": []portcommon.Invocation{{SmartContractName: nestedContractName, MethodName: "initializeB"}},
	}

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	assert.Equal(t, [][]byte{{0xAA}}, (*captured).MethodArgs["_first"])
	assert.Equal(t, [][]byte{{0xBB}}, (*captured).MethodArgs["_second"])
}

func TestDltIngressTxService_NestedCalls_UnknownNestedContract_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	txBuilderRegistry.Register(common.EVM, validSmartContractName, new(ctbmock.ContractTransactionBuilderMock))

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: "NotRegistered", MethodName: "initializeA"},
		},
	}

	result, err := svc.PrepareTransaction(ctx, req, network)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NotRegistered")
}

func TestDltIngressTxService_NestedCalls_EncodeError_IdentifiesTheCall(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	txBuilderRegistry.Register(common.EVM, validSmartContractName, new(ctbmock.ContractTransactionBuilderMock))
	nestedBuilder := registerNestedBuilder(txBuilderRegistry)
	nestedBuilder.On("EncodeCallData", "initializeA", mock.Anything).Return([]byte{0xAA}, nil)
	nestedBuilder.On("EncodeCallData", "initializeB", mock.Anything).Return(nil, errors.New("bad args"))

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA"},
			{SmartContractName: nestedContractName, MethodName: "initializeB"},
		},
	}

	result, err := svc.PrepareTransaction(ctx, req, network)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `nested call 1 of "_callData" (Facets.initializeB)`)
	assert.Contains(t, err.Error(), "bad args")
}

func TestDltIngressTxService_WithoutFlag_NestedCallsAreNotResolved(t *testing.T) {
	svc, nonceProvider, txBuilderRegistry, ethClientRegistry, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	nonce := expectEvmPreconditions(ctx, nonceProvider, ethClientRegistry)
	_, captured := captureOuterRequest(txBuilderRegistry, nonce)

	nestedBuilder := registerNestedBuilder(txBuilderRegistry)

	calls := []portcommon.Invocation{{SmartContractName: nestedContractName, MethodName: "initializeA"}}
	req := validRequest()
	req.MethodArgs = map[string]any{nestedCallsArg: calls}

	_, err := svc.PrepareTransaction(ctx, req, network)

	require.NoError(t, err)
	assert.Equal(t, calls, (*captured).MethodArgs[nestedCallsArg])
	nestedBuilder.AssertNotCalled(t, "EncodeCallData", mock.Anything, mock.Anything)
}

func TestDltIngressTxService_NestedCalls_SecondLevel_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	txBuilderRegistry.Register(common.EVM, validSmartContractName, new(ctbmock.ContractTransactionBuilderMock))
	nestedBuilder := registerNestedBuilder(txBuilderRegistry)

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA", MethodArgs: map[string]any{
				"_inner": []portcommon.Invocation{{SmartContractName: nestedContractName, MethodName: "initializeB"}},
			}},
		},
	}

	result, err := svc.PrepareTransaction(ctx, req, network)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one level of nesting is supported")
	nestedBuilder.AssertNotCalled(t, "EncodeCallData", mock.Anything, mock.Anything)
}

func TestDltIngressTxService_NestedCalls_SecondLevelInsideTuple_ReturnsError(t *testing.T) {
	svc, _, txBuilderRegistry, _, _, _, _ := newService()
	network := *config.DltIngressConfig.DltIngress.Networks[0]
	ctx := context.Background()

	txBuilderRegistry.Register(common.EVM, validSmartContractName, new(ctbmock.ContractTransactionBuilderMock))
	registerNestedBuilder(txBuilderRegistry)

	req := validRequest()
	req.ResolveNestedCalls = true
	req.MethodArgs = map[string]any{
		nestedCallsArg: []portcommon.Invocation{
			{SmartContractName: nestedContractName, MethodName: "initializeA", MethodArgs: map[string]any{
				"_config": map[string]any{
					"_inner": []portcommon.Invocation{{SmartContractName: nestedContractName, MethodName: "initializeB"}},
				},
			}},
		},
	}

	result, err := svc.PrepareTransaction(ctx, req, network)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only one level of nesting is supported")
}
