package transaction

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"dlt-ingress/src/main/dltingress/domain/transaction/svmtransaction"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var (
	svmValidTxId            = "5VfYmGBL5dV1aH6dvB1Yd9k3oQ8Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2nQ2Yk9eQ7Wd2"
	svmValidNetworkId       = "solana-devnet"
	svmValidNetworkUrl      = "https://api.devnet.solana.com"
	svmValidFeePayer        = "9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin"
	svmValidRecentBlockhash = "GHtXQBsoZHVnNFa9YevAzFr17DJjgHXk3ycTKD5xD3Zi"
	svmValidSerializedTx    = "AQABAgIDBAUGBwgJ"
	svmValidCuLimit, _      = amount.NewFromString("1")
	svmValidCuPrice, _      = amount.NewFromString("2")
)

func TestDltIngressNewSVMTransaction_Success(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, svmValidNetworkUrl,
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, svmValidTxId, tx.TxId)
	assert.Equal(t, svmValidNetworkId, tx.NetworkId)
	assert.Equal(t, svmValidNetworkUrl, tx.NetworkUrl)
	assert.Equal(t, common.SVM, tx.Dlt)
	assert.Equal(t, svmValidFeePayer, tx.FeePayer)
	assert.Equal(t, svmValidRecentBlockhash, tx.RecentBlockhash)
	assert.Equal(t, svmValidSerializedTx, tx.SerializedTransaction)
	assert.Equal(t, svmValidCuLimit, tx.CuLimit)
	assert.Equal(t, svmValidCuPrice, tx.CuPrice)
}

func TestDltIngressNewSVMTransaction_EmptyTxId(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		"", svmValidNetworkId, svmValidNetworkUrl,
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "TxId")
}

func TestDltIngressNewSVMTransaction_TxIdTooLong(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		strings.Repeat("a", 256), svmValidNetworkId, svmValidNetworkUrl,
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "TxId")
}

func TestDltIngressNewSVMTransaction_EmptyNetworkId(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, "", svmValidNetworkUrl,
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkId")
}

func TestDltIngressNewSVMTransaction_EmptyNetworkUrl(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, "",
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkUrl")
}

func TestDltIngressNewSVMTransaction_EmptyFeePayer(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, svmValidNetworkUrl,
		"", svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "FeePayer")
}

func TestDltIngressNewSVMTransaction_FeePayerTooLong(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, svmValidNetworkUrl,
		strings.Repeat("a", 256), svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "FeePayer")
}

func TestDltIngressNewSVMTransaction_RecentBlockhashTooLong(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, svmValidNetworkUrl,
		svmValidFeePayer, strings.Repeat("a", 101), svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "RecentBlockhash")
}

func TestDltIngressSVMTransaction_Clone(t *testing.T) {
	tx, err := svmtransaction.NewSvmTransaction(
		svmValidTxId, svmValidNetworkId, svmValidNetworkUrl,
		svmValidFeePayer, svmValidRecentBlockhash, svmValidSerializedTx,
		svmValidCuLimit, svmValidCuPrice,
	)
	assert.Nil(t, err)

	clone := tx.Clone()

	assert.NotNil(t, clone)
	assert.NotSame(t, tx, clone)
	assert.Equal(t, tx.TxId, clone.TxId)
	assert.Equal(t, tx.Dlt, clone.Dlt)
	assert.Equal(t, tx.FeePayer, clone.FeePayer)
	assert.Equal(t, tx.RecentBlockhash, clone.RecentBlockhash)
	assert.Equal(t, tx.SerializedTransaction, clone.SerializedTransaction)
	assert.Equal(t, tx.CuLimit, clone.CuLimit)
	assert.Equal(t, tx.CuPrice, clone.CuPrice)
}

func TestDltIngressSVMTransaction_CloneNil(t *testing.T) {
	var tx *svmtransaction.SvmTransaction
	assert.Nil(t, tx.Clone())
}
