package contracttransactionbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/infra/portcommon"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
)

const (
	nativeTransferEvmTo     = "0x000000000000000000000000000000000000dEaD"
	nativeTransferBlockhash = "11111111111111111111111111111111"
)

var nativeTransferSender = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
var nativeTransferRecipient = solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")

func TestDltIngressNativeTransfer_ImplementsPortInterface(t *testing.T) {
	var _ Port = (*NativeTransferContractTransactionBuilder)(nil)
}

func TestDltIngressNativeTransfer_InvalidArgsLength(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	cases := []struct {
		name string
		args map[string]any
	}{
		{"no args", map[string]any{}},
		{"one arg", map[string]any{"to": nativeTransferEvmTo}},
		{"three args", map[string]any{"to": nativeTransferEvmTo, "amount": "1", "extra": "x"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := NewEVMLegacyBuildTransactionRequest(
				"sender", smartContractId, "", c.args, chainId, nonce, gasLimit, gasPrice, value,
			)
			resp, err := builder.BuildTransaction(req)
			assert.Nil(t, resp)
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), "invalid native transfer args length")
		})
	}
}

func TestDltIngressNativeTransfer_InvalidArgs(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	cases := []struct {
		name string
		args map[string]any
	}{
		{"missing to", map[string]any{"amount": "1", "other": "x"}},
		{"missing amount", map[string]any{"to": nativeTransferEvmTo, "other": "x"}},
		{"to wrong type", map[string]any{"to": 123, "amount": "1"}},
		{"amount wrong type", map[string]any{"to": nativeTransferEvmTo, "amount": 1}},
		{"empty to", map[string]any{"to": "", "amount": "1"}},
		{"empty amount", map[string]any{"to": nativeTransferEvmTo, "amount": ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := NewEVMLegacyBuildTransactionRequest(
				"sender", smartContractId, "", c.args, chainId, nonce, gasLimit, gasPrice, value,
			)
			resp, err := builder.BuildTransaction(req)
			assert.Nil(t, resp)
			require.NotNil(t, err)
			assert.Contains(t, err.Error(), "invalid native transfer args")
		})
	}
}

func TestDltIngressNativeTransfer_InvalidAmount(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewEVMLegacyBuildTransactionRequest(
		"sender", smartContractId, "",
		map[string]any{"to": nativeTransferEvmTo, "amount": "not-a-number"},
		chainId, nonce, gasLimit, gasPrice, value,
	)
	resp, err := builder.BuildTransaction(req)
	assert.Nil(t, resp)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid native transfer amount: not-a-number")
}

func TestDltIngressNativeTransfer_EVM_LegacyOk(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085", smartContractId, "",
		map[string]any{"to": nativeTransferEvmTo, "amount": "1000"},
		chainId, nonce, gasLimit, gasPrice, value,
	)
	resp, err := builder.BuildTransaction(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.EVMTransactionResponse)
	assert.Nil(t, resp.SVMTransactionResponse)

	expectedAmount, _ := amount.NewFromString("1000")
	assert.Equal(t, portcommon.TransactionTypeLegacy, resp.EVMTransactionResponse.TransactionType)
	assert.Equal(t, chainId, resp.EVMTransactionResponse.ChainId)
	assert.Equal(t, nonce, resp.EVMTransactionResponse.Nonce)
	assert.Equal(t, gasLimit, resp.EVMTransactionResponse.GasLimit)
	assert.Equal(t, gasPrice, resp.EVMTransactionResponse.GasPrice)
	assert.Equal(t, "0x", resp.EVMTransactionResponse.Data)
	assert.Equal(t, expectedAmount, resp.EVMTransactionResponse.Value)
	assert.Equal(t, nativeTransferEvmTo, resp.EVMTransactionResponse.To)
	assert.Nil(t, resp.EVMTransactionResponse.MaxPriorityFeePerGas)
	assert.Nil(t, resp.EVMTransactionResponse.MaxFeePerGas)
}

