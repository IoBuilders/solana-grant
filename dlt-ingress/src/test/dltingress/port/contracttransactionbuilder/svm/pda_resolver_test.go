package svmidl

import (
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder/svm"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testProgramID  = solana.MustPublicKeyFromBase58("11111111111111111111111111111112")
	systemProgID   = "11111111111111111111111111111111"
	testSender     = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	testOtherInput = solana.MustPublicKeyFromBase58("4Nd1mYQbHmZ2HhZJ4uEjvR5tQv3qHN3F8XJyR9KkPp2C")
)

func resolver() *svmidl.PdaResolver {
	return svmidl.NewPdaResolver(testProgramID)
}

func TestDltIngressResolve_InputAccount_FromArgs(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "recipient", Writable: true},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{"recipient": testOtherInput.String()}, testSender)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, testOtherInput, out[0].PublicKey)
	assert.True(t, out[0].IsWritable)
	assert.False(t, out[0].IsSigner)
}

func TestDltIngressResolve_SignerFallsBackToSender(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "authority", Signer: true, Writable: true},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, testSender, out[0].PublicKey)
	assert.True(t, out[0].IsSigner)
}

func TestDltIngressResolve_MissingAccountErrors(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "missing"},
		},
	}
	_, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `account "missing"`)
	assert.Contains(t, err.Error(), "not provided")
}

func TestDltIngressResolve_InputAccount_WrongTypeErrors(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{{Name: "x"}},
	}
	_, err := resolver().Resolve(ix, map[string]any{"x": "not-a-key"}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a Solana pub key")
}

func TestDltIngressResolve_FixedAddress(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "systemProgram", Address: systemProgID},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.NoError(t, err)
	assert.Equal(t, solana.MustPublicKeyFromBase58(systemProgID), out[0].PublicKey)
}

func TestDltIngressResolve_FixedAddress_InvalidBase58(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "bad", Address: "not-base58!!!"},
		},
	}
	_, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid fixed address")
}

func TestDltIngressResolve_PDA_ConstSeedOnly(t *testing.T) {
	seedBytes := []byte("vault")
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "vault", Writable: true,
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: seedBytes},
					},
				},
			},
		},
	}

	out, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.NoError(t, err)
	expected, _, _ := solana.FindProgramAddress([][]byte{seedBytes}, testProgramID)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_PDA_AccountSeedReferencesEarlierAccount(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "user", Signer: true},
			{
				Name: "vault", Writable: true,
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: []byte("vault")},
						{Kind: "account", Path: "user"},
					},
				},
			},
		},
	}

	out, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.NoError(t, err)
	require.Len(t, out, 2)

	expected, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("vault"), testSender[:]}, testProgramID,
	)
	assert.Equal(t, testSender, out[0].PublicKey)
	assert.Equal(t, expected, out[1].PublicKey)
}

func TestDltIngressResolve_PDA_ArgSeed_U64(t *testing.T) {
	u64 := svmidl.IdlType{Kind: svmidl.KindU64}
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "pos",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: []byte("p")},
						{Kind: "arg", Path: "id", Type: &u64},
					},
				},
			},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{"id": uint64(7)}, testSender)
	require.NoError(t, err)

	idBytes := []byte{0x07, 0, 0, 0, 0, 0, 0, 0}
	expected, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("p"), idBytes}, testProgramID,
	)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_PDA_ArgSeed_String(t *testing.T) {
	str := svmidl.IdlType{Kind: svmidl.KindString}
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "named",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "arg", Path: "name", Type: &str},
					},
				},
			},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{"name": "hello"}, testSender)
	require.NoError(t, err)

	expected, _, _ := solana.FindProgramAddress([][]byte{[]byte("hello")}, testProgramID)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_PDA_ArgSeed_InfersTypeFromInstructionArgs(t *testing.T) {
	ix := &svmidl.Instruction{
		Args: []svmidl.Field{
			{Name: "coupon_id", Type: svmidl.IdlType{Kind: svmidl.KindU64}},
		},
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "coupon",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: []byte("coupon")},
						{Kind: "arg", Path: "coupon_id"},
					},
				},
			},
		},
	}

	out, err := resolver().Resolve(ix, map[string]any{"coupon_id": uint64(7)}, testSender)
	require.NoError(t, err)
	require.Len(t, out, 1)

	idBytes := []byte{0x07, 0, 0, 0, 0, 0, 0, 0}
	expected, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("coupon"), idBytes}, testProgramID,
	)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_PDA_ArgSeed_InferredTypeMissingFromInstructionArgsErrors(t *testing.T) {
	ix := &svmidl.Instruction{
		Args: []svmidl.Field{
			{Name: "other", Type: svmidl.IdlType{Kind: svmidl.KindU64}},
		},
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "coupon",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "arg", Path: "coupon_id"},
					},
				},
			},
		},
	}

	_, err := resolver().Resolve(ix, map[string]any{"coupon_id": uint64(7)}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `coupon_id`)
}

