package solanadecoder

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// decodeReallocate decodes Reallocate: accounts = [Account, Payer,
// SystemProgram, Owner(+ optional multisig signers)], data = extensionTypes
// []u16 (rest of buffer).
func decodeReallocate(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 4, "reallocate"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2], accounts[3])
	if err != nil {
		return nil, err
	}
	extensionTypes, err := decodeExtensionTypesArray(data, 4)
	if err != nil {
		return nil, err
	}
	params = append(params, extensionTypes)
	return appendTrailingSigners(params, accounts[4:])
}

// decodeWithdrawExcessLamports decodes WithdrawExcessLamports: accounts =
// [Source, Destination, Authority(+ optional multisig signers)], no data.
func decodeWithdrawExcessLamports(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "withdrawExcessLamports"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2])
	if err != nil {
		return nil, err
	}
	return appendTrailingSigners(params, accounts[3:])
}

// decodeUnwrapLamports decodes UnwrapLamports: accounts = [Source,
// Destination, Authority(+ optional multisig signers)], data = amount
// option<u64>.
func decodeUnwrapLamports(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 3, "unwrapLamports"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2])
	if err != nil {
		return nil, err
	}
	amount, _, err := decodeOptionUint64(data, 3)
	if err != nil {
		return nil, err
	}
	params = append(params, amount)
	return appendTrailingSigners(params, accounts[3:])
}
