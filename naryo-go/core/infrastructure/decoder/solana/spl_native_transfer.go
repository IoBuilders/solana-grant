package solanadecoder

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// decodeAmountWithAuthority returns a decoder for the common
// [account..., authority(+ optional multisig signers)] shape with a single
// u64 amount field: Transfer, Approve, MintTo, Burn. authorityIndex is
// where the fixed authority account sits (any accounts beyond it are
// optional multisig signers).
func decodeAmountWithAuthority(authorityIndex int) func([]byte, []string) ([]parameter.ContractEventParameter, error) {
	return func(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
		if err := requireAccounts(accounts, authorityIndex+1, "amount+authority instruction"); err != nil {
			return nil, err
		}
		if err := requireBytes(data, authorityIndex+1, 8); err != nil {
			return nil, err
		}
		params, err := pubkeyParams(accounts[:authorityIndex+1]...)
		if err != nil {
			return nil, err
		}
		amount, err := uintParam(authorityIndex+1, decodeUintLE(data[:8]).Uint64(), 64)
		if err != nil {
			return nil, err
		}
		params = append(params, amount)
		return appendTrailingSigners(params, accounts[authorityIndex+1:])
	}
}

// decodeAmountCheckedWithAuthority is decodeAmountWithAuthority plus a
// trailing decimals u8: MintToChecked, BurnChecked.
func decodeAmountCheckedWithAuthority(authorityIndex int) func([]byte, []string) ([]parameter.ContractEventParameter, error) {
	return func(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
		if err := requireAccounts(accounts, authorityIndex+1, "amount+decimals+authority instruction"); err != nil {
			return nil, err
		}
		if err := requireBytes(data, authorityIndex+1, 9); err != nil {
			return nil, err
		}
		params, err := pubkeyParams(accounts[:authorityIndex+1]...)
		if err != nil {
			return nil, err
		}
		amount, err := uintParam(authorityIndex+1, decodeUintLE(data[:8]).Uint64(), 64)
		if err != nil {
			return nil, err
		}
		decimals, err := uintParam(authorityIndex+2, uint64(data[8]), 8)
		if err != nil {
			return nil, err
		}
		params = append(params, amount, decimals)
		return appendTrailingSigners(params, accounts[authorityIndex+1:])
	}
}

// decodeCheckedTransferApprove decodes TransferChecked/ApproveChecked:
// accounts = [Source, Mint, Destination-or-Delegate, Authority(+ optional
// multisig signers)], data = amount u64, decimals u8.
func decodeCheckedTransferApprove(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 4, "transferChecked/approveChecked"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 4, 9); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1], accounts[2], accounts[3])
	if err != nil {
		return nil, err
	}
	amount, err := uintParam(4, decodeUintLE(data[:8]).Uint64(), 64)
	if err != nil {
		return nil, err
	}
	decimals, err := uintParam(5, uint64(data[8]), 8)
	if err != nil {
		return nil, err
	}
	params = append(params, amount, decimals)
	return appendTrailingSigners(params, accounts[4:])
}

// decodeRevoke decodes Revoke: accounts = [Source, Authority(+ optional
// multisig signers)], no data.
func decodeRevoke(_ []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 2, "revoke"); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1])
	if err != nil {
		return nil, err
	}
	return appendTrailingSigners(params, accounts[2:])
}

// decodeSetAuthority decodes SetAuthority: accounts = [Target,
// CurrentAuthority(+ optional multisig signers)], data = authorityType u8,
// newAuthority option<pubkey>.
func decodeSetAuthority(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
	if err := requireAccounts(accounts, 2, "setAuthority"); err != nil {
		return nil, err
	}
	if err := requireBytes(data, 2, 1); err != nil {
		return nil, err
	}
	params, err := pubkeyParams(accounts[0], accounts[1])
	if err != nil {
		return nil, err
	}
	authorityType, err := uintParam(2, uint64(data[0]), 8)
	if err != nil {
		return nil, err
	}
	newAuthority, _, err := decodeOptionPubkey(data[1:], 3)
	if err != nil {
		return nil, err
	}
	params = append(params, authorityType, newAuthority)
	return appendTrailingSigners(params, accounts[2:])
}
