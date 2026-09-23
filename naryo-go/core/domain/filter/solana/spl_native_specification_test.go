//go:build test

package solana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
)

func TestNewSplNativeSpecification(t *testing.T) {
	mint := "So11111111111111111111111111111111111111112"

	t.Run("ValidWithMint", func(t *testing.T) {
		spec, err := NewSplNativeSpecification(SplInstructionTransferChecked, &mint)
		assert.NoError(t, err)
		assert.Equal(t, StrategySplNative, spec.Strategy())
		assert.Equal(t, SplInstructionTransferChecked, spec.Instruction)
		require.NotNil(t, spec.MintAddress)
		assert.Equal(t, mint, *spec.MintAddress)
	})

	t.Run("ValidWithoutMint", func(t *testing.T) {
		// Nil is the only option for instructions whose raw data carries no
		// mint account at all (e.g. the unchecked Transfer), and a valid
		// choice for any other instruction that should match regardless of
		// mint.
		spec, err := NewSplNativeSpecification(SplInstructionTransfer, nil)
		assert.NoError(t, err)
		assert.Nil(t, spec.MintAddress)
	})

	t.Run("ValidToken2022Extension", func(t *testing.T) {
		spec, err := NewSplNativeSpecification(SplInstructionPausableExtension, &mint)
		assert.NoError(t, err)
		assert.Equal(t, SplInstructionPausableExtension, spec.Instruction)
	})

	t.Run("UnknownInstruction", func(t *testing.T) {
		_, err := NewSplNativeSpecification("frobnicate", &mint)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankInstruction", func(t *testing.T) {
		_, err := NewSplNativeSpecification("", &mint)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})

	t.Run("BlankMint", func(t *testing.T) {
		blank := "   "
		_, err := NewSplNativeSpecification(SplInstructionTransfer, &blank)
		assert.ErrorIs(t, err, domainerrors.ErrValidation)
	})
}

func TestSplNativeSpecification_Matches_AlwaysFalse(t *testing.T) {
	mint := "So11111111111111111111111111111111111111112"
	spec, err := NewSplNativeSpecification(SplInstructionTransfer, &mint)
	require.NoError(t, err)

	assert.False(t, spec.Matches(stubContractEvent{eventName: "transfer"}))
}

func TestIsKnownSplInstruction(t *testing.T) {
	assert.True(t, IsKnownSplInstruction(SplInstructionMintTo))
	assert.True(t, IsKnownSplInstruction(SplInstructionBurnChecked))
	assert.True(t, IsKnownSplInstruction(SplInstructionTransferHookExtension))
	assert.False(t, IsKnownSplInstruction("transferMaybe"))
}

func TestSplInstructionDiscriminator(t *testing.T) {
	t.Run("SingleByteInstruction", func(t *testing.T) {
		discriminator, extension, ok := SplInstructionDiscriminator(SplInstructionMintTo)
		assert.True(t, ok)
		assert.False(t, extension)
		assert.Equal(t, byte(7), discriminator)
	})

	t.Run("ExtensionFamily", func(t *testing.T) {
		// pausable's discriminator only identifies the family; pause vs resume
		// live in the following byte, which this specification does not model.
		discriminator, extension, ok := SplInstructionDiscriminator(SplInstructionPausableExtension)
		assert.True(t, ok)
		assert.True(t, extension)
		assert.Equal(t, byte(44), discriminator)
	})

	t.Run("Unknown", func(t *testing.T) {
		discriminator, extension, ok := SplInstructionDiscriminator("frobnicate")
		assert.False(t, ok)
		assert.False(t, extension)
		assert.Equal(t, byte(0), discriminator)
	})
}
