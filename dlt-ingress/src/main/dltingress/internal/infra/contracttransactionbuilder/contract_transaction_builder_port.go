package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type Port interface {
	BuildTransaction(request *BuildTransactionRequest) (*portcommon.TransactionResponse, error)
	OverrideGasLimit(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newGasLimit *amount.Amount) (*portcommon.TransactionResponse, error)
	OverrideCuPrice(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newCuPrice *amount.Amount) (*portcommon.TransactionResponse, error)
	EncodeCallData(methodName string, methodArgs map[string]any) ([]byte, error)
}
type BuildTransactionRequest struct {
	SenderDltAccountId  string
	SignerDltAccountIds []string // For multiple signing
	SmartContractId     string
	MethodName          string
	MethodArgs          map[string]any
	*EVMBuildTransactionRequest
	*SVMBuildTransactionRequest
}

func NewEVMLegacyBuildTransactionRequest(
	senderDltAccountId string,
	SmartContractId string,
	methodName string,
	methodArgs map[string]any,
	chainId *amount.Amount,
	nonce *amount.Amount,
	gasLimit *amount.Amount,
	gasPrice *amount.Amount,
	value *amount.Amount,
) *BuildTransactionRequest {
	return &BuildTransactionRequest{
		SenderDltAccountId: senderDltAccountId,
		SmartContractId:    SmartContractId,
		MethodName:         methodName,
		MethodArgs:         methodArgs,
		EVMBuildTransactionRequest: &EVMBuildTransactionRequest{
			TransactionType: portcommon.TransactionTypeLegacy,
			ChainId:         chainId,
			Nonce:           nonce,
			GasLimit:        gasLimit,
			GasPrice:        gasPrice,
			Value:           value,
		},
	}
}

func NewEVMDynamicFeeBuildTransactionRequest(
	senderDltAccountId string,
	SmartContractId string,
	methodName string,
	methodArgs map[string]any,
	chainId *amount.Amount,
	nonce *amount.Amount,
	gasLimit *amount.Amount,
	maxPriorityFeePerGas *amount.Amount,
	maxFeePerGas *amount.Amount,
	value *amount.Amount,
) *BuildTransactionRequest {
	return &BuildTransactionRequest{
		SenderDltAccountId: senderDltAccountId,
		SmartContractId:    SmartContractId,
		MethodName:         methodName,
		MethodArgs:         methodArgs,
		EVMBuildTransactionRequest: &EVMBuildTransactionRequest{
			TransactionType:      portcommon.TransactionTypeDynamicFee,
			ChainId:              chainId,
			Nonce:                nonce,
			GasLimit:             gasLimit,
			MaxPriorityFeePerGas: maxPriorityFeePerGas,
			MaxFeePerGas:         maxFeePerGas,
			Value:                value,
		},
	}
}

type EVMBuildTransactionRequest struct {
	TransactionType      uint
	ChainId              *amount.Amount
	Nonce                *amount.Amount
	GasLimit             *amount.Amount
	GasPrice             *amount.Amount
	MaxPriorityFeePerGas *amount.Amount
	MaxFeePerGas         *amount.Amount
	Value                *amount.Amount
}

type SVMBuildTransactionRequest struct {
	RecentBlockHash string
	CuPrice         *amount.Amount
	CuLimit         *amount.Amount
}

func NewSVMBuildTransactionRequest(
	senderDltAccountId string,
	smartContractId string,
	methodName string,
	methodArgs map[string]any,
	recentBlockHash string,
	cuPrice *amount.Amount,
	cuLimit *amount.Amount,
) *BuildTransactionRequest {
	return &BuildTransactionRequest{
		SenderDltAccountId: senderDltAccountId,
		SmartContractId:    smartContractId,
		MethodName:         methodName,
		MethodArgs:         methodArgs,
		SVMBuildTransactionRequest: &SVMBuildTransactionRequest{
			RecentBlockHash: recentBlockHash,
			CuPrice:         cuPrice,
			CuLimit:         cuLimit,
		},
	}
}