func TestDltIngressNativeTransfer_EVM_DynamicFeeOk(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewEVMDynamicFeeBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085", smartContractId, "",
		map[string]any{"to": nativeTransferEvmTo, "amount": "1000"},
		chainId, nonce, gasLimit, maxPriorityFeePerGas, maxFeePerGas, value,
	)
	resp, err := builder.BuildTransaction(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.EVMTransactionResponse)

	expectedAmount, _ := amount.NewFromString("1000")
	assert.Equal(t, portcommon.TransactionTypeDynamicFee, resp.EVMTransactionResponse.TransactionType)
	assert.Nil(t, resp.EVMTransactionResponse.GasPrice)
	assert.Equal(t, maxPriorityFeePerGas, resp.EVMTransactionResponse.MaxPriorityFeePerGas)
	assert.Equal(t, maxFeePerGas, resp.EVMTransactionResponse.MaxFeePerGas)
	assert.Equal(t, "0x", resp.EVMTransactionResponse.Data)
	assert.Equal(t, expectedAmount, resp.EVMTransactionResponse.Value)
	assert.Equal(t, nativeTransferEvmTo, resp.EVMTransactionResponse.To)
}

func TestDltIngressNativeTransfer_EVM_OverrideGasLimit_SetsNewLimit(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085", smartContractId, "",
		map[string]any{"to": nativeTransferEvmTo, "amount": "1000"},
		chainId, nonce, gasLimit, gasPrice, value,
	)
	original, _ := builder.BuildTransaction(req)
	originalLimit := original.EVMTransactionResponse.GasLimit

	newLimit, _ := amount.NewFromString("999999")
	updated, err := builder.OverrideGasLimit(*req, *original, newLimit)

	require.NoError(t, err)
	assert.Equal(t, newLimit, updated.EVMTransactionResponse.GasLimit)
	assert.Equal(t, originalLimit, original.EVMTransactionResponse.GasLimit, "original response must not be mutated")
}

func TestDltIngressNativeTransfer_SVM_Ok(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		nativeTransferBlockhash, nil, nil,
	)
	resp, err := builder.BuildTransaction(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.SVMTransactionResponse)
	assert.Nil(t, resp.EVMTransactionResponse)

	svm := resp.SVMTransactionResponse
	assert.Equal(t, nativeTransferSender.String(), svm.FeePayer)
	assert.Equal(t, nativeTransferBlockhash, svm.RecentBlockhash)
	assert.Equal(t, uint8(1), svm.Header.NumRequiredSignatures)
	assert.Contains(t, svm.AccountKeys, nativeTransferSender.String())
	assert.Contains(t, svm.AccountKeys, nativeTransferRecipient.String())
	require.Len(t, svm.Instructions, 1)

	tx, err := solana.TransactionFromBase64(svm.SerializedTransaction)
	require.NoError(t, err)
	assert.Equal(t, nativeTransferSender, tx.Message.AccountKeys[0])
	require.Len(t, tx.Message.Instructions, 1)
}

func TestDltIngressNativeTransfer_SVM_InvalidTo(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": "not-a-pubkey", "amount": "42"},
		nativeTransferBlockhash, nil, nil,
	)
	resp, err := builder.BuildTransaction(req)
	assert.Nil(t, resp)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid native transfer to: not-a-pubkey")
}

func TestDltIngressNativeTransfer_SVM_MissingSVMRequest(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := &BuildTransactionRequest{
		SenderDltAccountId: nativeTransferSender.String(),
		MethodArgs:         map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
	}
	resp, err := builder.BuildTransaction(req)
	assert.Nil(t, resp)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "SVMBuildTransactionRequest is required")
}

func TestDltIngressNativeTransfer_SVM_InvalidSender(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		"not-a-pubkey", "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		nativeTransferBlockhash, nil, nil,
	)
	resp, err := builder.BuildTransaction(req)
	assert.Nil(t, resp)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid sender")
}

func TestDltIngressNativeTransfer_SVM_InvalidBlockhash(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		"not-a-hash!!!", nil, nil,
	)
	resp, err := builder.BuildTransaction(req)
	assert.Nil(t, resp)
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid recent blockhash")
}

