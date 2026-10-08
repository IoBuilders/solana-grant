package custody

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
)

func buildEVMUnsignedTx(tx *portcommon.TransactionResponse) (*types.Transaction, types.Signer, common.Hash, error) {
	data, err := hexutil.Decode(tx.Data)
	if err != nil {
		return nil, nil, common.Hash{}, fmt.Errorf("decode transaction data: %w", err)
	}

	toAddr := common.HexToAddress(tx.To)

	var unsignedTx *types.Transaction
	switch tx.TransactionType {
	case portcommon.TransactionTypeLegacy:
		unsignedTx = types.NewTx(&types.LegacyTx{
			Nonce:    tx.Nonce.RawValue().Uint64(),
			To:       &toAddr,
			Value:    tx.Value.RawValue(),
			Gas:      tx.GasLimit.RawValue().Uint64(),
			GasPrice: tx.GasPrice.RawValue(),
			Data:     data,
		})
	case portcommon.TransactionTypeDynamicFee:
		unsignedTx = types.NewTx(&types.DynamicFeeTx{
			Nonce:     tx.Nonce.RawValue().Uint64(),
			To:        &toAddr,
			Value:     tx.Value.RawValue(),
			Gas:       tx.GasLimit.RawValue().Uint64(),
			GasTipCap: tx.MaxPriorityFeePerGas.RawValue(),
			GasFeeCap: tx.MaxFeePerGas.RawValue(),
			Data:      data,
		})
	default:
		return nil, nil, common.Hash{}, fmt.Errorf("unsupported tx type: %d", tx.TransactionType)
	}

	signer := types.NewEIP155Signer(tx.ChainId.RawValue())
	return unsignedTx, signer, signer.Hash(unsignedTx), nil
}

func encodeSignedEVMTx(tx *types.Transaction, signer types.Signer, sigBytes []byte) (string, string, error) {
	signedTx, err := tx.WithSignature(signer, sigBytes)
	if err != nil {
		return "", "", fmt.Errorf("build signed transaction: %w", err)
	}
	rawTxBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		return "", "", fmt.Errorf("encode RLP: %w", err)
	}
	return hexutil.Encode(rawTxBytes), signedTx.Hash().Hex(), nil
}

// assembleECDSASig right-aligns r and s into a 65-byte Evm ECDSA signature [r(32) | s(32) | v(1)].
func assembleECDSASig(rBytes, sBytes []byte, recid int) []byte {
	sig := make([]byte, 65)
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	sig[64] = byte(recid)
	return sig
}
