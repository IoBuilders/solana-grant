package contracttransactionbuilder

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"math"

	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/contracttransactionbuilder/svm"
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"

	"github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type SvmIdlTransactionBuilder struct {
	idl         *svmidl.Idl
	programID   solana.PublicKey
	pdaResolver *svmidl.PdaResolver
}

func NewSvmIdlTransactionBuilder(idlJSON []byte) (*SvmIdlTransactionBuilder, error) {
	idl, err := svmidl.ParseIdl(idlJSON)
	if err != nil {
		return nil, err
	}
	programID, err := solana.PublicKeyFromBase58(idl.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid IDL address %q: %w", idl.Address, err)
	}
	return &SvmIdlTransactionBuilder{
		idl:         idl,
		programID:   programID,
		pdaResolver: svmidl.NewPdaResolver(programID),
	}, nil
}

func (b *SvmIdlTransactionBuilder) EncodeCallData(methodName string, _ map[string]any) ([]byte, error) {
	return nil, fmt.Errorf("encoding call data for nesting is not supported on SVM (method %s): batch calls must be emitted as transaction instructions", methodName)
}

func (b *SvmIdlTransactionBuilder) validateSmartContractId(smartContractId string) error {
	if smartContractId == "" {
		return nil
	}
	if id, err := solana.PublicKeyFromBase58(smartContractId); err != nil || !id.Equals(b.programID) {
		return domainerrors.NewSmartContractIdMismatchDomainError(smartContractId, b.programID.String())
	}
	return nil
}

func (b *SvmIdlTransactionBuilder) BuildTransaction(req *BuildTransactionRequest) (*portcommon.TransactionResponse, error) {
	if err := b.validateSmartContractId(req.SmartContractId); err != nil {
		return nil, err
	}

	instr, err := b.idl.FindInstruction(req.MethodName)
	if err != nil {
		return nil, err
	}

	sender, blockhash, instructions, err := buildSvmCommonTransactionData(req)
	if err != nil {
		return nil, err
	}

	accounts, err := b.pdaResolver.Resolve(instr, req.MethodArgs, sender)
	if err != nil {
		return nil, err
	}

	data, err := b.encodeInstructionData(instr, req.MethodArgs)
	if err != nil {
		return nil, err
	}

	instructions = append(instructions, solana.NewInstruction(b.programID, accounts, data))

	tx, err := solana.NewTransaction(instructions, blockhash, solana.TransactionPayer(sender))
	if err != nil {
		return nil, fmt.Errorf("build transaction: %w", err)
	}

	return buildSVMResponse(tx, sender, req.RecentBlockHash, req.CuLimit, req.CuPrice)
}

func (b *SvmIdlTransactionBuilder) OverrideGasLimit(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newGasLimit *amount.Amount) (*portcommon.TransactionResponse, error) {
	inner := *originalRequest.SVMBuildTransactionRequest
	inner.CuLimit = newGasLimit
	originalRequest.SVMBuildTransactionRequest = &inner
	return b.BuildTransaction(&originalRequest)
}

func (b *SvmIdlTransactionBuilder) OverrideCuPrice(originalRequest BuildTransactionRequest, originalResponse portcommon.TransactionResponse, newCuPrice *amount.Amount) (*portcommon.TransactionResponse, error) {
	inner := *originalRequest.SVMBuildTransactionRequest
	inner.CuPrice = newCuPrice
	originalRequest.SVMBuildTransactionRequest = &inner
	return b.BuildTransaction(&originalRequest)
}

func buildSvmCommonTransactionData(req *BuildTransactionRequest) (sender solana.PublicKey, blockhash solana.Hash, instructions []solana.Instruction, err error) {
	if req.SVMBuildTransactionRequest == nil {
		err = fmt.Errorf("SVMBuildTransactionRequest is required")
		return
	}

	sender, err = solana.PublicKeyFromBase58(req.SenderDltAccountId)
	if err != nil {
		err = fmt.Errorf("invalid sender %q: %w", req.SenderDltAccountId, err)
		return
	}

	blockhash, err = solana.HashFromBase58(req.RecentBlockHash)
	if err != nil {
		err = fmt.Errorf("invalid recent blockhash %q: %w", req.RecentBlockHash, err)
		return
	}

	instructions = []solana.Instruction{}
	if req.CuLimit != nil && !req.CuLimit.IsZero() {
		cuLimit := req.CuLimit.RawValue().Uint64()
		if cuLimit > math.MaxUint32 {
			err = fmt.Errorf("compute unit limit %d exceeds the maximum representable value (%d)", cuLimit, uint32(math.MaxUint32))
			return
		}
		instructions = append(instructions, computebudget.NewSetComputeUnitLimitInstruction(uint32(cuLimit)).Build())
	}
	if req.CuPrice != nil && !req.CuPrice.IsZero() {
		instructions = append(instructions, computebudget.NewSetComputeUnitPriceInstruction(req.CuPrice.RawValue().Uint64()).Build())
	}
	return
}

func (b *SvmIdlTransactionBuilder) encodeInstructionData(instr *svmidl.Instruction, args map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(instr.Discriminator); err != nil {
		return nil, fmt.Errorf("write discriminator: %w", err)
	}
	enc := bin.NewBorshEncoder(&buf)
	for _, arg := range instr.Args {
		v, present := args[arg.Name]
		if !present && arg.Type.Kind != svmidl.KindOption {
			return nil, fmt.Errorf("missing arg %q", arg.Name)
		}
		if err := svmidl.BorshEncode(arg.Type, v, enc, b.idl); err != nil {
			return nil, fmt.Errorf("encode arg %q: %w", arg.Name, err)
		}
	}
	return buf.Bytes(), nil
}

func buildSVMResponse(tx *solana.Transaction, feePayer solana.PublicKey, blockhash string, cuLimit *amount.Amount, cuPrice *amount.Amount) (*portcommon.TransactionResponse, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("serialize transaction: %w", err)
	}
	if len(raw) > svmtransaction.MaxSvmTransactionSize {
		return nil, fmt.Errorf("serialized transaction size %d bytes exceeds the %d byte limit", len(raw), svmtransaction.MaxSvmTransactionSize)
	}
	serialized := base64.StdEncoding.EncodeToString(raw)

	accountKeys := make([]string, len(tx.Message.AccountKeys))
	for i, ak := range tx.Message.AccountKeys {
		accountKeys[i] = ak.String()
	}

	instructions := make([]portcommon.SVMCompiledInstruction, len(tx.Message.Instructions))
	for i, ci := range tx.Message.Instructions {
		instructions[i] = portcommon.SVMCompiledInstruction{
			ProgramIDIndex: ci.ProgramIDIndex,
			AccountIndices: ci.Accounts,
			Data:           base64.StdEncoding.EncodeToString(ci.Data),
		}
	}

	return &portcommon.TransactionResponse{
		SVMTransactionResponse: &portcommon.SVMTransactionResponse{
			SerializedTransaction: serialized,
			FeePayer:              feePayer.String(),
			RecentBlockhash:       blockhash,
			CuLimit:               cuLimit,
			CuPrice:               cuPrice,
			Header: portcommon.SVMMessageHeader{
				NumRequiredSignatures:       tx.Message.Header.NumRequiredSignatures,
				NumReadonlySignedAccounts:   tx.Message.Header.NumReadonlySignedAccounts,
				NumReadonlyUnsignedAccounts: tx.Message.Header.NumReadonlyUnsignedAccounts,
			},
			AccountKeys:  accountKeys,
			Instructions: instructions,
		},
	}, nil
}
