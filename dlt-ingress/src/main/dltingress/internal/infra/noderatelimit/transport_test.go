//go:build test

package noderatelimit_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rateLimitedBody = `{"error":{"code":-32605,"message":"IP Rate limit exceeded on eth_call"},"jsonrpc":"2.0","id":1}`

func newNode(t *testing.T, status int, headers map[string]string, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func post(t *testing.T, url string) (*http.Response, error) {
	client := &http.Client{Transport: noderatelimit.NewTransport(http.DefaultTransport)}
	resp, err := client.Post(url, "application/json", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"eth_call"}`))
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}
	return resp, err
}

func TestDltIngressNodeRateLimitTransport_429_ReturnsTooManyRequestsErrorWithHeadersAndBody(t *testing.T) {
	server := newNode(t, http.StatusTooManyRequests, map[string]string{"Retry-After": "30"}, rateLimitedBody)

	resp, err := post(t, server.URL)

	assert.Nil(t, resp)
	var tooManyRequestsErr *noderatelimit.TooManyRequestsError
	require.True(t, errors.As(err, &tooManyRequestsErr), "expected TooManyRequestsError, got %T: %v", err, err)
	assert.Equal(t, "30", tooManyRequestsErr.Header.Get("Retry-After"))
	assert.Equal(t, rateLimitedBody, string(tooManyRequestsErr.Body))
}

func TestDltIngressNodeRateLimitTransport_Non429_PassesResponseThrough(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := newNode(t, status, map[string]string{"Retry-After": "30"}, `{"jsonrpc":"2.0","id":1,"result":"0x"}`)

			resp, err := post(t, server.URL)

			require.NoError(t, err)
			assert.Equal(t, status, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, `{"jsonrpc":"2.0","id":1,"result":"0x"}`, string(body))
		})
	}
}
