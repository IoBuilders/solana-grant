//go:build test

package contractcallbuilder

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDltIngressContractCallBuilderRegistry_GetRegisteredBuilder(t *testing.T) {
	registry := NewRegistry()
	mock := &EvmContractCallBuilder{}
	registry.Register(common.EVM, "MyContract", mock)

	builder, err := registry.GetContractCallBuilder(common.EVM, "MyContract")

	require.NoError(t, err)
	assert.Equal(t, mock, builder)
}

func TestDltIngressContractCallBuilderRegistry_NotFound_UnknownDlt(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.GetContractCallBuilder(common.EVM, "Missing")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "contract call builder not found")
}

func TestDltIngressContractCallBuilderRegistry_NotFound_UnknownContract(t *testing.T) {
	registry := NewRegistry()
	registry.Register(common.EVM, "MyContract", &EvmContractCallBuilder{})

	_, err := registry.GetContractCallBuilder(common.EVM, "Other")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "contract call builder not found")
}

func TestDltIngressContractCallBuilderRegistry_RegisterMultipleDlts(t *testing.T) {
	registry := NewRegistry()
	evmMock := &EvmContractCallBuilder{}
	svmMock := &EvmContractCallBuilder{}
	registry.Register(common.EVM, "Token", evmMock)
	registry.Register(common.SVM, "Token", svmMock)

	evmBuilder, err := registry.GetContractCallBuilder(common.EVM, "Token")
	require.NoError(t, err)
	assert.Same(t, evmMock, evmBuilder)

	svmBuilder, err := registry.GetContractCallBuilder(common.SVM, "Token")
	require.NoError(t, err)
	assert.Same(t, svmMock, svmBuilder)
}

func TestDltIngressContractCallBuilderRegistry_RegisterOverwritesExisting(t *testing.T) {
	registry := NewRegistry()
	first := &EvmContractCallBuilder{}
	second := &EvmContractCallBuilder{}
	registry.Register(common.EVM, "Token", first)
	registry.Register(common.EVM, "Token", second)

	builder, err := registry.GetContractCallBuilder(common.EVM, "Token")
	require.NoError(t, err)
	assert.Same(t, second, builder)
}
