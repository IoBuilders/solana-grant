//go:build test

package parameter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_IsValid(t *testing.T) {
	assert.True(t, TypeBool.IsValid())
	assert.True(t, TypeInt.IsValid())
	assert.True(t, TypeUint.IsValid())
	assert.True(t, TypeFloat.IsValid())
	assert.True(t, TypeString.IsValid())
	assert.True(t, TypeBytes.IsValid())
	assert.True(t, TypeBytesFixed.IsValid())
	assert.True(t, TypePublicKey.IsValid())
	assert.True(t, TypeArray.IsValid())
	assert.True(t, TypeStruct.IsValid())
	assert.True(t, TypeOption.IsValid())
	assert.False(t, Type("UNKNOWN").IsValid())
	assert.False(t, Type("").IsValid())
}
