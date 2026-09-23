package solanadecoder

import (
	"fmt"
	"math/big"

	"github.com/mr-tron/base58"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// requireAccounts returns a typed error if fewer than needed accounts
// remain for instruction to decode.
func requireAccounts(accounts []string, needed int, instruction string) error {
	if len(accounts) < needed {
		return domainerrors.NewInvalidFieldError(
			"Accounts", "SplNativeDecoder",
			fmt.Sprintf("instruction %q requires at least %d accounts, got %d", instruction, needed, len(accounts)),
		)
	}
	return nil
}

func pubkeyParam(position int, address string) (parameter.ContractEventParameter, error) {
	return parameter.NewSolanaPublicKeyParameter(position, address)
}

func uintParam(position int, value uint64, bitSize int) (parameter.ContractEventParameter, error) {
	return parameter.NewSolanaUintParameter(position, new(big.Int).SetUint64(value), bitSize)
}

// pubkeyParams wraps each address as a positional PUBLIC_KEY parameter, in
// order, starting at position 0.
func pubkeyParams(addresses ...string) ([]parameter.ContractEventParameter, error) {
	params := make([]parameter.ContractEventParameter, 0, len(addresses))
	for i, addr := range addresses {
		p, err := pubkeyParam(i, addr)
		if err != nil {
			return nil, err
		}
		params = append(params, p)
	}
	return params, nil
}

// pubkeyArrayParam builds a PUBLIC_KEY array parameter from addresses, used
// both for the required signer list of InitializeMultisig(2) and for the
// optional trailing multisig-signer accounts of instructions with a
// single-owner-or-multisig authority.
func pubkeyArrayParam(position int, addresses []string) (parameter.ContractEventParameter, error) {
	elements := make([]parameter.ContractEventParameter, 0, len(addresses))
	for i, addr := range addresses {
		p, err := pubkeyParam(i, addr)
		if err != nil {
			return nil, err
		}
		elements = append(elements, p)
	}
	return parameter.NewSolanaArrayParameter(position, elements)
}

// appendTrailingSigners appends a PUBLIC_KEY array parameter (possibly
// empty) built from any accounts beyond an instruction's fixed positions —
// the optional multisig signer accounts a single-owner-or-multisig
// authority may be followed by.
func appendTrailingSigners(params []parameter.ContractEventParameter, trailing []string) ([]parameter.ContractEventParameter, error) {
	signers, err := pubkeyArrayParam(len(params), trailing)
	if err != nil {
		return nil, err
	}
	return append(params, signers), nil
}

// decodeSingleAccountNoData decodes any instruction whose layout is exactly
// one account and no data: InitializeImmutableOwner, SyncNative,
// InitializeNonTransferableMint.
func decodeSingleAccountNoData(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "single-account instruction"); err != nil {
		return nil, err
	}
	return pubkeyParams(accounts[0])
}

// decodeOptionPubkey decodes a 1-byte tag (0 = None, 1 = Some) followed by a
// 32-byte pubkey when Some, matching TokenInstruction's pack_pubkey_option.
func decodeOptionPubkey(data []byte, position int) (parameter.ContractEventParameter, int, error) {
	if err := requireBytes(data, position, 1); err != nil {
		return nil, 0, err
	}
	switch data[0] {
	case 0:
		p, err := parameter.NewSolanaOptionParameter(position, nil)
		return p, 1, err
	case 1:
		if err := requireBytes(data[1:], position, 32); err != nil {
			return nil, 0, err
		}
		inner, err := pubkeyParam(position, base58.Encode(data[1:33]))
		if err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaOptionParameter(position, inner)
		return p, 33, err
	default:
		return nil, 0, domainerrors.NewInvalidFieldError("Tag", "SplNativeDecoder", fmt.Sprintf("invalid option tag byte %d, must be 0 or 1", data[0]))
	}
}

// decodeOptionUint64 decodes a 1-byte tag (0 = None, 1 = Some) followed by an
// 8-byte LE u64 when Some, matching TokenInstruction's pack_u64_option.
func decodeOptionUint64(data []byte, position int) (parameter.ContractEventParameter, int, error) {
	if err := requireBytes(data, position, 1); err != nil {
		return nil, 0, err
	}
	switch data[0] {
	case 0:
		p, err := parameter.NewSolanaOptionParameter(position, nil)
		return p, 1, err
	case 1:
		if err := requireBytes(data[1:], position, 8); err != nil {
			return nil, 0, err
		}
		inner, err := uintParam(position, decodeUintLE(data[1:9]).Uint64(), 64)
		if err != nil {
			return nil, 0, err
		}
		p, err := parameter.NewSolanaOptionParameter(position, inner)
		return p, 9, err
	default:
		return nil, 0, domainerrors.NewInvalidFieldError("Tag", "SplNativeDecoder", fmt.Sprintf("invalid option tag byte %d, must be 0 or 1", data[0]))
	}
}

// decodeExtensionTypesArray decodes the rest of data as a sequence of 2-byte
// LE extension-type values — GetAccountDataSize and Reallocate encode this
// with no length prefix, unlike Borsh's u32-prefixed Vec.
func decodeExtensionTypesArray(data []byte, position int) (parameter.ContractEventParameter, error) {
	if len(data)%2 != 0 {
		return nil, domainerrors.NewInvalidFieldError("ExtensionTypes", "SplNativeDecoder", "trailing byte is not a complete u16 extension type")
	}
	elements := make([]parameter.ContractEventParameter, 0, len(data)/2)
	for i := 0; i+2 <= len(data); i += 2 {
		p, err := uintParam(i/2, decodeUintLE(data[i:i+2]).Uint64(), 16)
		if err != nil {
			return nil, err
		}
		elements = append(elements, p)
	}
	return parameter.NewSolanaArrayParameter(position, elements)
}
