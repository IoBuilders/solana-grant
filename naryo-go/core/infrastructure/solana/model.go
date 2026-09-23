package solana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mr-tron/base58"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
)

// rpcBlock is the getBlock result shape for encoding=jsonParsed, rewards=false.
type rpcBlock struct {
	BlockHeight       *uint64          `json:"blockHeight"`
	BlockTime         *int64           `json:"blockTime"`
	Blockhash         string           `json:"blockhash"`
	ParentSlot        uint64           `json:"parentSlot"`
	PreviousBlockhash string           `json:"previousBlockhash"`
	Transactions      []rpcTransaction `json:"transactions"`
}

type rpcTransaction struct {
	Meta        rpcTransactionMeta     `json:"meta"`
	Transaction rpcTransactionEnvelope `json:"transaction"`
}

type rpcTransactionEnvelope struct {
	Signatures []string          `json:"signatures"`
	Message    rpcTransactionMsg `json:"message"`
}

type rpcTransactionMsg struct {
	AccountKeys  []rpcAccountKey  `json:"accountKeys"`
	Instructions []rpcInstruction `json:"instructions"`
}

// rpcAccountKey is one accountKeys entry in the jsonParsed message format
// (an object, unlike the plain pubkey strings of the json/base64 formats).
type rpcAccountKey struct {
	Pubkey string `json:"pubkey"`
}

type rpcTransactionMeta struct {
	// Err is left raw: on success the RPC sends the JSON literal null,
	// otherwise an error object whose shape depends on the failure. Unlike
	// unmarshalling into a pointer, json.RawMessage captures "null" verbatim
	// rather than leaving the field nil, so toDomain checks for it explicitly.
	Err               json.RawMessage    `json:"err"`
	LogMessages       []string           `json:"logMessages"`
	InnerInstructions []innerInstruction `json:"innerInstructions"`
}

// rpcInstruction is one instruction in the jsonParsed message format, which
// takes one of two shapes depending on whether the program is recognized:
//   - ParsedInstruction (recognized program): Parsed is set, Data/Accounts are not.
//   - PartiallyDecodedInstruction (unrecognized program): Data/Accounts are
//     set, Parsed is not.
type rpcInstruction struct {
	ProgramID string                `json:"programId"`
	Parsed    *rpcParsedInstruction `json:"parsed"`
	Data      string                `json:"data,omitempty"`
	Accounts  []string              `json:"accounts,omitempty"`
}

// rpcParsedInstruction is the parsed shape of a recognized-program
// instruction. Only Info is read, and only to pull out the accounts it
// references (accountsFromParsedInfo) — its instruction-specific meaning
// (the Type field, and what each Info field represents) is not decoded.
//
// Not every recognized program parses to an {info: ...} object, though: the
// Memo program (among others) parses to a bare JSON string instead.
// UnmarshalJSON tolerates that by leaving Info unset — a string has no
// accounts to recover — rather than failing to decode the whole block.
type rpcParsedInstruction struct {
	Info json.RawMessage `json:"info"`
}

type innerInstruction struct {
	Index        uint8            `json:"index"`
	Instructions []rpcInstruction `json:"instructions"`
}

func (p *rpcParsedInstruction) UnmarshalJSON(data []byte) error {
	if trimmed := bytes.TrimSpace(data); len(trimmed) == 0 || trimmed[0] != '{' {
		return nil
	}
	type alias rpcParsedInstruction
	return json.Unmarshal(data, (*alias)(p))
}

// toDomain maps ri to a block.Instruction. ProgramID identifies the
// instruction either way. A recognized-program instruction has no flat
// accounts list of its own — Solana only exposes its touched accounts as
// named fields inside Info, shaped differently per instruction type — so
// Accounts is recovered by scanning Info for pubkey-shaped strings rather
// than decoding what the instruction does; Data is left empty, since the
// node never sends raw instruction bytes alongside a parsed one. An
// unrecognized-program instruction's base58 Data is decoded as-is alongside
// its own flat Accounts list — except when empty, which some inner
// instructions legitimately send (base58.Decode errors on "" as invalid
// input, not as a valid zero-length value).
func (ri rpcInstruction) toDomain(inner bool) (block.Instruction, error) {
	if ri.Parsed != nil {
		accounts, err := accountsFromParsedInfo(ri.Parsed.Info)
		if err != nil {
			return block.Instruction{}, err
		}
		return block.Instruction{ProgramID: ri.ProgramID, Accounts: accounts, Inner: inner}, nil
	}

	var data []byte
	if ri.Data != "" {
		var err error
		data, err = base58.Decode(ri.Data)
		if err != nil {
			return block.Instruction{}, fmt.Errorf("solana rpc: decode instruction data: %w", err)
		}
	}
	return block.Instruction{
		ProgramID: ri.ProgramID,
		Data:      data,
		Accounts:  ri.Accounts,
		Inner:     inner,
	}, nil
}

