//go:build test

package svm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"dlt-ingress/src/main/dltingress/internal/infra/svm"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDltIngressSvmNodeChecker_Name(t *testing.T) {
	assert.Equal(t, "svm-node-test-network", svm.NewNodeChecker("test-network", &svm.SvmClientMock{}).Name())
}

func TestDltIngressSvmNodeChecker_Check(t *testing.T) {
	tests := []struct {
		name           string
		healthErr      error
		expectedStatus health.Status
	}{
		{"healthy node is up", nil, health.StatusUp},
		{"unhealthy node is down", errors.New("node is behind by 100 slots"), health.StatusDown},
		{"local rate limit means the node is reachable", ratelimit.NewExceededError("getHealth", time.Second), health.StatusUp},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &svm.SvmClientMock{}
			client.On("GetHealth", mock.Anything).Return(tt.healthErr).Once()

			result := svm.NewNodeChecker("test-network", client).Check(context.Background())

			assert.Equal(t, tt.expectedStatus, result.Status)
			client.AssertExpectations(t)
		})
	}
}
