//go:build test

package svm_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/svm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
