package contractcallbuilder

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/umbracle/ethgo/abi"
)

type EvmContractCallBuilder struct {
	contractAbi *abi.ABI
}

func NewEvmContractCallBuilder(abiString string) (*EvmContractCallBuilder, error) {
	contractAbi, err := abi.NewABI(abiString)
	if err != nil {
		return nil, fmt.Errorf("error parsing smart contract contractAbi: %w", err)
	}
	return &EvmContractCallBuilder{contractAbi: contractAbi}, nil
}

func (b *EvmContractCallBuilder) BuildCall(request *BuildCallRequest) (*CallData, error) {
	method := b.contractAbi.GetMethod(request.MethodName)
	if method == nil {
		return nil, fmt.Errorf("error building call: method %s does not exist", request.MethodName)
	}
	encoded, err := method.Encode(request.MethodArgs)
	if err != nil {
		return nil, fmt.Errorf("error encoding EVM call: %w", err)
	}
	return &CallData{
		To:   request.SmartContractId,
		Data: fmt.Sprintf("0x%x", encoded),
	}, nil
}

func (b *EvmContractCallBuilder) DecodeResult(methodName string, result string) (map[string]any, error) {
	rawBytes, err := hexutil.Decode(result)
	if err != nil {
		return nil, fmt.Errorf("error decoding call result hex: %w", err)
	}

	method := b.contractAbi.GetMethod(methodName)
	if method == nil {
		return nil, fmt.Errorf("error decoding result: method %s does not exist", methodName)
	}

	decoded, err := method.Outputs.Decode(rawBytes)
	if err != nil {
		return nil, fmt.Errorf("error decoding EVM call result: %w", err)
	}

	if m, ok := decoded.(map[string]interface{}); ok {
		return m, nil
	}

	return map[string]any{"": decoded}, nil
}

var _ Port = (*EvmContractCallBuilder)(nil)
