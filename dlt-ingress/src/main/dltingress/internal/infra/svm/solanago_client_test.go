//go:build test

package svm_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/svm"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A well-formed 64-byte signature; the content is irrelevant because the node is a stub.
const testSignature = "5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW"

type recordedRpcRequest struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

func newRecordingSolanaNode(t *testing.T, result string) (*httptest.Server, *atomic.Pointer[recordedRpcRequest]) {
	t.Helper()
	recorded := &atomic.Pointer[recordedRpcRequest]{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var req recordedRpcRequest
		require.NoError(t, json.Unmarshal(body, &req))
		recorded.Store(&req)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","result":` + result + `,"id":1}`))
	}))
	t.Cleanup(server.Close)
	return server, recorded
}

func TestDltIngressSvmSolanaGoClient_GetRecentPrioritizationFees_SendsAccounts(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `[{"slot":1,"prioritizationFee":1000},{"slot":2,"prioritizationFee":2000}]`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })
	accounts := []string{"9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin", "So11111111111111111111111111111111111111112"}

	fees, err := client.GetRecentPrioritizationFees(context.Background(), accounts)

	require.NoError(t, err)
	assert.Equal(t, []uint64{1000, 2000}, fees)
	req := recorded.Load()
	require.NotNil(t, req)
	assert.Equal(t, "getRecentPrioritizationFees", req.Method)
	require.Len(t, req.Params, 1)
	var sentAccounts []string
	require.NoError(t, json.Unmarshal(req.Params[0], &sentAccounts))
	assert.Equal(t, accounts, sentAccounts)
}

func TestDltIngressSvmSolanaGoClient_GetRecentPrioritizationFees_InvalidAccount_DoesNotReachNode(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `[]`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.GetRecentPrioritizationFees(context.Background(), []string{"not-a-public-key"})

	assert.ErrorContains(t, err, "not-a-public-key")
	assert.Nil(t, recorded.Load(), "an unparseable account must fail before the node is called")
}

func serializedTestTransaction(t *testing.T) string {
	t.Helper()
	payer := solana.MustPublicKeyFromBase58("9xQeWvG816bUx9EPjHmaT23yvVM2ZWbrrpZb9PusVFin")
	tx, err := solana.NewTransaction(
		[]solana.Instruction{solana.NewInstruction(solana.SystemProgramID, solana.AccountMetaSlice{solana.Meta(payer).WRITE().SIGNER()}, []byte{})},
		solana.MustHashFromBase58("4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM"),
		solana.TransactionPayer(payer),
	)
	require.NoError(t, err)
	data, err := tx.MarshalBinary()
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(data)
}

func lastParamAsMap(t *testing.T, req *recordedRpcRequest) map[string]any {
	t.Helper()
	require.NotNil(t, req)
	require.NotEmpty(t, req.Params)
	var opts map[string]any
	require.NoError(t, json.Unmarshal(req.Params[len(req.Params)-1], &opts))
	return opts
}

func TestDltIngressSvmSolanaGoClient_SimulateTransaction_UsesConfirmedCommitment(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `{"context":{"slot":1},"value":{"err":null,"logs":[],"unitsConsumed":1234}}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	units, err := client.SimulateTransaction(context.Background(), serializedTestTransaction(t))

	require.NoError(t, err)
	assert.Equal(t, uint64(1234), units)
	assert.Equal(t, "simulateTransaction", recorded.Load().Method)
	assert.Equal(t, "confirmed", lastParamAsMap(t, recorded.Load())["commitment"],
		"the blockhash is fetched at confirmed, so simulation must run at confirmed too")
}

func TestDltIngressSvmSolanaGoClient_SendTransaction_UsesConfirmedPreflightCommitment(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `"5VERv8NMvzbJMEkV8xnrLkEaWRtSz9CosKDYjCJjBRnbJLgp8uirBgmQpjKhoR4tjF3ZpRzrFmBV6UjKdiSZkQUW"`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.SendTransaction(context.Background(), serializedTestTransaction(t))

	require.NoError(t, err)
	assert.Equal(t, "sendTransaction", recorded.Load().Method)
	opts := lastParamAsMap(t, recorded.Load())
	assert.Equal(t, "confirmed", opts["preflightCommitment"],
		"the blockhash is fetched at confirmed, so preflight must run at confirmed too")
	assert.NotEqual(t, true, opts["skipPreflight"])
}

func TestDltIngressSvmSolanaGoClient_GetRecentBlockhash_UsesConfirmedCommitment(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `{"context":{"slot":1},"value":{"blockhash":"4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM","lastValidBlockHeight":100}}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.GetRecentBlockhash(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "confirmed", lastParamAsMap(t, recorded.Load())["commitment"])
}

