package evmtransaction

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

var (
	validTxId       = "0xabc123def456abc123def456abc123def456abc123def456abc123def456abc123"
	validNetworkId  = "1"
	validNetworkUrl = "http://localhost:8545"
	validDlt        = string(common.EVM)
	validFrom       = "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12"
	validTo         = "0x1234567890AbCdEf1234567890AbCdEf12345678"
	validData       = "0x1234567890AbCdEf1234567890AbCdEf12345678"
)

func validAmounts() (nonce, value, gasLimit, gasPrice *amount.Amount) {
	nonce, _ = amount.NewFromString("1")
	value, _ = amount.NewFromString("0")
	gasLimit, _ = amount.NewFromString("21000")
	gasPrice, _ = amount.NewFromString("1000000000")
	return
}

func TestDltIngressNewEthereumTransaction_Success(t *testing.T) {
	tests := []struct {
		name            string
		transactionType uint
	}{
		{name: "type 0 (legacy)", transactionType: 0},
		{name: "type 2 (EIP-1559)", transactionType: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nonce, value, gasLimit, gasPrice := validAmounts()

			tx, err := NewEvmTransaction(
				validTxId, validNetworkId, validNetworkUrl, validDlt,
				validFrom, validTo,
				nonce, value, tt.transactionType,
				gasLimit, gasPrice, nil, nil, validData,
			)

			assert.Nil(t, err)
			assert.NotNil(t, tx)
			assert.Equal(t, validTxId, tx.TxId)
			assert.Equal(t, validNetworkId, tx.NetworkId)
			assert.Equal(t, validNetworkUrl, tx.NetworkUrl)
			assert.Equal(t, common.EVM, tx.Dlt)
			assert.Equal(t, validFrom, tx.FromAddress)
			assert.Equal(t, validTo, tx.ToAddress)
			assert.Equal(t, tt.transactionType, tx.TransactionType)
		})
	}
}

func TestDltIngressNewEthereumTransaction_InvalidTransactionType(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, validDlt,
		validFrom, validTo,
		nonce, value, 1,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidTransactionType)
}

func TestDltIngressNewEthereumTransaction_InvalidDlt(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, "INVALID_DLT",
		validFrom, validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidDlt)
}

func TestDltIngressNewEthereumTransaction_EmptyTxId(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		"", validNetworkId, validNetworkUrl, validDlt,
		validFrom, validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "TxId")
}

func TestDltIngressNewEthereumTransaction_TxIdTooLong(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		strings.Repeat("a", 256), validNetworkId, validNetworkUrl, validDlt,
		validFrom, validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "TxId")
}

func TestDltIngressNewEthereumTransaction_EmptyNetworkId(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, "", validNetworkUrl, validDlt,
		validFrom, validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkId")
}

func TestDltIngressNewEthereumTransaction_EmptyNetworkUrl(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, "", validDlt,
		validFrom, validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkUrl")
}

func TestDltIngressNewEthereumTransaction_EmptyFromAddress(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, validDlt,
		"", validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "FromAddress")
}

func TestDltIngressNewEthereumTransaction_EmptyToAddress(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, validDlt,
		validFrom, "",
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "ToAddress")
}

func TestDltIngressNewEthereumTransaction_FromAddressTooLong(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, validDlt,
		strings.Repeat("a", 256), validTo,
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "FromAddress")
}

func TestDltIngressNewEthereumTransaction_ToAddressTooLong(t *testing.T) {
	nonce, value, gasLimit, gasPrice := validAmounts()

	tx, err := NewEvmTransaction(
		validTxId, validNetworkId, validNetworkUrl, validDlt,
		validFrom, strings.Repeat("a", 256),
		nonce, value, 0,
		gasLimit, gasPrice, nil, nil, validData,
	)

	assert.Nil(t, tx)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "ToAddress")
}
