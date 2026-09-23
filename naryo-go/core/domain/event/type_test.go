//go:build test

package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_IsValid(t *testing.T) {
	assert.True(t, TypeBlock.IsValid())
	assert.True(t, TypeTransaction.IsValid())
	assert.True(t, TypeContract.IsValid())
	assert.True(t, TypeSlot.IsValid())
	assert.False(t, Type("UNKNOWN").IsValid())
	assert.False(t, Type("").IsValid())
}