func TestDltIngressSvmSolanaGoClient_GetSignatureStatus_SearchesHistory(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `{"context":{"slot":9},"value":[{"slot":8,"confirmations":null,"err":null,"confirmationStatus":"finalized"}]}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	status, err := client.GetSignatureStatus(context.Background(), testSignature)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.False(t, status.Failed)
	assert.Equal(t, "finalized", status.ConfirmationStatus)
	assert.Equal(t, "getSignatureStatuses", recorded.Load().Method)
	require.Len(t, recorded.Load().Params, 2)
	assert.JSONEq(t, `{"searchTransactionHistory":true}`, string(recorded.Load().Params[1]),
		"without a history search a landed transaction older than the recent status cache looks unknown")
}

func TestDltIngressSvmSolanaGoClient_GetSignatureStatus_FailedTransactionIsReportedAsFailed(t *testing.T) {
	node, _ := newRecordingSolanaNode(t, `{"context":{"slot":9},"value":[{"slot":8,"confirmations":1,"err":{"InstructionError":[0,{"Custom":1}]},"confirmationStatus":"confirmed"}]}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	status, err := client.GetSignatureStatus(context.Background(), testSignature)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.True(t, status.Failed)
}

func TestDltIngressSvmSolanaGoClient_GetSignatureStatus_UnknownSignature_ReturnsNil(t *testing.T) {
	node, _ := newRecordingSolanaNode(t, `{"context":{"slot":9},"value":[null]}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	status, err := client.GetSignatureStatus(context.Background(), testSignature)

	require.NoError(t, err)
	assert.Nil(t, status)
}

func TestDltIngressSvmSolanaGoClient_GetSignatureStatus_InvalidSignature_DoesNotReachNode(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `{"context":{"slot":9},"value":[null]}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.GetSignatureStatus(context.Background(), "not-a-signature")

	assert.Error(t, err)
	assert.Nil(t, recorded.Load())
}

func TestDltIngressSvmSolanaGoClient_IsBlockhashValid(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprintf("valid=%t", valid), func(t *testing.T) {
			node, recorded := newRecordingSolanaNode(t, fmt.Sprintf(`{"context":{"slot":9},"value":%t}`, valid))
			client := svm.NewSolanaGoClient(node.URL)
			t.Cleanup(func() { _ = client.Close() })

			got, err := client.IsBlockhashValid(context.Background(), "4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM")

			require.NoError(t, err)
			assert.Equal(t, valid, got)
			assert.Equal(t, "isBlockhashValid", recorded.Load().Method)
			assert.Equal(t, "confirmed", lastParamAsMap(t, recorded.Load())["commitment"],
				"the blockhash was fetched at confirmed, so validity must be judged at confirmed")
		})
	}
}

func TestDltIngressSvmSolanaGoClient_IsBlockhashValid_InvalidBlockhash_DoesNotReachNode(t *testing.T) {
	node, recorded := newRecordingSolanaNode(t, `{"context":{"slot":9},"value":true}`)
	client := svm.NewSolanaGoClient(node.URL)
	t.Cleanup(func() { _ = client.Close() })

	_, err := client.IsBlockhashValid(context.Background(), "not-a-blockhash")

	assert.Error(t, err)
	assert.Nil(t, recorded.Load())
}
