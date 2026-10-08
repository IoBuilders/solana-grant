package custody

import (
	"encoding/hex"
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// TransformBase64SvmTransactionIntoHex converts a base64-encoded string into a hex-encoded string
func TransformBase64SvmTransactionIntoHex(base64Tx string) (string, error) {
	tx, err := solana.TransactionFromBase64(base64Tx)
	if err != nil {
		return "", err
	}

	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(txBytes), nil
}

// InjectSvmSignatures injects signatures into the transaction at the position
// each signer's public key occupies in tx.Message.AccountKeys. It returns the
// base64-encoded signed transaction along with its transaction id, which is
// the base58-encoded first signature (tx.Signatures[0]) and can be computed
// without submitting the transaction to the network.
func InjectSvmSignatures(serializedTx string, signatures map[string][]byte) (string, string, error) {
	tx, err := solana.TransactionFromBase64(serializedTx)
	if err != nil {
		return "", "", fmt.Errorf("deserialize transaction: %w", err)
	}

	for pubkey, sig := range signatures {
		pk, err := solana.PublicKeyFromBase58(pubkey)
		if err != nil {
			return "", "", fmt.Errorf("invalid signer pubkey %q: %w", pubkey, err)
		}
		idx := -1
		for i, ak := range tx.Message.AccountKeys {
			if ak.Equals(pk) {
				idx = i
				break
			}
		}
		if idx < 0 || idx >= int(tx.Message.Header.NumRequiredSignatures) {
			return "", "", fmt.Errorf("public key %q is not a required signer in this transaction", pubkey)
		}
		copy(tx.Signatures[idx][:], sig)
	}

	encodedTx, err := tx.ToBase64()
	if err != nil {
		return "", "", err
	}
	return encodedTx, tx.Signatures[0].String(), nil
}
