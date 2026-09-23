//go:build test

package target

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_IsValid(t *testing.T) {
	assert.True(t, TypeBlock.IsValid())
	assert.True(t, TypeTransaction.IsValid())
	assert.True(t, TypeContractEvent.IsValid())
	assert.True(t, TypeFilter.IsValid())
	assert.True(t, TypeAll.IsValid())
	assert.False(t, Type("UNKNOWN").IsValid())
}