func TestDltIngressResolve_PDA_ArgSeed_MissingArgErrors(t *testing.T) {
	str := svmidl.IdlType{Kind: svmidl.KindString}
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "x",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "arg", Path: "missing", Type: &str},
					},
				},
			},
		},
	}
	_, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `arg seed "missing" references missing arg`)
}

func TestDltIngressResolve_PDA_ArgSeed_ExplicitTypeStillWorks(t *testing.T) {
	u64 := svmidl.IdlType{Kind: svmidl.KindU64}
	ix := &svmidl.Instruction{
		Args: []svmidl.Field{
			{Name: "coupon_id", Type: svmidl.IdlType{Kind: svmidl.KindString}},
		},
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "coupon",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: []byte("coupon")},
						{Kind: "arg", Path: "coupon_id", Type: &u64},
					},
				},
			},
		},
	}

	out, err := resolver().Resolve(ix, map[string]any{"coupon_id": uint64(9)}, testSender)
	require.NoError(t, err)
	require.Len(t, out, 1)

	idBytes := []byte{0x09, 0, 0, 0, 0, 0, 0, 0}
	expected, _, _ := solana.FindProgramAddress(
		[][]byte{[]byte("coupon"), idBytes}, testProgramID,
	)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_PDA_AccountSeed_UnresolvedReference(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "vault",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "account", Path: "user"},
					},
				},
			},
		},
	}
	_, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `account seed "user" references unresolved account`)
}

func TestDltIngressResolve_PDA_UnknownSeedKindErrors(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "x",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "weird"},
					},
				},
			},
		},
	}
	_, err := resolver().Resolve(ix, map[string]any{}, testSender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown seed kind "weird"`)
}

func TestDltIngressResolve_PDA_ProgramOverride(t *testing.T) {
	otherProgram := solana.MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111")
	pubkey := svmidl.IdlType{Kind: svmidl.KindPubkey}
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{
				Name: "external",
				PDA: &svmidl.PDADef{
					Seeds: []svmidl.Seed{
						{Kind: "const", Value: []byte("ext")},
					},
					Program: &svmidl.Seed{
						Kind: "arg", Path: "prog", Type: &pubkey,
					},
				},
			},
		},
	}

	out, err := resolver().Resolve(ix, map[string]any{"prog": otherProgram}, testSender)
	require.NoError(t, err)

	expected, _, _ := solana.FindProgramAddress([][]byte{[]byte("ext")}, otherProgram)
	assert.Equal(t, expected, out[0].PublicKey)
}

func TestDltIngressResolve_AccountMetaFlags(t *testing.T) {
	ix := &svmidl.Instruction{
		Accounts: []svmidl.AccountMetaDef{
			{Name: "a", Writable: true, Signer: true},
			{Name: "b"},
		},
	}
	out, err := resolver().Resolve(ix, map[string]any{
		"a": testSender.String(),
		"b": testOtherInput.String(),
	}, testSender)
	require.NoError(t, err)

	assert.True(t, out[0].IsWritable)
	assert.True(t, out[0].IsSigner)
	assert.False(t, out[1].IsWritable)
	assert.False(t, out[1].IsSigner)
}
