//go:build test || integration

package integration

import (
	"context"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/domain/transaction/evmtransaction"
	"dlt-ingress/src/main/dltingress/internal/domain/transaction/svmtransaction"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

// A transaction whose send fails is deleted and then rebuilt by the caller's retry. On SVM the rebuilt
// transaction can carry the very same tx_id (Ed25519 signatures are deterministic), so the delete must free it.
func TestDltIngressSvmTransactionHardDelete(t *testing.T) {
	repo := DltIngress.Repositories.SvmTransactionRepo
	ctx := context.Background()

	t.Run("Hard delete frees the tx_id for a rebuilt transaction", func(t *testing.T) {
		txId := "svm-" + uuid.NewString()
		require.NoError(t, repo.Save(ctx, newSvmTransaction(t, txId)))

		require.NoError(t, repo.HardDelete(ctx, mustFindSvmTransaction(t, repo, txId)))

		require.NoError(t, repo.Save(ctx, newSvmTransaction(t, txId)))
		mustFindSvmTransaction(t, repo, txId)
	})

	t.Run("Soft delete keeps the tx_id taken", func(t *testing.T) {
		txId := "svm-" + uuid.NewString()
		require.NoError(t, repo.Save(ctx, newSvmTransaction(t, txId)))

		require.NoError(t, repo.Delete(ctx, mustFindSvmTransaction(t, repo, txId)))

		require.ErrorContains(t, repo.Save(ctx, newSvmTransaction(t, txId)), "svm_transactions_tx_id_key")
	})
}

func TestDltIngressEvmTransactionHardDelete(t *testing.T) {
	repo := DltIngress.Repositories.EvmTransactionRepo
	ctx := context.Background()

	t.Run("Hard delete frees the tx_id for a rebuilt transaction", func(t *testing.T) {
		txId := "0x" + uuid.NewString()
		require.NoError(t, repo.Save(ctx, newEvmTransaction(t, txId)))

		require.NoError(t, repo.HardDelete(ctx, mustFindEvmTransaction(t, repo, txId)))

		require.NoError(t, repo.Save(ctx, newEvmTransaction(t, txId)))
		mustFindEvmTransaction(t, repo, txId)
	})
}

func newSvmTransaction(t *testing.T, txId string) *svmtransaction.SvmTransaction {
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")
	tx, err := svmtransaction.NewSvmTransaction(
		txId,
		"solana-localnet",
		"http://localhost:8899",
		"9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin",
		"GHtXQBsoZHVnNFa9YevAzFr17DJjgHXk3ycTKD5xD3Zi",
		"AQABAgIDBAUGBwgJ",
		cuLimit,
		cuPrice,
	)
	require.NoError(t, err)
	return tx
}

func newEvmTransaction(t *testing.T, txId string) *evmtransaction.EvmTransaction {
	nonce, _ := amount.NewFromString("1")
	value, _ := amount.NewFromString("0")
	gasLimit, _ := amount.NewFromString("21000")
	gasPrice, _ := amount.NewFromString("1000000000")
	tx, err := evmtransaction.NewEvmTransaction(
		txId,
		"default",
		"http://localhost:8545",
		"EVM",
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		"0x1234567890AbCdEf1234567890AbCdEf12345678",
		nonce,
		value,
		0,
		gasLimit,
		gasPrice,
		nil,
		nil,
		"0x",
	)
	require.NoError(t, err)
	return tx
}

func mustFindSvmTransaction(t *testing.T, repo svmtransaction.Repository, txId string) *svmtransaction.SvmTransaction {
	tx, err := repo.FindByTxId(context.Background(), txId)
	require.NoError(t, err)
	return tx
}

func mustFindEvmTransaction(t *testing.T, repo evmtransaction.Repository, txId string) *evmtransaction.EvmTransaction {
	tx, err := repo.FindByTxId(context.Background(), txId)
	require.NoError(t, err)
	return tx
}
