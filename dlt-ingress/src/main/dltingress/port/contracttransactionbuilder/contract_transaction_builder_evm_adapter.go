package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"fmt"

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

func (ctb *EvmContractTransactionBuilder) BuildTransaction(request *BuildTransactionRequest) (*portcommon.TransactionResponse, error) {
	method := ctb.goethAbi.GetMethod(request.MethodName)
	if method == nil {
		return nil, fmt.Errorf("error building transaction: method %s does not exist", request.MethodName)
	}
	encode, err := method.Encode(request.MethodArgs)
	if err != nil {
		return nil, fmt.Errorf("error encoding EVM transaction: %w", err)
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
