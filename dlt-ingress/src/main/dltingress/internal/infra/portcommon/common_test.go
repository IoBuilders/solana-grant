package portcommon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDltIngressSVMTransactionResponse_WritableAccountKeys(t *testing.T) {
	tests := []struct {
		name        string
		header      SVMMessageHeader
		accountKeys []string
		expected    []string
	}{
		{
			name:        "fee payer only",
			header:      SVMMessageHeader{NumRequiredSignatures: 1},
			accountKeys: []string{"payer"},
			expected:    []string{"payer"},
		},
		{
			name:        "fee payer and readonly program",
			header:      SVMMessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
			accountKeys: []string{"payer", "program"},
			expected:    []string{"payer"},
		},
		{
			name: "one account of each kind",
			header: SVMMessageHeader{
				NumRequiredSignatures:       2,
				NumReadonlySignedAccounts:   1,
				NumReadonlyUnsignedAccounts: 1,
			},
			accountKeys: []string{"payer", "signedReadonly", "unsignedWritable", "program"},
			expected:    []string{"payer", "unsignedWritable"},
		},
		{
			name: "several signed and unsigned writable accounts",
			header: SVMMessageHeader{
				NumRequiredSignatures:       3,
				NumReadonlySignedAccounts:   1,
				NumReadonlyUnsignedAccounts: 2,
			},
			accountKeys: []string{"payer", "cosigner", "signedReadonly", "pda", "recipient", "mint", "program"},
			expected:    []string{"payer", "cosigner", "pda", "recipient"},
		},
		{
			name:        "no accounts",
			header:      SVMMessageHeader{},
			accountKeys: []string{},
			expected:    []string{},
		},
		{
			name: "header counts beyond account keys are clamped",
			header: SVMMessageHeader{
				NumRequiredSignatures:       5,
				NumReadonlySignedAccounts:   9,
				NumReadonlyUnsignedAccounts: 9,
			},
			accountKeys: []string{"payer", "other"},
			expected:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := &SVMTransactionResponse{Header: tt.header, AccountKeys: tt.accountKeys}

			assert.Equal(t, tt.expected, tx.WritableAccountKeys())
		})
	}
}
