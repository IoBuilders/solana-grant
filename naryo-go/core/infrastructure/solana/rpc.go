package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// rpcRequest is a JSON-RPC 2.0 request envelope. id is fixed since each
// call is a single HTTP request/response pair — there is no persistent
// connection multiplexing several in-flight calls that would need
// distinguishing by id.
type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response envelope. Result is left raw so the
// caller can unmarshal it into the type it expects, including detecting a
// JSON null result (e.g. getBlock for a skipped slot).
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

// rpcError is a JSON-RPC 2.0 error object.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// call issues method over HTTP with params, decoding a successful result
// into result (a pointer). A JSON-RPC error result is mapped to a typed
// error by mapRPCError instead of being unmarshalled.
func (i *SolanaRpcBlockInteractor) call(ctx context.Context, method string, params []any, result any) error {
	reqBody, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: 1, Method: method, Params: params})
	if err != nil {
		return fmt.Errorf("solana rpc: marshal %s request: %w", method, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, i.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("solana rpc: build %s request: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := i.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("solana rpc: %s request failed: %w", method, err)
	}
	defer func() { _ = httpResp.Body.Close() }()

	var resp rpcResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return fmt.Errorf("solana rpc: decode %s response: %w", method, err)
	}

	if resp.Error != nil {
		return mapRPCError(method, *resp.Error)
	}
	if result == nil || len(resp.Result) == 0 {
		return nil
	}
	if err := json.Unmarshal(resp.Result, result); err != nil {
		return fmt.Errorf("solana rpc: unmarshal %s result: %w", method, err)
	}
	return nil
}
