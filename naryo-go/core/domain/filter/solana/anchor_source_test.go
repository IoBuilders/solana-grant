//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnchorSource_IsValid(t *testing.T) {
	assert.True(t, AnchorSourceEmitLog.IsValid())
	assert.True(t, AnchorSourceEmitCPI.IsValid())
	assert.True(t, AnchorSourceInstruction.IsValid())
	assert.False(t, AnchorSource("LOGS").IsValid())
	assert.False(t, AnchorSource("").IsValid())
}
