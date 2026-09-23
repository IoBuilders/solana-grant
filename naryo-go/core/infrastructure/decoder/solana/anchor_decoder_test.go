//go:build test

package solanadecoder

import (
	"crypto/sha256"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/decoder"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/naryo-go/core/domain/filter/solana"
)

// discriminator computes the same sighash AnchorDecoder does, so tests
// exercise its routing/wiring rather than hand-derived hex.
func discriminator(t *testing.T, namespace, name string) []byte {
	t.Helper()
	sum := sha256.Sum256([]byte(namespace + ":" + name))
	return sum[:8]
}

func mustAnchorSpec(t *testing.T, source solana.AnchorSource, eventName string, params []solana.ParameterDefinition) *solana.AnchorSpecification {
	t.Helper()
	accountAddress := "SomeAccount"
	spec, err := solana.NewAnchorSpecification(source, &accountAddress, eventName, params)
	require.NoError(t, err)
	return spec
}

func TestAnchorDecoder_Decode(t *testing.T) {
	d := NewAnchorDecoder()

	t.Run("EmitLog_DiscriminatorOnlyMatch", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitLog, "Ping", nil)
		raw := decoder.RawPayload{Data: discriminator(t, "event", "Ping")}

		params, err := d.Decode(spec, raw)
		require.NoError(t, err)
		assert.NotNil(t, params)
		assert.Empty(t, params)
	})

	t.Run("EmitLog_DecodesParameters", func(t *testing.T) {
		uintDef, err := solana.NewUintParameterDefinition(0, 64)
		require.NoError(t, err)
		pubKeyDef, err := solana.NewPublicKeyParameterDefinition(1)
		require.NoError(t, err)
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitLog, "Transfer", []solana.ParameterDefinition{uintDef, pubKeyDef})

		data := append([]byte{}, discriminator(t, "event", "Transfer")...)
		data = append(data, 0x2C, 0x01, 0, 0, 0, 0, 0, 0) // uint64 amount = 300
		data = append(data, make([]byte, 32)...)          // pubkey
		raw := decoder.RawPayload{Data: data}

		params, err := d.Decode(spec, raw)
		require.NoError(t, err)
		require.Len(t, params, 2)
		assert.Equal(t, big.NewInt(300), params[0].Value())
	})

	t.Run("EmitLog_DiscriminatorMismatch", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitLog, "Ping", nil)
		raw := decoder.RawPayload{Data: discriminator(t, "event", "Pong")}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("EmitCPI_SentinelAndDiscriminatorMatch", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitCPI, "Ping", nil)
		data := append([]byte{}, anchorEventCPISentinel[:]...)
		data = append(data, discriminator(t, "event", "Ping")...)
		raw := decoder.RawPayload{Data: data}

		params, err := d.Decode(spec, raw)
		require.NoError(t, err)
		assert.NotNil(t, params)
		assert.Empty(t, params)
	})

	t.Run("EmitCPI_SentinelPresentButDiscriminatorWrong", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitCPI, "Ping", nil)
		data := append([]byte{}, anchorEventCPISentinel[:]...)
		data = append(data, discriminator(t, "event", "Pong")...)
		raw := decoder.RawPayload{Data: data}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("EmitCPI_SentinelWrong", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitCPI, "Ping", nil)
		data := make([]byte, 8) // all zero, not the real sentinel
		data = append(data, discriminator(t, "event", "Ping")...)
		raw := decoder.RawPayload{Data: data}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("Instruction_GlobalNamespaceMatch", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceInstruction, "initialize", nil)
		raw := decoder.RawPayload{Data: discriminator(t, "global", "initialize")}

		params, err := d.Decode(spec, raw)
		require.NoError(t, err)
		assert.NotNil(t, params)
		assert.Empty(t, params)
	})

	t.Run("EmitCPI_PayloadShorterThanSentinel", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitCPI, "Ping", nil)
		raw := decoder.RawPayload{Data: []byte{0x01, 0x02}}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("EmitLog_PayloadShorterThanDiscriminator", func(t *testing.T) {
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitLog, "Ping", nil)
		raw := decoder.RawPayload{Data: []byte{0x01, 0x02}}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("NonAnchorSpecification", func(t *testing.T) {
		str := "SomeMint"
		spec, err := solana.NewSplNativeSpecification(solana.SplInstructionTransfer, &str)
		require.NoError(t, err)
		raw := decoder.RawPayload{Data: discriminator(t, "event", "Ping")}

		params, err := d.Decode(spec, raw)
		assert.NoError(t, err)
		assert.Nil(t, params)
	})

	t.Run("DiscriminatorMatchesButBufferUnderflows", func(t *testing.T) {
		uintDef, err := solana.NewUintParameterDefinition(0, 64)
		require.NoError(t, err)
		spec := mustAnchorSpec(t, solana.AnchorSourceEmitLog, "Transfer", []solana.ParameterDefinition{uintDef})

		data := append([]byte{}, discriminator(t, "event", "Transfer")...)
		data = append(data, 0x01, 0x02) // only 2 of the required 8 bytes
		raw := decoder.RawPayload{Data: data}

		params, err := d.Decode(spec, raw)
		assert.Nil(t, params)
		assert.ErrorIs(t, err, domainerrors.ErrDecoding)
	})
}
