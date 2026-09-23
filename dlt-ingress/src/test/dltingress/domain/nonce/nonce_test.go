package nonce

import (
	"dlt-ingress/src/main/dltingress/domain/nonce"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/domain/domainerrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/amount"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestDltIngressNewNonce_Success(t *testing.T) {
	tests := []struct {
		name         string
		dltAccountId string
		networkId    string
	}{
		{
			name:         "with valid length",
			dltAccountId: "random-dltAccountId",
			networkId:    "random-networkId",
		},
		{
			name:         "with max length of dltAccountId",
			dltAccountId: strings.Repeat("a", 255),
			networkId:    "random-networkId",
		},
		{
			name:         "with max length of networkId",
			dltAccountId: "random-dltAccountId",
			networkId:    strings.Repeat("a", 255),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := amount.Zero()
			createdNonce, err := nonce.NewNonce(
				tt.dltAccountId,
				tt.networkId,
				value,
			)

			assert.Nil(t, err)
			assert.NotNil(t, createdNonce)
			assert.Equal(t, tt.dltAccountId, createdNonce.DltAccountId)
			assert.Equal(t, tt.networkId, createdNonce.NetworkId)
			assert.Equal(t, value, createdNonce.Value)
		})
	}
}

func TestDltIngressNewNonce_InvalidEmptyDltAccountId(t *testing.T) {
	createdNonce, err := nonce.NewNonce(
		"",
		"random-networkId",
		amount.Zero(),
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "DltAccountId")
	assert.ErrorContains(t, err, "Nonce")
}

func TestDltIngressNewNonce_InvalidTooLongDltAccountId(t *testing.T) {
	createdNonce, err := nonce.NewNonce(
		strings.Repeat("a", 256),
		"random-networkId",
		amount.Zero(),
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "DltAccountId")
	assert.ErrorContains(t, err, "Nonce")
}

func TestDltIngressNewNonce_InvalidEmptyNetworkId(t *testing.T) {
	createdNonce, err := nonce.NewNonce(
		"random-dltAccountId",
		"",
		amount.Zero(),
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkId")
	assert.ErrorContains(t, err, "Nonce")
}

func TestDltIngressNewNonce_InvalidTooLongNetworkId(t *testing.T) {
	createdNonce, err := nonce.NewNonce(
		"random-dltAccountId",
		strings.Repeat("a", 256),
		amount.Zero(),
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
	assert.ErrorContains(t, err, "NetworkId")
	assert.ErrorContains(t, err, "Nonce")
}

func TestDltIngressNewNonce_InvalidValueWithDecimals(t *testing.T) {
	invalidValue, _ := amount.NewFromString("1.1")
	createdNonce, err := nonce.NewNonce(
		"random-dltAccountId",
		"random-networkId",
		invalidValue,
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	assert.ErrorContains(t, err, invalidValue.String())
}

func TestDltIngressNewNonce_InvalidValueNegative(t *testing.T) {
	invalidValue, _ := amount.NewFromString("-1")
	createdNonce, err := nonce.NewNonce(
		"random-dltAccountId",
		"random-networkId",
		invalidValue,
	)

	assert.Nil(t, createdNonce)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	assert.ErrorContains(t, err, invalidValue.String())
}

func TestDltIngressNonceUpdateValue_InvalidValueWithDecimals(t *testing.T) {
	createdNonce := NewNonceTestFactory().CreateEntity()
	invalidValue, _ := amount.NewFromString("1.1")

	err := createdNonce.UpdateValue(invalidValue)

	require.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	assert.ErrorContains(t, err, invalidValue.String())
}

func TestDltIngressNonceUpdateValue_InvalidValueNegative(t *testing.T) {
	createdNonce := NewNonceTestFactory().CreateEntity()
	invalidValue, _ := amount.NewFromString("-1")

	err := createdNonce.UpdateValue(invalidValue)

	require.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidNonceValue)
	assert.ErrorContains(t, err, invalidValue.String())
}
