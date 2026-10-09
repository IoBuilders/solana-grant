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

// WritableAccountKeys returns the accounts the transaction write-locks, following Solana's compiled-message
// ordering: signed writable, signed readonly, unsigned writable, unsigned readonly. Header counts that do not
// fit AccountKeys are clamped rather than trusted, so a malformed header yields fewer keys instead of a panic.
func (tx *SVMTransactionResponse) WritableAccountKeys() []string {
	total := len(tx.AccountKeys)
	signed := min(int(tx.Header.NumRequiredSignatures), total)
	signedWritable := max(signed-int(tx.Header.NumReadonlySignedAccounts), 0)
	unsignedWritableEnd := max(total-int(tx.Header.NumReadonlyUnsignedAccounts), signed)

	writable := make([]string, 0, signedWritable+unsignedWritableEnd-signed)
	writable = append(writable, tx.AccountKeys[:signedWritable]...)
	writable = append(writable, tx.AccountKeys[signed:unsignedWritableEnd]...)
	return writable
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
