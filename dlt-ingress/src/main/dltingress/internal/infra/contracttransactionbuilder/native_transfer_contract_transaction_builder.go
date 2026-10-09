package contracttransactionbuilder

import (
	"fmt"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type NativeTransferContractTransactionBuilder struct {
}

func (ctb *NativeTransferContractTransactionBuilder) EncodeCallData(_ string, _ map[string]any) ([]byte, error) {
	return nil, fmt.Errorf("encoding call data is not supported for native transfers")
}

func (ctb *NativeTransferContractTransactionBuilder) BuildTransaction(req *BuildTransactionRequest) (*portcommon.TransactionResponse, error) {
	argsLength := len(req.MethodArgs)
	if argsLength != 2 {
		return nil, domainerrors.NewInvalidNativeTransferArgsLength(argsLength)
	}
	to, okTo := req.MethodArgs["to"].(string)
	transferAmount, okTransferAmount := req.MethodArgs["amount"].(string)
	if !okTo || !okTransferAmount || to == "" || transferAmount == "" {
		return nil, domainerrors.NewInvalidNativeTransferArgs(req.MethodArgs)
	}
	transferAmountValue, err := amount.NewFromString(transferAmount)
	if err != nil {
		return nil, domainerrors.NewInvalidNativeTransferAmount(transferAmount)
	}
	if req.EVMBuildTransactionRequest != nil {
		return &portcommon.TransactionResponse{
			EVMTransactionResponse: &portcommon.EVMTransactionResponse{
				TransactionType:      req.TransactionType,
				ChainId:              req.ChainId,
				Nonce:                req.Nonce,
				GasLimit:             req.GasLimit,
				GasPrice:             req.GasPrice,
				MaxPriorityFeePerGas: req.MaxPriorityFeePerGas,
				MaxFeePerGas:         req.MaxFeePerGas,
				Data:                 "0x",
				Value:                transferAmountValue,
				To:                   to,
			},
		}, nil
	}

	sender, blockhash, instructions, err := buildSvmCommonTransactionData(req)
	if err != nil {
		return nil, err
	}
	toPubKey, err := solana.PublicKeyFromBase58(to)
	if err != nil {
		return nil, domainerrors.NewInvalidNativeTransferTo(to)
	}
	nativeTransferInstruction := system.NewTransferInstruction(
		transferAmountValue.RawValue().Uint64(),
		sender,
		toPubKey,
	).Build()

	instructions = append(instructions, nativeTransferInstruction)

	tx, err := solana.NewTransaction(instructions, blockhash, solana.TransactionPayer(sender))
	if err != nil {
		return nil, fmt.Errorf("build transaction: %w", err)
	}

	return buildSVMResponse(tx, sender, req.RecentBlockHash, req.CuLimit, req.CuPrice)
}

func (ctb *NativeTransferContractTransactionBuilder) OverrideGasLimit(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newGasLimit *amount.Amount) (*portcommon.TransactionResponse, error) {
	if originalRequest.EVMBuildTransactionRequest != nil {
		inner := *originalRequest.EVMBuildTransactionRequest
		inner.GasLimit = newGasLimit
		originalRequest.EVMBuildTransactionRequest = &inner
	} else {
		inner := *originalRequest.SVMBuildTransactionRequest
		inner.CuLimit = newGasLimit
		originalRequest.SVMBuildTransactionRequest = &inner
	}
	return ctb.BuildTransaction(&originalRequest)
}

func (ctb *NativeTransferContractTransactionBuilder) OverrideCuPrice(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newCuPrice *amount.Amount) (*portcommon.TransactionResponse, error) {
	if originalRequest.SVMBuildTransactionRequest == nil {
		return nil, fmt.Errorf("overriding compute unit price is not supported on EVM")
	}
	inner := *originalRequest.SVMBuildTransactionRequest
	inner.CuPrice = newCuPrice
	originalRequest.SVMBuildTransactionRequest = &inner
	return ctb.BuildTransaction(&originalRequest)
}
