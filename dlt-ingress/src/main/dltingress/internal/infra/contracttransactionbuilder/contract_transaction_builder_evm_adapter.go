package contracttransactionbuilder

import (
	"fmt"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"github.com/ethereum/go-ethereum/common"
	"github.com/umbracle/ethgo/abi"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type EvmContractTransactionBuilder struct {
	goethAbi *abi.ABI
}

func NewEvmContractTransactionBuilder(abiString string) (*EvmContractTransactionBuilder, error) {
	goethAbi, err := abi.NewABI(abiString)
	if err != nil {
		return nil, fmt.Errorf("error parsing smart contract abi: %w", err)
	}
	return &EvmContractTransactionBuilder{
		goethAbi: goethAbi,
	}, nil
}

func (ctb *EvmContractTransactionBuilder) EncodeCallData(methodName string, methodArgs map[string]any) ([]byte, error) {
	method := ctb.goethAbi.GetMethod(methodName)
	if method == nil {
		return nil, fmt.Errorf("error encoding EVM call data: method %s does not exist", methodName)
	}
	encoded, err := method.Encode(methodArgs)
	if err != nil {
		return nil, fmt.Errorf("error encoding EVM call data for method %s: %w", methodName, err)
	}
	return encoded, nil
}

func (ctb *EvmContractTransactionBuilder) BuildTransaction(request *BuildTransactionRequest) (*portcommon.TransactionResponse, error) {
	if !common.IsHexAddress(request.SmartContractId) {
		return nil, domainerrors.NewInvalidSmartContractAddressDomainError(request.SmartContractId)
	}

	encode, err := ctb.EncodeCallData(request.MethodName, request.MethodArgs)
	if err != nil {
		return nil, err
	}
	data := fmt.Sprintf("0x%x", encode)
	return &portcommon.TransactionResponse{
		EVMTransactionResponse: &portcommon.EVMTransactionResponse{
			TransactionType:      request.TransactionType,
			ChainId:              request.ChainId,
			Nonce:                request.Nonce,
			GasLimit:             request.GasLimit,
			GasPrice:             request.GasPrice,
			MaxPriorityFeePerGas: request.MaxPriorityFeePerGas,
			MaxFeePerGas:         request.MaxFeePerGas,
			Data:                 data,
			Value:                request.Value,
			To:                   request.SmartContractId,
		},
	}, nil
}

func (ctb *EvmContractTransactionBuilder) OverrideGasLimit(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newGasLimit *amount.Amount) (*portcommon.TransactionResponse, error) {
	evmResp := *originalResponse.EVMTransactionResponse
	evmResp.GasLimit = newGasLimit
	originalResponse.EVMTransactionResponse = &evmResp
	return &originalResponse, nil
}

func (ctb *EvmContractTransactionBuilder) OverrideCuPrice(_ BuildTransactionRequest, _ portcommon.TransactionResponse, _ *amount.Amount) (*portcommon.TransactionResponse, error) {
	return nil, fmt.Errorf("overriding compute unit price is not supported on EVM")
}
