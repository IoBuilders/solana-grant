package solanadecoder

import (
	"github.com/mr-tron/base58"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// decodeInitializeMint decodes InitializeMint/InitializeMint2: accounts =
// [Mint, (Rent)], data = decimals u8, mintAuthority pubkey, freezeAuthority
// option<pubkey>.
func decodeInitializeMint(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "initializeMint"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 1, 1+32); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	decimals, err := uintParam(1, uint64(data[0]), 8)
	if err != nil {
		return nil, err
	}
	mintAuthority, err := pubkeyParam(2, base58.Encode(data[1:33]))
	if err != nil {
		return nil, err
	}
	freezeAuthority, _, err := decodeOptionPubkey(data[33:], 3)
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, decimals, mintAuthority, freezeAuthority}, nil
}

// decodeInitializeMintCloseAuthority decodes InitializeMintCloseAuthority:
// accounts = [Mint], data = closeAuthority option<pubkey>.
func decodeInitializeMintCloseAuthority(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "initializeMintCloseAuthority"); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	closeAuthority, _, err := decodeOptionPubkey(data, 1)
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, closeAuthority}, nil
}

// decodeInitializePermanentDelegate decodes InitializePermanentDelegate:
// accounts = [Mint], data = delegate pubkey.
func decodeInitializePermanentDelegate(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "initializePermanentDelegate"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 1, 32); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	delegate, err := pubkeyParam(1, base58.Encode(data[:32]))
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, delegate}, nil
}

// decodeCreateNativeMint decodes CreateNativeMint: accounts = [Payer,
// NativeMint, SystemProgram], no data.
func decodeCreateNativeMint(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "createNativeMint"); err != nil {
		return nil, err
	}
	return pubkeyParams(accounts[0], accounts[1], accounts[2])
}

// decodeGetAccountDataSize decodes GetAccountDataSize: accounts = [Mint],
// data = extensionTypes []u16 (rest of buffer).
func decodeGetAccountDataSize(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "getAccountDataSize"); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	extensionTypes, err := decodeExtensionTypesArray(data, 1)
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, extensionTypes}, nil
}

// decodeAmountToUiAmount decodes AmountToUiAmount: accounts = [Mint], data =
// amount u64.
func decodeAmountToUiAmount(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "amountToUiAmount"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 1, 8); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	amount, err := uintParam(1, decodeUintLE(data[:8]).Uint64(), 64)
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, amount}, nil
}

// decodeUiAmountToAmount decodes UiAmountToAmount: accounts = [Mint], data =
// uiAmount string (rest of buffer, no length prefix).
func decodeUiAmountToAmount(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 1, "uiAmountToAmount"); err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	uiAmount, err := parameter.NewSolanaStringParameter(1, string(data))
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{mint, uiAmount}, nil
}
