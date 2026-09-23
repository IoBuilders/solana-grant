package solanadecoder

import (
	"github.com/mr-tron/base58"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"
)

// decodeInitializeAccount decodes InitializeAccount: accounts = [Account,
// Mint, Owner, (Rent)], no data.
func decodeInitializeAccount(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "initializeAccount"); err != nil {
		return nil, err
	}
	return pubkeyParams(accounts[0], accounts[1], accounts[2])
}

// decodeInitializeAccount23 decodes InitializeAccount2/3: accounts =
// [Account, Mint, (Rent)], data = owner pubkey.
func decodeInitializeAccount23(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 2, "initializeAccount2/3"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 2, 32); err != nil {
		return nil, err
	}
	account, err := pubkeyParam(0, accounts[0])
	if err != nil {
		return nil, err
	}
	mint, err := pubkeyParam(1, accounts[1])
	if err != nil {
		return nil, err
	}
	owner, err := pubkeyParam(2, base58.Encode(data[:32]))
	if err != nil {
		return nil, err
	}
	return []parameter.ContractEventParameter{account, mint, owner}, nil
}

// decodeCloseAccount decodes CloseAccount: accounts = [Account, Destination,
// Owner(+ optional multisig signers)], no data.
func decodeCloseAccount(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "closeAccount"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2])
	if err != nil {
		return nil, err
	}
	return appendTrailingSigners(params, accounts[3:])
}

// decodeFreezeThawAccount decodes FreezeAccount/ThawAccount: accounts =
// [Account, Mint, FreezeAuthority(+ optional multisig signers)], no data.
func decodeFreezeThawAccount(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "freezeAccount/thawAccount"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2])
	if err != nil {
		return nil, err
	}
	return appendTrailingSigners(params, accounts[3:])
}
