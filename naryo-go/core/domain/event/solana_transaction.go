package event

import (
	"github.com/google/uuid"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

// SolanaTransaction is a raw transaction as it appears inside a Solana
// block, before any TransactionFilter has matched it. It deliberately does
// not implement TransactionEvent: it is not yet a dispatched/persisted
// event, only the BlockTransaction view BlockEvent.Transactions() exposes.
type SolanaTransaction struct {
	nodeID uuid.UUID

	Signature string
	Slot      uint64
	// Accounts is the transaction message's full account list, as returned
	// by the RPC response — distinct from any single InstructionData's
	// Accounts, which only lists the accounts that instruction references.
	Accounts     []string
	Instructions []InstructionData
	Logs         []string
	Err          *string
	// DecodedReason is the human-readable reason for a failed transaction,
	// computed by SolanaTransactionProcessorPermanentTrigger via
	// DecodedError before dispatch. It is nil until WithDecodedReason sets
	// it -- NewSolanaTransactionEvent never populates it.
	DecodedReason *string
}

// InstructionData is a single instruction within a transaction.
type InstructionData struct {
	ProgramID string
	Data      []byte
	Accounts  []string
	Inner     bool
}

func NewSolanaTransaction(nodeID uuid.UUID, signature string, slot uint64, instructions []InstructionData, accounts []string, logs []string, err *string) (SolanaTransaction, error) {
	if nodeID == uuid.Nil {
		return SolanaTransaction{}, domainerrors.NewEmptyFieldError("NodeID", "SolanaTransaction")
	}
	if signature == "" {
		return SolanaTransaction{}, domainerrors.NewEmptyFieldError("Signature", "SolanaTransaction")
	}
	if instructions == nil {
		instructions = []InstructionData{}
	}
	if accounts == nil {
		accounts = []string{}
	}
	if logs == nil {
		logs = []string{}
	}
	return SolanaTransaction{
		nodeID:       nodeID,
		Signature:    signature,
		Slot:         slot,
		Accounts:     accounts,
		Instructions: instructions,
		Logs:         logs,
		Err:          err,
	}, nil
}

func (t SolanaTransaction) NodeID() uuid.UUID {
	return t.nodeID
}

func (t SolanaTransaction) Id() string {
	return t.Signature
}

func (t SolanaTransaction) HasError() bool {
	return t.Err != nil
}

// ToSolanaTransactionEvent converts t into the SolanaTransactionEvent
// dispatched for it once a TransactionFilter has matched it. t is already
// fully validated, so this conversion cannot fail.
func (t SolanaTransaction) ToSolanaTransactionEvent() SolanaTransactionEvent {
	return SolanaTransactionEvent{SolanaTransaction: t}
}

var _ BlockTransaction = SolanaTransaction{}

// SolanaTransactionEvent is the Solana implementation of TransactionEvent:
// a SolanaTransaction that has matched an active TransactionFilter and is
// being dispatched/persisted.
type SolanaTransactionEvent struct {
	SolanaTransaction
}

func NewSolanaTransactionEvent(nodeID uuid.UUID, signature string, slot uint64, instructions []InstructionData, accounts []string, logs []string, err *string) (SolanaTransactionEvent, error) {
	tx, txErr := NewSolanaTransaction(nodeID, signature, slot, instructions, accounts, logs, err)
	if txErr != nil {
		return SolanaTransactionEvent{}, txErr
	}
	return tx.ToSolanaTransactionEvent(), nil
}

func (t SolanaTransactionEvent) EventType() Type {
	return TypeTransaction
}

func (t SolanaTransactionEvent) HasError() bool {
	return t.Err != nil
}

// WithDecodedReason returns a copy of t with DecodedReason set to reason.
func (t SolanaTransactionEvent) WithDecodedReason(reason string) SolanaTransactionEvent {
	t.DecodedReason = &reason
	return t
}

var _ TransactionEvent = SolanaTransactionEvent{}
