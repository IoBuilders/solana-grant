//go:build test

package solana

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/interactor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/block"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/httpclient"
)

func newTestInteractor(t *testing.T, serverURL string) *SolanaRpcBlockInteractor {
	t.Helper()
	endpoint, err := common.NewConnectionEndpointFromURL(serverURL)
	require.NoError(t, err)
	connection, err := common.NewHttpConnection(endpoint, common.DefaultRetryConfiguration())
	require.NoError(t, err)
	client := httpclient.NewHttpClient(10, time.Second, time.Second, time.Second, time.Second, 5*time.Second, time.Second, false)
	rpcInteractor, err := NewSolanaRpcBlockInteractor(connection, client)
	require.NoError(t, err)
	return rpcInteractor
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func jsonRPCResult(result string) []byte {
	return []byte(`{"jsonrpc":"2.0","id":1,"result":` + result + `}`)
}

func jsonRPCErrorBody(code int, message string) []byte {
	body, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"error":   map[string]any{"code": code, "message": message},
	})
	return body
}

func TestNewSolanaRpcBlockInteractor(t *testing.T) {
	endpoint, err := common.NewConnectionEndpointFromURL("https://api.mainnet-beta.solana.com")
	require.NoError(t, err)
	httpConnection, err := common.NewHttpConnection(endpoint, common.DefaultRetryConfiguration())
	require.NoError(t, err)
	client := httpclient.NewHttpClient(10, time.Second, time.Second, time.Second, time.Second, 5*time.Second, time.Second, false)

	t.Run("Valid", func(t *testing.T) {
		i, err := NewSolanaRpcBlockInteractor(httpConnection, client)
		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, interactor.TypeBlock, i.Type())
	})

	t.Run("NilClient", func(t *testing.T) {
		_, err := NewSolanaRpcBlockInteractor(httpConnection, nil)
		assert.Error(t, err)
	})

	t.Run("WsConnection", func(t *testing.T) {
		wsEndpoint, err := common.NewConnectionEndpointFromURL("wss://api.mainnet-beta.solana.com")
		require.NoError(t, err)
		wsConnection, err := common.NewWsConnection(wsEndpoint, common.DefaultRetryConfiguration())
		require.NoError(t, err)
		_, err = NewSolanaRpcBlockInteractor(wsConnection, client)
		assert.Error(t, err)
	})
}

