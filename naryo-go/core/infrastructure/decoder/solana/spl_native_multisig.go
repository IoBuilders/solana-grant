package solanadecoder

import "gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event/parameter"

// decodeInitializeMultisig returns a decoder for InitializeMultisig(2):
// accounts = [Multisig, (Rent,) signers...], data = m u8. rentAccounts is 1
// for InitializeMultisig, which has a Rent account right after the
// Multisig account, and 0 for InitializeMultisig2, which has none.
func decodeInitializeMultisig(rentAccounts int) func([]byte, []string) ([]parameter.ContractEventParameter, error) {
	signersFrom := 1 + rentAccounts
	return func(data []byte, accounts []string) ([]parameter.ContractEventParameter, error) {
		if err := requireAccounts(accounts, signersFrom, "initializeMultisig"); err != nil {
			return nil, err
		}
		if err := requireBytes(data, 1, 1); err != nil {
			return nil, err
		}
		multisig, err := pubkeyParam(0, accounts[0])
		if err != nil {
			return nil, err
		}
		m, err := uintParam(1, uint64(data[0]), 8)
		if err != nil {
			return nil, err
		}
		signers, err := pubkeyArrayParam(2, accounts[signersFrom:])
		if err != nil {
			return nil, err
		}
		return []parameter.ContractEventParameter{multisig, m, signers}, nil
	}
}