// accountsFromParsedInfo recovers the accounts a recognized-program
// instruction touches from its parsed Info object, without knowing that
// instruction type's field names: every account field Solana's own parsers
// emit is a base58-encoded 32-byte pubkey, and nothing else in Info decodes
// to exactly that length, so collecting every pubkey-shaped string found
// anywhere in Info (however deeply nested) recovers the account list
// generically, across every program Solana can parse. The result is
// deduplicated and sorted: Info's field order does not carry the
// positional/role significance a PartiallyDecodedInstruction's accounts
// list has, so there is no meaningful order to preserve.
func accountsFromParsedInfo(info json.RawMessage) ([]string, error) {
	if len(info) == 0 || string(info) == "null" {
		return nil, nil
	}

	var value any
	if err := json.Unmarshal(info, &value); err != nil {
		return nil, fmt.Errorf("solana rpc: unmarshal parsed instruction info: %w", err)
	}

	seen := map[string]struct{}{}
	collectPubkeys(value, seen)

	accounts := make([]string, 0, len(seen))
	for account := range seen {
		accounts = append(accounts, account)
	}
	sort.Strings(accounts)
	return accounts, nil
}

// collectPubkeys recursively visits value — the result of unmarshalling
// arbitrary JSON into `any` — collecting every string that decodes as
// base58 to exactly 32 bytes, the fixed size of a Solana pubkey.
func collectPubkeys(value any, seen map[string]struct{}) {
	switch v := value.(type) {
	case string:
		if decoded, err := base58.Decode(v); err == nil && len(decoded) == 32 {
			seen[v] = struct{}{}
		}
	case map[string]any:
		for _, nested := range v {
			collectPubkeys(nested, seen)
		}
	case []any:
		for _, nested := range v {
			collectPubkeys(nested, seen)
		}
	}
}

// toDomain maps rt to a block.SolanaTransaction. Signature is the
// transaction's first (fee payer's) signature.
func (rt rpcTransaction) toDomain() (block.SolanaTransaction, error) {
	var signature string
	if len(rt.Transaction.Signatures) > 0 {
		signature = rt.Transaction.Signatures[0]
	}

	accounts := make([]string, len(rt.Transaction.Message.AccountKeys))
	for i, key := range rt.Transaction.Message.AccountKeys {
		accounts[i] = key.Pubkey
	}

	// Capacity only accounts for the top-level instructions: len(InnerInstructions) is the
	// number of inner-instruction *groups*, not the total individual inner instructions each
	// group holds, so it can't be used to preallocate their share up front. append grows the
	// slice for those instead, once their actual total is known.
	instructions := make([]block.Instruction, 0, len(rt.Transaction.Message.Instructions))
	for _, ri := range rt.Transaction.Message.Instructions {
		instr, err := ri.toDomain(false)
		if err != nil {
			return block.SolanaTransaction{}, err
		}
		instructions = append(instructions, instr)
	}

	for _, in := range rt.Meta.InnerInstructions {
		for _, ri := range in.Instructions {
			instr, err := ri.toDomain(true)
			if err != nil {
				return block.SolanaTransaction{}, err
			}
			instructions = append(instructions, instr)
		}
	}

	var txErr *string
	if len(rt.Meta.Err) > 0 && string(rt.Meta.Err) != "null" {
		s := string(rt.Meta.Err)
		txErr = &s
	}

	return block.SolanaTransaction{
		Signature:    signature,
		Accounts:     accounts,
		Instructions: instructions,
		Logs:         rt.Meta.LogMessages,
		Err:          txErr,
	}, nil
}

// toDomain maps rb, the getBlock result for slot, to a block.SolanaBlock.
func (rb rpcBlock) toDomain(slot uint64) (*block.SolanaBlock, error) {
	var blockTime *time.Time
	if rb.BlockTime != nil {
		t := time.Unix(*rb.BlockTime, 0).UTC()
		blockTime = &t
	}

	transactions := make([]block.SolanaTransaction, len(rb.Transactions))
	for i, rt := range rb.Transactions {
		tx, err := rt.toDomain()
		if err != nil {
			return nil, err
		}
		transactions[i] = tx
	}

	return &block.SolanaBlock{
		Slot:              slot,
		Blockhash:         rb.Blockhash,
		PreviousBlockhash: rb.PreviousBlockhash,
		ParentSlot:        rb.ParentSlot,
		BlockHeight:       rb.BlockHeight,
		BlockTime:         blockTime,
		Transactions:      transactions,
	}, nil
}
