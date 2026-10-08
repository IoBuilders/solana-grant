//go:build test

package svm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testBlockhash = "CSymwgTNX1j3E4qhKfJAUE41nBWEwXufoYryPbkde5ci"
	testSignedTx  = "AQABAgME"
	testAccount   = "11111111111111111111111111111111"
)

func defaultBucket(requestsPerMinute int) *ratelimit.RateLimiter {
	return ratelimit.New(ratelimit.WithBucket("DEFAULT", requestsPerMinute, ratelimit.MatchAll()))
}

func TestDltIngressSvmRateLimitedClient_DelegatesWhenAllowed(t *testing.T) {
	inner := &svm.SvmClientMock{}
	inner.On("GetRecentBlockhash", mock.Anything).Return(testBlockhash, nil).Once()
	client := svm.NewRateLimitedClient(inner, defaultBucket(1), nil)

	blockhash, err := client.GetRecentBlockhash(context.Background())

	require.NoError(t, err)
	assert.Equal(t, testBlockhash, blockhash)
	inner.AssertExpectations(t)
}

func TestDltIngressSvmRateLimitedClient_RejectsWhenBucketExhausted(t *testing.T) {
	inner := &svm.SvmClientMock{}
	inner.On("SendTransaction", mock.Anything, testSignedTx).Return("signature", nil).Once()
	client := svm.NewRateLimitedClient(inner, defaultBucket(1), nil)

	_, err := client.SendTransaction(context.Background(), testSignedTx)
	require.NoError(t, err)
	_, err = client.SendTransaction(context.Background(), testSignedTx)

	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "expected ExceededError, got %v", err)
	assert.Equal(t, "sendTransaction", exceededErr.Method)
	inner.AssertNumberOfCalls(t, "SendTransaction", 1)
}

func TestDltIngressSvmRateLimitedClient_UsesTheJsonRpcMethodNameAsBucketKey(t *testing.T) {
	inner := &svm.SvmClientMock{}
	inner.On("GetBalance", mock.Anything, testAccount).Return(nil, nil).Once()
	inner.On("GetRecentBlockhash", mock.Anything).Return(testBlockhash, nil).Once()
	limiter := ratelimit.New(ratelimit.WithBucket("BALANCE", 1, ratelimit.Methods("getBalance")))
	client := svm.NewRateLimitedClient(inner, limiter, nil)

	_, err := client.GetBalance(context.Background(), testAccount)
	require.NoError(t, err)
	_, err = client.GetBalance(context.Background(), testAccount)
	assert.True(t, ratelimit.IsExceededError(err))

	_, err = client.GetRecentBlockhash(context.Background())
	assert.NoError(t, err, "getLatestBlockhash is not in the BALANCE bucket")
	inner.AssertExpectations(t)
}

func TestDltIngressSvmRateLimitedClient_NodeRateLimitBlocksTheBucket(t *testing.T) {
	inner := &svm.SvmClientMock{}
	nodeErr := fmt.Errorf("error retrieving latest blockhash: %w", &noderatelimit.TooManyRequestsError{})
	inner.On("GetRecentBlockhash", mock.Anything).Return("", nodeErr).Once()
	client := svm.NewRateLimitedClient(inner, defaultBucket(100), nil)

	_, err := client.GetRecentBlockhash(context.Background())
	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "the node's 429 must surface as ExceededError, got %v", err)
	assert.Equal(t, "getLatestBlockhash", exceededErr.Method)

	_, err = client.GetRecentPrioritizationFees(context.Background(), []string{testAccount})
	assert.True(t, ratelimit.IsExceededError(err), "the node's 429 must block the bucket without reaching the node")
	inner.AssertExpectations(t)
}

func TestDltIngressSvmRateLimitedClient_ReturnsOtherErrorsUnchanged(t *testing.T) {
	inner := &svm.SvmClientMock{}
	nodeErr := errors.New("rpc error")
	inner.On("SimulateTransaction", mock.Anything, testSignedTx).Return(uint64(0), nodeErr).Once()
	client := svm.NewRateLimitedClient(inner, defaultBucket(100), nil)

	_, err := client.SimulateTransaction(context.Background(), testSignedTx)

	assert.Equal(t, nodeErr, err)
	inner.AssertExpectations(t)
}

func TestDltIngressSvmRateLimitedClient_GetHealthUsesItsOwnBucketKey(t *testing.T) {
	inner := &svm.SvmClientMock{}
	inner.On("GetHealth", mock.Anything).Return(nil).Once()
	inner.On("GetRecentBlockhash", mock.Anything).Return(testBlockhash, nil).Once()
	limiter := ratelimit.New(ratelimit.WithBucket("HEALTH", 1, ratelimit.Methods("getHealth")))
	client := svm.NewRateLimitedClient(inner, limiter, nil)

	require.NoError(t, client.GetHealth(context.Background()))
	err := client.GetHealth(context.Background())
	exceededErr, ok := errors.AsType[*ratelimit.ExceededError](err)
	require.True(t, ok, "expected ExceededError, got %v", err)
	assert.Equal(t, "getHealth", exceededErr.Method)

	_, err = client.GetRecentBlockhash(context.Background())
	assert.NoError(t, err, "getLatestBlockhash is not in the HEALTH bucket")
	inner.AssertNumberOfCalls(t, "GetHealth", 1)
}
