package parameter

// Type identifies the concrete shape of a ContractEventParameter's value.
type Type string

const (
	TypeBool       Type = "BOOL"
	TypeInt        Type = "INT"
	TypeUint       Type = "UINT"
	TypeFloat      Type = "FLOAT"
	TypeString     Type = "STRING"
	TypeBytes      Type = "BYTES"
	TypeBytesFixed Type = "BYTES_FIXED"
	TypePublicKey  Type = "PUBLIC_KEY"
	TypeArray      Type = "ARRAY"
	TypeStruct     Type = "STRUCT"
	TypeOption     Type = "OPTION"
)

func (t Type) IsValid() bool {
	switch t {
	case TypeBool, TypeInt, TypeUint, TypeFloat, TypeString, TypeBytes, TypeBytesFixed,
		TypePublicKey, TypeArray, TypeStruct, TypeOption:
		return true
	}
	return false
}

func (t Type) String() string {
	return string(t)
}
