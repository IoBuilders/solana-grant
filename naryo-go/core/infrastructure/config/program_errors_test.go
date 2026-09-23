//go:build test

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

func TestProgramErrors_Map_AggregatesMultiplePrograms(t *testing.T) {
	props := ProgramErrors{
		{
			ProgramID: "program-1",
			Codes: []ProgramErrorCodeProperties{
				{Name: "NotSettlementOperator", Code: 6002},
				{Name: "InsufficientBalance", Code: 6010},
			},
		},
		{
			ProgramID: "program-2",
			Codes: []ProgramErrorCodeProperties{
				{Name: "Unauthorized", Code: 6000},
			},
		},
	}

	result, err := props.Map()

	require.NoError(t, err)
	assert.Equal(t, event.ProgramErrorRegistry{
		"program-1": {6002: "NotSettlementOperator", 6010: "InsufficientBalance"},
		"program-2": {6000: "Unauthorized"},
	}, result)
}

func TestProgramErrors_Map_EmptyInput(t *testing.T) {
	var props ProgramErrors

	result, err := props.Map()

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestProgramErrors_Map_DuplicateProgramID_LaterEntryWins(t *testing.T) {
	props := ProgramErrors{
		{ProgramID: "program-1", Codes: []ProgramErrorCodeProperties{{Name: "First", Code: 1}}},
		{ProgramID: "program-1", Codes: []ProgramErrorCodeProperties{{Name: "Second", Code: 2}}},
	}

	result, err := props.Map()

	require.NoError(t, err)
	assert.Equal(t, event.ProgramErrorRegistry{
		"program-1": {2: "Second"},
	}, result)
}