func TestSolanaRpcBlockInteractor_GetSlot(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var capturedBody map[string]any
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&capturedBody))
			_, _ = w.Write(jsonRPCResult("123456789"))
		})

		i := newTestInteractor(t, server.URL)
		slot, err := i.GetSlot(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(123456789), slot)

		assert.Equal(t, "getSlot", capturedBody["method"])
		params, ok := capturedBody["params"].([]any)
		require.True(t, ok)
		require.Len(t, params, 1)
		config, ok := params[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "confirmed", config["commitment"])
	})

	t.Run("RPCError", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(jsonRPCErrorBody(-32005, "Node is unhealthy"))
		})

		i := newTestInteractor(t, server.URL)
		_, err := i.GetSlot(context.Background())
		assert.Error(t, err)
	})

	t.Run("ContextCanceled", func(t *testing.T) {
		unblock := make(chan struct{})
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			<-unblock
		})
		t.Cleanup(func() { close(unblock) })

		i := newTestInteractor(t, server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := i.GetSlot(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestSolanaRpcBlockInteractor_GetBlock(t *testing.T) {
	t.Run("SkippedSlotReturnsNilNil", func(t *testing.T) {
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(jsonRPCResult("null"))
		})

		i := newTestInteractor(t, server.URL)
		b, err := i.GetBlock(context.Background(), 42)
		require.NoError(t, err)
		assert.Nil(t, b)
	})

	t.Run("Success", func(t *testing.T) {
		const blockJSON = `{
			"blockHeight": 100,
			"blockTime": 1700000000,
			"blockhash": "hash",
			"parentSlot": 41,
			"previousBlockhash": "prevhash",
			"transactions": [
				{
					"meta": {
						"err": null,
						"logMessages": ["log1", "log2"]
					},
					"transaction": {
						"signatures": ["sig1", "sig2"],
						"message": {
							"accountKeys": [
								{"pubkey": "acct1", "signer": true, "writable": true},
								{"pubkey": "acct2", "signer": false, "writable": false}
							],
							"instructions": [
								{
									"programId": "11111111111111111111111111111111",
									"parsed": {"type": "transfer", "info": {
										"source": "4wBqpZM9xaSheZzJSMawUKKwhdpChKbZ5eu5ky4Vigw",
										"destination": "JEJUoGfGEPTZ1XTwN39dYdFxYxDiDaSKVNy5qYWJmZt3",
										"lamports": 1000
									}}
								},
								{
									"programId": "unrecognizedProgram",
									"accounts": ["acct1", "acct2"],
									"data": "3DUrX2Ba"
								}
							]
						}
					}
				},
				{
					"meta": {
						"err": {"InstructionError": [0, "Custom"]},
						"logMessages": []
					},
					"transaction": {
						"signatures": ["sig3"],
						"message": {
							"accountKeys": [],
							"instructions": []
						}
					}
				}
			]
		}`

		var capturedBody map[string]any
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, json.NewDecoder(r.Body).Decode(&capturedBody))
			_, _ = w.Write(jsonRPCResult(blockJSON))
		})

		i := newTestInteractor(t, server.URL)
		b, err := i.GetBlock(context.Background(), 42)
		require.NoError(t, err)
		require.NotNil(t, b)

		assert.Equal(t, "getBlock", capturedBody["method"])
		params, ok := capturedBody["params"].([]any)
		require.True(t, ok)
		require.Len(t, params, 2)
		assert.InEpsilon(t, float64(42), params[0], 0)
		config, ok := params[1].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "jsonParsed", config["encoding"])
		assert.Equal(t, false, config["rewards"])
		assert.InDelta(t, float64(0), config["maxSupportedTransactionVersion"], 0)

		assert.Equal(t, uint64(42), b.Slot)
		assert.Equal(t, "hash", b.Blockhash)
		assert.Equal(t, "prevhash", b.PreviousBlockhash)
		assert.Equal(t, uint64(41), b.ParentSlot)
		require.NotNil(t, b.BlockHeight)
		assert.Equal(t, uint64(100), *b.BlockHeight)
		require.NotNil(t, b.BlockTime)
		assert.Equal(t, int64(1700000000), b.BlockTime.Unix())
		require.Len(t, b.Transactions, 2)

		tx1 := b.Transactions[0]
		assert.Equal(t, "sig1", tx1.Signature)
		assert.Equal(t, []string{"acct1", "acct2"}, tx1.Accounts)
		assert.Equal(t, []string{"log1", "log2"}, tx1.Logs)
		assert.Nil(t, tx1.Err)
		require.Len(t, tx1.Instructions, 2)

		parsedInstr := tx1.Instructions[0]
		assert.Equal(t, "11111111111111111111111111111111", parsedInstr.ProgramID)
		assert.Empty(t, parsedInstr.Data)
		assert.Equal(t, []string{
			"4wBqpZM9xaSheZzJSMawUKKwhdpChKbZ5eu5ky4Vigw",
			"JEJUoGfGEPTZ1XTwN39dYdFxYxDiDaSKVNy5qYWJmZt3",
		}, parsedInstr.Accounts)

		partialInstr := tx1.Instructions[1]
		assert.Equal(t, "unrecognizedProgram", partialInstr.ProgramID)
		assert.Equal(t, []string{"acct1", "acct2"}, partialInstr.Accounts)
		assert.NotEmpty(t, partialInstr.Data)

		tx2 := b.Transactions[1]
		assert.Equal(t, "sig3", tx2.Signature)
		require.NotNil(t, tx2.Err)
		assert.JSONEq(t, `{"InstructionError": [0, "Custom"]}`, *tx2.Err)
	})

	t.Run("ErrorMapping", func(t *testing.T) {
		tests := []struct {
			name       string
			code       int
			wantTarget error
		}{
			{"SlotSkipped", -32007, block.ErrSlotSkipped},
			{"LongTermStorageSlotSkipped", -32009, block.ErrSlotSkipped},
			{"BlockNotAvailable", -32004, block.ErrBlockNotAvailable},
			{"BlockStatusNotAvailableYet", -32014, block.ErrBlockNotAvailable},
			{"BlockCleanedUp", -32001, block.ErrBlockCleanedUp},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write(jsonRPCErrorBody(tt.code, "some message"))
				})

				i := newTestInteractor(t, server.URL)
				_, err := i.GetBlock(context.Background(), 42)
				assert.ErrorIs(t, err, tt.wantTarget)
			})
		}

		t.Run("UnknownCode", func(t *testing.T) {
			server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(jsonRPCErrorBody(-32099, "something else"))
			})

			i := newTestInteractor(t, server.URL)
			_, err := i.GetBlock(context.Background(), 42)
			assert.Error(t, err)
			assert.NotErrorIs(t, err, block.ErrSlotSkipped)
			assert.NotErrorIs(t, err, block.ErrBlockNotAvailable)
			assert.NotErrorIs(t, err, block.ErrBlockCleanedUp)
		})
	})

	t.Run("ContextCanceled", func(t *testing.T) {
		unblock := make(chan struct{})
		server := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
			<-unblock
		})
		t.Cleanup(func() { close(unblock) })

		i := newTestInteractor(t, server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := i.GetBlock(ctx, 42)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
