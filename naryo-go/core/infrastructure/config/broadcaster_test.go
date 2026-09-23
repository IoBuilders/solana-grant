//go:build test

package config

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/broadcaster/target"
)

const (
	validBroadcasterID              = "00000000-0000-0000-0000-000000000001"
	validBroadcasterConfigurationID = "00000000-0000-0000-0000-000000000002"
	validTargetFilterID             = "00000000-0000-0000-0000-000000000003"
)

// --- BroadcasterConfigurationProperties.Map ---

func TestBroadcasterConfigurationProperties_Map_InvalidID(t *testing.T) {
	bc := &BroadcasterConfigurationProperties{ID: "not-a-uuid", Type: broadcaster.TypeHTTP.String()}

	_, err := bc.Map()

	assert.Error(t, err)
}

func TestBroadcasterConfigurationProperties_Map_Success(t *testing.T) {
	bc := &BroadcasterConfigurationProperties{
		ID:                   validBroadcasterID,
		Type:                 broadcaster.TypeHTTP.String(),
		AdditionalProperties: map[string]any{"endpoint": map[string]any{"url": "http://example.com/webhook"}},
	}

	result, err := bc.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, broadcaster.TypeHTTP, result.Type())
	id, _ := uuid.Parse(validBroadcasterID)
	assert.Equal(t, id, result.ID())
	assert.Equal(t, bc.AdditionalProperties, result.AdditionalProperties())
}

func TestBroadcasterConfigurationProperties_AdditionalProperties_CapturesUnknownKeys(t *testing.T) {
	t.Setenv("CONFIG_PATH", "")
	yml := `
naryo:
  broadcasting:
    configuration:
      - id: "` + validBroadcasterID + `"
        type: "KAFKA"
        acks: "all"
        compressiontype: "gzip"
`

	cfg, err := LoadConfig[EnvironmentProperties](context.Background(), yml, "naryo")

	require.NoError(t, err)
	require.NotNil(t, cfg.Broadcasting)
	require.Len(t, cfg.Broadcasting.Configuration, 1)
	bc := cfg.Broadcasting.Configuration[0]
	assert.Equal(t, "all", bc.AdditionalProperties["acks"])
	assert.Equal(t, "gzip", bc.AdditionalProperties["compressiontype"])
	_, hasType := bc.AdditionalProperties["type"]
	assert.False(t, hasType, "a known field must not leak into AdditionalProperties")
}

// --- BroadcasterProperties.Map ---

func TestBroadcasterProperties_Map_InvalidID(t *testing.T) {
	b := &BroadcasterProperties{ID: "not-a-uuid", ConfigurationID: validBroadcasterConfigurationID}

	_, err := b.Map()

	assert.Error(t, err)
}

func TestBroadcasterProperties_Map_InvalidConfigurationID(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: "not-a-uuid",
		Target:          BroadcasterTarget{Type: target.TypeBlock.String(), Destinations: []string{"/hook"}},
	}

	_, err := b.Map()

	assert.Error(t, err)
}

func TestBroadcasterProperties_Map_UnsupportedTargetType(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target:          BroadcasterTarget{Type: "UNKNOWN", Destinations: []string{"/hook"}},
	}

	_, err := b.Map()

	assert.ErrorContains(t, err, "unsupported broadcaster target type: UNKNOWN")
}

func TestBroadcasterProperties_Map_BlockTarget_Success(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target:          BroadcasterTarget{Type: target.TypeBlock.String(), Destinations: []string{"/hook"}},
	}

	result, err := b.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, target.TypeBlock, result.Target.Type())
}

func TestBroadcasterProperties_Map_TransactionTarget_Success(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target:          BroadcasterTarget{Type: target.TypeTransaction.String(), Destinations: []string{"/hook"}},
	}

	result, err := b.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, target.TypeTransaction, result.Target.Type())
}

func TestBroadcasterProperties_Map_ContractEventTarget_Success(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target:          BroadcasterTarget{Type: target.TypeContractEvent.String(), Destinations: []string{"/hook"}},
	}

	result, err := b.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, target.TypeContractEvent, result.Target.Type())
}

func TestBroadcasterProperties_Map_AllTarget_Success(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target:          BroadcasterTarget{Type: target.TypeAll.String(), Destinations: []string{"/hook"}},
	}

	result, err := b.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, target.TypeAll, result.Target.Type())
}

func TestBroadcasterProperties_Map_FilterTarget_Success(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target: BroadcasterTarget{
			Type:         target.TypeFilter.String(),
			Destinations: []string{"/hook"},
			FilterID:     validTargetFilterID,
		},
	}

	result, err := b.Map()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, target.TypeFilter, result.Target.Type())
	filterTarget, ok := result.Target.(*target.FilterTarget)
	require.True(t, ok)
	expectedFilterID, _ := uuid.Parse(validTargetFilterID)
	assert.Equal(t, expectedFilterID, filterTarget.FilterID)
}

func TestBroadcasterProperties_Map_FilterTarget_InvalidFilterID(t *testing.T) {
	b := &BroadcasterProperties{
		ID:              validBroadcasterID,
		ConfigurationID: validBroadcasterConfigurationID,
		Target: BroadcasterTarget{
			Type:         target.TypeFilter.String(),
			Destinations: []string{"/hook"},
			FilterID:     "not-a-uuid",
		},
	}

	_, err := b.Map()

	assert.Error(t, err)
}
