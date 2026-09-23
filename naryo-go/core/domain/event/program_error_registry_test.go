//go:build test

package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgramErrorRegistry_Lookup_Hit(t *testing.T) {
	registry := ProgramErrorRegistry{
		"program-1": {6002: "NotSettlementOperator"},
	}

	name, ok := registry.Lookup("program-1", 6002)

	assert.True(t, ok)
	assert.Equal(t, "NotSettlementOperator", name)
}

func TestProgramErrorRegistry_Lookup_UnknownProgram(t *testing.T) {
	registry := ProgramErrorRegistry{
		"program-1": {6002: "NotSettlementOperator"},
	}

	name, ok := registry.Lookup("program-2", 6002)

	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestProgramErrorRegistry_Lookup_UnknownCode(t *testing.T) {
	registry := ProgramErrorRegistry{
		"program-1": {6002: "NotSettlementOperator"},
	}

	name, ok := registry.Lookup("program-1", 9999)

	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestProgramErrorRegistry_Lookup_NilRegistry(t *testing.T) {
	var registry ProgramErrorRegistry

	name, ok := registry.Lookup("program-1", 6002)

	assert.False(t, ok)
	assert.Empty(t, name)
}

func TestProgramErrorRegistry_Lookup_EmptyRegistry(t *testing.T) {
	registry := ProgramErrorRegistry{}

	name, ok := registry.Lookup("program-1", 6002)

	assert.False(t, ok)
	assert.Empty(t, name)
}
