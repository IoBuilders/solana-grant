//go:build test

package node

import (
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_IsValid(t *testing.T) {
	assert.True(t, TypeSolana.IsValid())
	assert.False(t, Type("ETHEREUM").IsValid())
	assert.False(t, Type("").IsValid())
}

func TestConnectionType_IsValid(t *testing.T) {
	assert.True(t, common.ConnectionTypeWs.IsValid())
	assert.True(t, common.ConnectionTypeHttp.IsValid())
	assert.False(t, common.ConnectionType("GRPC").IsValid())
}

func TestBlockSubscriptionMethod_IsValid(t *testing.T) {
	assert.True(t, BlockSubscriptionMethodPubSub.IsValid())
	assert.True(t, BlockSubscriptionMethodPoll.IsValid())
	assert.False(t, BlockSubscriptionMethod("WEBHOOK").IsValid())
}
