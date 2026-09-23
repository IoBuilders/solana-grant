package portcommon

import (
	"math/big"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

type TransactionResponse struct {
	*EVMTransactionResponse
	*SVMTransactionResponse
}
type EVMTransactionResponse struct {
	TransactionType      uint
	ChainId              *amount.Amount
	Nonce                *amount.Amount
	GasLimit             *amount.Amount
	GasPrice             *amount.Amount
	MaxPriorityFeePerGas *amount.Amount
	MaxFeePerGas         *amount.Amount
	Data                 string
	Value                *amount.Amount
	To                   string
}

// SVMMessageHeader mirrors Solana's compiled message header.
type SVMMessageHeader struct {
	NumRequiredSignatures       uint8
	NumReadonlySignedAccounts   uint8
	NumReadonlyUnsignedAccounts uint8
}

// SVMCompiledInstruction is an instruction as it appears in the compiled message,
// with account references expressed as indices into AccountKeys.
type SVMCompiledInstruction struct {
	ProgramIDIndex uint16
	AccountIndices []uint16
	Data           string // base64-encoded instruction data
}

// SVMTransactionResponse holds both the serialized transaction (for signing) and
// the decoded message fields (for persistence and retry).
type SVMTransactionResponse struct {
	SerializedTransaction string // base64-encoded unsigned Solana transaction
	FeePayer              string
	RecentBlockhash       string // base58-encoded; needed to check expiry on retry
	CuLimit               *amount.Amount
	CuPrice               *amount.Amount
	Header                SVMMessageHeader
	AccountKeys           []string // base58-encoded public keys, compiled-message order
	Instructions          []SVMCompiledInstruction
}

// ValueBigInt returns Value as a *big.Int, defaulting to zero when nil.
func (tx *EVMTransactionResponse) ValueBigInt() *big.Int {
	if tx.Value == nil {
		return new(big.Int)
	}
	return tx.Value.RawValue()
}

const (
	TransactionTypeLegacy     uint = 0
	TransactionTypeDynamicFee uint = 2
)
