//go:build test

package contractcaller

import (
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDltIngressContractCallerRegistry_GetRegisteredCaller(t *testing.T) {
	registry := NewRegistry()
	mock := &EvmContractCaller{}
	registry.Register(common.EVM, mock)

	caller, err := registry.GetContractCaller(common.EVM)

	require.NoError(t, err)
	assert.Same(t, mock, caller)
}

func TestDltIngressContractCallerRegistry_NotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.GetContractCaller(common.EVM)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no contract caller configured for dlt")
}

func TestDltIngressContractCallerRegistry_RegisterMultipleDlts(t *testing.T) {
	registry := NewRegistry()
	evmMock := &EvmContractCaller{}
	svmMock := &EvmContractCaller{}
	registry.Register(common.EVM, evmMock)
	registry.Register(common.SVM, svmMock)

	evmCaller, err := registry.GetContractCaller(common.EVM)
	require.NoError(t, err)
	assert.Same(t, evmMock, evmCaller)

	svmCaller, err := registry.GetContractCaller(common.SVM)
	require.NoError(t, err)
	assert.Same(t, svmMock, svmCaller)
}

func TestDltIngressContractCallerRegistry_RegisterOverwritesExisting(t *testing.T) {
	registry := NewRegistry()
	first := &EvmContractCaller{}
	second := &EvmContractCaller{}
	registry.Register(common.EVM, first)
	registry.Register(common.EVM, second)

	caller, err := registry.GetContractCaller(common.EVM)

	require.NoError(t, err)
	assert.Same(t, second, caller)
	assert.NotSame(t, first, caller)
}
