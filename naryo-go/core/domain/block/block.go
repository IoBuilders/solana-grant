package block

import "time"

// SolanaBlock mirrors the result of the Solana getBlock RPC call.
type SolanaBlock struct {
	Slot              uint64
	Blockhash         string
	PreviousBlockhash string
	ParentSlot        uint64
	BlockHeight       *uint64    // nullable in RPC
	BlockTime         *time.Time // nullable in RPC
	Transactions      []SolanaTransaction
}

// SolanaTransaction mirrors a single transaction within a Solana getBlock RPC
// response.
type SolanaTransaction struct {
	Signature string
	// Accounts is the transaction message's full account list, as returned
	// by the RPC response — distinct from any single Instruction's Accounts,
	// which only lists the accounts that instruction references.
	Accounts     []string
	Instructions []Instruction
	Logs         []string
	Err          *string
}

// Instruction is a single instruction within a SolanaTransaction.
type Instruction struct {
	ProgramID string
	Data      []byte
	Accounts  []string
	Inner     bool
}
