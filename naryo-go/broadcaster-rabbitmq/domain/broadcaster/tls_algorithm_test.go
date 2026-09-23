//go:build test

package broadcaster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLSAlgorithm_IsValid(t *testing.T) {
	assert.True(t, TLSAlgorithmTLS12.IsValid())
	assert.True(t, TLSAlgorithmTLS13.IsValid())
	assert.False(t, TLSAlgorithm("TLSv1.1").IsValid())
	assert.False(t, TLSAlgorithm("SSLv3").IsValid())
	assert.False(t, TLSAlgorithm("").IsValid())
}

func TestTLSAlgorithm_String(t *testing.T) {
	assert.Equal(t, "TLSv1.2", TLSAlgorithmTLS12.String())
	assert.Equal(t, "TLSv1.3", TLSAlgorithmTLS13.String())
}
