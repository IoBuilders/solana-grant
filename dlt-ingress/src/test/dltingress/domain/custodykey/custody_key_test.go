package custodykey

import (
	"dlt-ingress/src/main/dltingress/domain/common"
	"strings"
	"testing"

	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"

	"github.com/stretchr/testify/assert"
	coredomainerrors "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/coreerror"
)

func TestDltIngressNewCustodyKey_Success(t *testing.T) {
	tests := []struct {
		name            string
		keyType         string
		dlt             string
		custodyProvider string
	}{
		{
			name:            "ECDSA_SECP256K1 / EVM / DFNS",
			keyType:         string(common.ECDSASecp256k1),
			dlt:             string(common.EVM),
			custodyProvider: string(custodykey.CustodyProviderDFNS),
		},
		{
			name:            "ED25519 / SVM / DFNS",
			keyType:         string(common.ED25519),
			dlt:             string(common.SVM),
			custodyProvider: string(custodykey.CustodyProviderDFNS),
		},
		{
			name:            "ECDSA_SECP256K1 / Hashgraph / DFNS",
			keyType:         string(common.ECDSASecp256k1),
			dlt:             string(common.Hashgraph),
			custodyProvider: string(custodykey.CustodyProviderDFNS),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := custodykey.NewCustodyKey(
				tt.keyType,
				"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
				tt.dlt,
				"ext-id-123",
				tt.custodyProvider,
			)

			assert.Nil(t, err)
			assert.NotNil(t, key)
			assert.Equal(t, common.KeyType(tt.keyType), key.KeyType)
			assert.Equal(t, custodykey.StatusActive, key.Status)
			assert.Equal(t, "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12", key.DltAccountId)
			assert.Equal(t, common.Dlt(tt.dlt), key.Dlt)
			assert.Equal(t, "ext-id-123", key.ExternalId)
			assert.Equal(t, custodykey.CustodyProvider(tt.custodyProvider), key.CustodyProvider)
		})
	}
}

func TestDltIngressNewCustodyKey_InvalidKeyType(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		"INVALID_KEY_TYPE",
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		string(common.EVM),
		"ext-id-123",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidKeyType)
}

func TestDltIngressNewCustodyKey_InvalidDlt(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		"INVALID_DLT",
		"ext-id-123",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidDlt)
}

func TestDltIngressNewCustodyKey_InvalidCustodyProvider(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		string(common.EVM),
		"ext-id-123",
		"INVALID_PROVIDER",
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), domainerrors.ErrorCodeInvalidCustodyProvider)
}

func TestDltIngressNewCustodyKey_EmptyDltAccountId(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"",
		string(common.EVM),
		"ext-id-123",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewCustodyKey_EmptyExternalId(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		string(common.EVM),
		"",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewCustodyKey_DltAccountIdTooLong(t *testing.T) {
	tooLong := strings.Repeat("a", 256)

	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		tooLong,
		string(common.EVM),
		"ext-id-123",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewCustodyKey_ExternalIdTooLong(t *testing.T) {
	tooLong := strings.Repeat("a", 256)

	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		string(common.EVM),
		tooLong,
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, key)
	assert.NotNil(t, err)
	assert.Equal(t, err.(coreerror.DomainError).ErrorCode(), coredomainerrors.ErrorCodeValidation)
}

func TestDltIngressNewCustodyKey_DefaultStatusIsActive(t *testing.T) {
	key, err := custodykey.NewCustodyKey(
		string(common.ECDSASecp256k1),
		"0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
		string(common.EVM),
		"ext-id-123",
		string(custodykey.CustodyProviderDFNS),
	)

	assert.Nil(t, err)
	assert.Equal(t, custodykey.StatusActive, key.Status)
}

func TestDltIngressCustodyKey_Discontinue(t *testing.T) {
	factory := NewCustodyKeyTestFactory()
	key := factory.CreateEntity()

	assert.Equal(t, custodykey.StatusActive, key.Status)
	key.Discontinue()
	assert.Equal(t, custodykey.StatusDiscontinued, key.Status)
}

func TestDltIngressParseKeyType(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{string(common.ECDSASecp256k1), true},
		{string(common.ED25519), true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			kt, err := common.ParseKeyType(tt.value)
			if tt.valid {
				assert.Nil(t, err)
				assert.Equal(t, common.KeyType(tt.value), kt)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDltIngressParseDlt(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{string(common.EVM), true},
		{string(common.Hashgraph), true},
		{string(common.SVM), true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			d, err := common.ParseDlt(tt.value)
			if tt.valid {
				assert.Nil(t, err)
				assert.Equal(t, common.Dlt(tt.value), d)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDltIngressParseCustodyProvider(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{string(custodykey.CustodyProviderDFNS), true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			cp, err := custodykey.ParseCustodyProvider(tt.value)
			if tt.valid {
				assert.Nil(t, err)
				assert.Equal(t, custodykey.CustodyProvider(tt.value), cp)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestDltIngressParseStatus(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{string(custodykey.StatusActive), true},
		{string(custodykey.StatusDiscontinued), true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			s, err := custodykey.ParseStatus(tt.value)
			if tt.valid {
				assert.Nil(t, err)
				assert.Equal(t, custodykey.Status(tt.value), s)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
