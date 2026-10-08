//go:build test

package custody

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFakeKMS answers every KMS call with the given AWS JSON error.
func newFakeKMS(t *testing.T, errorCode string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	requests := &atomic.Int32{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		w.Header().Set("X-Amzn-ErrorType", errorCode)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"__type":"` + errorCode + `","message":"fake KMS error"}`))
	}))
	t.Cleanup(server.Close)
	return server, requests
}

func newTestKMSAdapter(url string, requestsPerMinute int) *KmsAdapter {
	return &KmsAdapter{
		client: kms.New(kms.Options{
			Region:           "us-east-1",
			BaseEndpoint:     aws.String(url),
			Credentials:      credentials.NewStaticCredentialsProvider("test", "test", ""),
			RetryMaxAttempts: 1,
		}),
		guard:       defaultBucketGuard(requestsPerMinute, ""),
		callTimeout: DefaultRequestTimeout,
	}
}

func TestDltIngressKmsAdapter_RejectsWithoutCallingKmsWhenBucketExhausted(t *testing.T) {
	server, requests := newFakeKMS(t, "ValidationException")
	adapter := newTestKMSAdapter(server.URL, 1)
	request := &CreateKeyRequest{KeyType: string(common.ED25519), Dlt: string(common.SVM)}

	_, err := adapter.CreateKey(context.Background(), request)
	require.Error(t, err)
	assert.False(t, ratelimit.IsExceededError(err), "a non throttling KMS error must pass through, got %v", err)
	_, err = adapter.CreateKey(context.Background(), request)

	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "expected ExceededError, got %v", err)
	assert.Equal(t, kmsMethodCreateKey, exceededErr.Method)
	assert.Equal(t, int32(1), requests.Load(), "the rejected call must not reach KMS")
}

func TestDltIngressKmsAdapter_ThrottlingBlocksTheBucket(t *testing.T) {
	server, requests := newFakeKMS(t, kmsThrottlingErrorCode)
	adapter := newTestKMSAdapter(server.URL, 100)
	request := &CreateKeyRequest{KeyType: string(common.ED25519), Dlt: string(common.SVM)}

	_, err := adapter.CreateKey(context.Background(), request)
	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "KMS throttling must surface as ExceededError, got %v", err)
	assert.Equal(t, kmsMethodCreateKey, exceededErr.Method)

	_, err = adapter.CreateKey(context.Background(), request)
	assert.True(t, ratelimit.IsExceededError(err))
	assert.Equal(t, int32(1), requests.Load(), "the throttled bucket must keep the next call from reaching KMS")
}

func TestDltIngressKmsAdapter_CallTimesOutWhenKmsHangs(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { <-release }))
	// Cleanups run last in, first out: release the hung handler before Close waits for it.
	t.Cleanup(server.Close)
	t.Cleanup(func() { close(release) })

	adapter := newTestKMSAdapter(server.URL, 100)
	adapter.callTimeout = 200 * time.Millisecond

	start := time.Now()
	_, err := adapter.CreateKey(context.Background(), &CreateKeyRequest{KeyType: string(common.ED25519), Dlt: string(common.SVM)})

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second, "the call must give up at the timeout instead of waiting on KMS")
}
