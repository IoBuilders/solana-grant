package portcommon

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// EncodeUnsignedTransaction returns the unsigned RLP-encoded transaction as a 0x-prefixed hex string.
// The format matches what EIP-155 (legacy) and EIP-1559 (dynamic fee)
func EncodeUnsignedTransaction(tx *EVMTransactionResponse) (string, error) {
	data, err := hexutil.Decode(tx.Data)
	if err != nil {
		return "", fmt.Errorf("decode transaction data: %w", err)
	}

	if tx.Nonce == nil {
		return "", fmt.Errorf("nonce is required")
	}
	if tx.GasLimit == nil {
		return "", fmt.Errorf("gasLimit is required")
	}

	toAddr := common.HexToAddress(tx.To)
	nonce := tx.Nonce.RawValue().Uint64()
	gasLimit := tx.GasLimit.RawValue().Uint64()

	value := tx.ValueBigInt()

	var rawBytes []byte

	switch tx.TransactionType {
	case TransactionTypeLegacy:
		if tx.ChainId == nil {
			return "", fmt.Errorf("chainId is required for EIP-155 legacy transaction")
		}
		if tx.GasPrice == nil {
			return "", fmt.Errorf("gasPrice is required for legacy transaction")
		}
		// EIP-155 signing pre-image: RLP([nonce, gasPrice, gasLimit, to, value, data, chainId, 0, 0])
		rawBytes, err = rlp.EncodeToBytes([]interface{}{
			nonce,
			tx.GasPrice.RawValue(),
			gasLimit,
			&toAddr,
			value,
			data,
			tx.ChainId.RawValue(),
			uint(0),
			uint(0),
		})
		if err != nil {
			return "", fmt.Errorf("encode legacy transaction: %w", err)
		}

	case TransactionTypeDynamicFee:
		if tx.ChainId == nil {
			return "", fmt.Errorf("chainId is required for EIP-1559 dynamic fee transaction")
		}
		if tx.MaxPriorityFeePerGas == nil {
			return "", fmt.Errorf("maxPriorityFeePerGas is required for dynamic fee transaction")
		}
		if tx.MaxFeePerGas == nil {
			return "", fmt.Errorf("maxFeePerGas is required for dynamic fee transaction")
		}
		// EIP-1559 signing pre-image: 0x02 || RLP([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, accessList])
		var inner []byte
		inner, err = rlp.EncodeToBytes([]interface{}{
			tx.ChainId.RawValue(),
			nonce,
			tx.MaxPriorityFeePerGas.RawValue(),
			tx.MaxFeePerGas.RawValue(),
			gasLimit,
			&toAddr,
			value,
			data,
			types.AccessList{},
		})
		if err != nil {
			return "", fmt.Errorf("encode dynamic fee transaction: %w", err)
		}
		rawBytes = append([]byte{byte(TransactionTypeDynamicFee)}, inner...)

	default:
		return "", fmt.Errorf("unsupported transaction type: %d", tx.TransactionType)
	}

	return hexutil.Encode(rawBytes), nil
}
