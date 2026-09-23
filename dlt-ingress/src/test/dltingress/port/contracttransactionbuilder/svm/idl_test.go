package svmidl

import (
	"dlt-ingress/src/main/dltingress/port/contracttransactionbuilder/svm"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleIdlJSON = `{
  "address": "Stub1111111111111111111111111111111111111111",
  "metadata": { "name": "sample", "version": "0.1.0", "spec": "0.1.0" },
  "instructions": [
    {
      "name": "initialize",
      "discriminator": [175, 175, 109, 31, 13, 152, 155, 237],
      "accounts": [
        { "name": "user", "writable": true, "signer": true },
        { "name": "vault", "writable": true, "pda": {
            "seeds": [
              { "kind": "const", "value": [118, 97, 117, 108, 116] },
              { "kind": "account", "path": "user" },
              { "kind": "arg", "path": "id" }
            ]
        }},
        { "name": "systemProgram", "address": "11111111111111111111111111111111" }
      ],
      "args": [
        { "name": "id", "type": "u64" },
        { "name": "amount", "type": "u128" },
        { "name": "memo", "type": "string" },
        { "name": "tags", "type": { "vec": "string" } },
        { "name": "note", "type": { "option": "u32" } },
        { "name": "seed", "type": { "array": ["u8", 32] } },
        { "name": "config", "type": { "defined": { "name": "Config" } } }
      ]
    }
  ],
  "accounts": [
    { "name": "Vault", "discriminator": [1, 2, 3, 4, 5, 6, 7, 8] }
  ],
  "types": [
    {
      "name": "Config",
      "type": {
        "kind": "struct",
        "fields": [
          { "name": "owner", "type": "pubkey" },
          { "name": "active", "type": "bool" }
        ]
      }
    },
    {
      "name": "Status",
      "type": {
        "kind": "enum",
        "variants": [
          { "name": "Pending" },
          { "name": "Done", "fields": [ { "name": "ts", "type": "i64" } ] }
        ]
      }
    }
  ]
}`

func TestDltIngressParseIdl_TopLevelFields(t *testing.T) {
	idl, err := svmidl.ParseIdl([]byte(sampleIdlJSON))

	require.NoError(t, err)
	assert.Equal(t, "Stub1111111111111111111111111111111111111111", idl.Address)
	assert.Equal(t, "sample", idl.Metadata.Name)
	assert.Len(t, idl.Instructions, 1)
	assert.Len(t, idl.Accounts, 1)
	assert.Len(t, idl.Types, 2)
}

func TestDltIngressParseIdl_InstructionDiscriminatorAndArgs(t *testing.T) {
	idl, err := svmidl.ParseIdl([]byte(sampleIdlJSON))
	require.NoError(t, err)

	ix, err := idl.FindInstruction("initialize")
	require.NoError(t, err)

	assert.Equal(t, []byte{175, 175, 109, 31, 13, 152, 155, 237}, ix.Discriminator)
	require.Len(t, ix.Args, 7)
	assert.Equal(t, svmidl.KindU64, ix.Args[0].Type.Kind)
	assert.Equal(t, svmidl.KindU128, ix.Args[1].Type.Kind)
	assert.Equal(t, svmidl.KindString, ix.Args[2].Type.Kind)
}

func TestDltIngressParseIdl_VecOptionArrayDefined(t *testing.T) {
	idl, _ := svmidl.ParseIdl([]byte(sampleIdlJSON))
	ix, _ := idl.FindInstruction("initialize")

	vec := ix.Args[3].Type
	assert.Equal(t, svmidl.KindVec, vec.Kind)
	require.NotNil(t, vec.Inner)
	assert.Equal(t, svmidl.KindString, vec.Inner.Kind)

	opt := ix.Args[4].Type
	assert.Equal(t, svmidl.KindOption, opt.Kind)
	require.NotNil(t, opt.Inner)
	assert.Equal(t, svmidl.KindU32, opt.Inner.Kind)

	arr := ix.Args[5].Type
	assert.Equal(t, svmidl.KindArray, arr.Kind)
	assert.Equal(t, 32, arr.Length)
	assert.Equal(t, svmidl.KindU8, arr.Inner.Kind)

	def := ix.Args[6].Type
	assert.Equal(t, svmidl.KindDefined, def.Kind)
	assert.Equal(t, "Config", def.Defined)
}

func TestDltIngressParseIdl_AccountsShapes(t *testing.T) {
	idl, _ := svmidl.ParseIdl([]byte(sampleIdlJSON))
	ix, _ := idl.FindInstruction("initialize")
	require.Len(t, ix.Accounts, 3)

	user := ix.Accounts[0]
	assert.Equal(t, "user", user.Name)
	assert.True(t, user.Writable)
	assert.True(t, user.Signer)
	assert.Nil(t, user.PDA)
	assert.Empty(t, user.Address)

	vault := ix.Accounts[1]
	require.NotNil(t, vault.PDA)
	require.Len(t, vault.PDA.Seeds, 3)
	assert.Equal(t, "const", vault.PDA.Seeds[0].Kind)
	assert.Equal(t, []byte{118, 97, 117, 108, 116}, vault.PDA.Seeds[0].Value)
	assert.Equal(t, "account", vault.PDA.Seeds[1].Kind)
	assert.Equal(t, "user", vault.PDA.Seeds[1].Path)
	assert.Equal(t, "arg", vault.PDA.Seeds[2].Kind)
	assert.Equal(t, "id", vault.PDA.Seeds[2].Path)

	sysProg := ix.Accounts[2]
	assert.Equal(t, "11111111111111111111111111111111", sysProg.Address)
}

func TestDltIngressParseIdl_ResolveDefinedStructAndEnum(t *testing.T) {
	idl, _ := svmidl.ParseIdl([]byte(sampleIdlJSON))

	config, err := idl.ResolveDefined("Config")
	require.NoError(t, err)
	assert.Equal(t, svmidl.KindStruct, config.Type.Kind)
	require.Len(t, config.Type.Fields, 2)
	assert.Equal(t, "owner", config.Type.Fields[0].Name)
	assert.Equal(t, svmidl.KindPubkey, config.Type.Fields[0].Type.Kind)
	assert.Equal(t, svmidl.KindBool, config.Type.Fields[1].Type.Kind)

	status, err := idl.ResolveDefined("Status")
	require.NoError(t, err)
	assert.Equal(t, svmidl.KindEnum, status.Type.Kind)
	require.Len(t, status.Type.Variants, 2)
	assert.Equal(t, "Pending", status.Type.Variants[0].Name)
	assert.Empty(t, status.Type.Variants[0].Fields)
	assert.Equal(t, "Done", status.Type.Variants[1].Name)
	require.Len(t, status.Type.Variants[1].Fields, 1)
	assert.Equal(t, svmidl.KindI64, status.Type.Variants[1].Fields[0].Type.Kind)
}

func TestDltIngressParseIdl_UnknownPrimitive(t *testing.T) {
	bad := `{"address":"x","metadata":{"name":"x","version":"0","spec":"0"},
		"instructions":[{"name":"i","discriminator":[0],"accounts":[],"args":[{"name":"a","type":"u256"}]}]}`
	_, err := svmidl.ParseIdl([]byte(bad))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown primitive IDL type")
}

func TestDltIngressParseIdl_InstructionNotFound(t *testing.T) {
	idl, _ := svmidl.ParseIdl([]byte(sampleIdlJSON))
	_, err := idl.FindInstruction("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "instruction \"missing\" not found")
}

func TestDltIngressParseIdl_DefinedNotFound(t *testing.T) {
	idl, _ := svmidl.ParseIdl([]byte(sampleIdlJSON))
	_, err := idl.ResolveDefined("Nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "defined type \"Nope\" not found")
}