func TestDltIngressNativeTransfer_SVM_PrependsTwoComputeBudgetInstructions_WhenCuLimitAndPriceSet(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	cuLimit, _ := amount.NewFromString("200000")
	cuPrice, _ := amount.NewFromString("5000")

	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		nativeTransferBlockhash, cuPrice, cuLimit,
	)
	resp, err := builder.BuildTransaction(req)
	require.NoError(t, err)

	tx, err := solana.TransactionFromBase64(resp.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)

	// Two compute budget instructions (SetComputeUnitLimit + SetComputeUnitPrice) + 1 native transfer instruction.
	require.Len(t, tx.Message.Instructions, 3)

	computeBudgetProgram := solana.ComputeBudget

	ix0 := tx.Message.Instructions[0]
	assert.Equal(t, computeBudgetProgram, tx.Message.AccountKeys[ix0.ProgramIDIndex])
	assert.Equal(t, computebudget.Instruction_SetComputeUnitLimit, []byte(ix0.Data)[0])

	ix1 := tx.Message.Instructions[1]
	assert.Equal(t, computeBudgetProgram, tx.Message.AccountKeys[ix1.ProgramIDIndex])
	assert.Equal(t, computebudget.Instruction_SetComputeUnitPrice, []byte(ix1.Data)[0])

	assert.Equal(t, cuLimit, resp.SVMTransactionResponse.CuLimit)
	assert.Equal(t, cuPrice, resp.SVMTransactionResponse.CuPrice)
}

func TestDltIngressNativeTransfer_SVM_OverrideGasLimit_SetsNewLimit(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		nativeTransferBlockhash, nil, nil,
	)
	original, _ := builder.BuildTransaction(req)
	originalCuLimit := req.SVMBuildTransactionRequest.CuLimit

	cuLimit, _ := amount.NewFromString("200000")
	updated, err := builder.OverrideGasLimit(*req, *original, cuLimit)

	require.NoError(t, err)
	require.NotNil(t, updated.SVMTransactionResponse)
	assert.Equal(t, cuLimit, updated.SVMTransactionResponse.CuLimit)
	assert.Equal(t, originalCuLimit, req.SVMBuildTransactionRequest.CuLimit, "original request must not be mutated")
}

func TestDltIngressNativeTransfer_SVM_OverrideCuPrice_SetsNewPrice(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewSVMBuildTransactionRequest(
		nativeTransferSender.String(), "", "",
		map[string]any{"to": nativeTransferRecipient.String(), "amount": "42"},
		nativeTransferBlockhash, nil, nil,
	)
	original, _ := builder.BuildTransaction(req)

	cuPrice, _ := amount.NewFromString("5000")
	updated, err := builder.OverrideCuPrice(*req, *original, cuPrice)

	require.NoError(t, err)
	require.NotNil(t, updated.SVMTransactionResponse)
	assert.Equal(t, cuPrice, updated.SVMTransactionResponse.CuPrice)
	assert.Nil(t, req.SVMBuildTransactionRequest.CuPrice, "original request must not be mutated")

	tx, err := solana.TransactionFromBase64(updated.SVMTransactionResponse.SerializedTransaction)
	require.NoError(t, err)
	// SetComputeUnitPrice + transfer.
	require.Len(t, tx.Message.Instructions, 2)
	assert.Equal(t, computebudget.Instruction_SetComputeUnitPrice, []byte(tx.Message.Instructions[0].Data)[0])
}

func TestDltIngressNativeTransfer_EVM_OverrideCuPrice_NotSupported(t *testing.T) {
	builder := &NativeTransferContractTransactionBuilder{}
	req := NewEVMLegacyBuildTransactionRequest(
		"0x92BDb699940f5a5Fb21C7D6EFA5a67C84B84A085", smartContractId, "",
		map[string]any{"to": nativeTransferEvmTo, "amount": "1000"},
		chainId, nonce, gasLimit, gasPrice, value,
	)
	original, _ := builder.BuildTransaction(req)

	cuPrice, _ := amount.NewFromString("5000")
	updated, err := builder.OverrideCuPrice(*req, *original, cuPrice)

	assert.Nil(t, updated)
	assert.ErrorContains(t, err, "not supported on EVM")
}
