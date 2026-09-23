//go:build test

package configurationmanager

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/descriptor"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/app/sourceprovider"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/event"
)

// --- stubs ---

type stubProgramErrorRegistrySourceProvider struct {
	item descriptor.ProgramErrorRegistry
	err  error
}

func (s *stubProgramErrorRegistrySourceProvider) Load(_ context.Context) (descriptor.ProgramErrorRegistry, error) {
	return s.item, s.err
}

func (s *stubProgramErrorRegistrySourceProvider) Priority() int { return 1 }

type stubProgramErrorRegistryDescriptor struct {
	result event.ProgramErrorRegistry
	err    error
}

func (s *stubProgramErrorRegistryDescriptor) Map() (event.ProgramErrorRegistry, error) {
	return s.result, s.err
}

// --- tests ---

func TestDefaultProgramErrorRegistryConfigurationManager_Load_Success(t *testing.T) {
	expected := event.ProgramErrorRegistry{"program-1": {6002: "NotSettlementOperator"}}
	provider := &stubProgramErrorRegistrySourceProvider{
		item: &stubProgramErrorRegistryDescriptor{result: expected},
	}
	cm := NewDefaultProgramErrorRegistryConfigurationManager([]sourceprovider.ProgramErrorRegistrySourceProvider{provider})

	result, err := cm.Load(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestDefaultProgramErrorRegistryConfigurationManager_Load_ProviderError(t *testing.T) {
	providerErr := errors.New("source unavailable")
	provider := &stubProgramErrorRegistrySourceProvider{err: providerErr}
	cm := NewDefaultProgramErrorRegistryConfigurationManager([]sourceprovider.ProgramErrorRegistrySourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, providerErr)
}

func TestDefaultProgramErrorRegistryConfigurationManager_Load_DescriptorMapError(t *testing.T) {
	mapErr := errors.New("invalid descriptor")
	provider := &stubProgramErrorRegistrySourceProvider{
		item: &stubProgramErrorRegistryDescriptor{err: mapErr},
	}
	cm := NewDefaultProgramErrorRegistryConfigurationManager([]sourceprovider.ProgramErrorRegistrySourceProvider{provider})

	_, err := cm.Load(context.Background())

	assert.ErrorIs(t, err, mapErr)
}
