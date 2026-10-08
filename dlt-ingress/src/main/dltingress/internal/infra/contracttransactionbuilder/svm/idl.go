package svmidl

import (
	"encoding/json"
	"fmt"
)

type Idl struct {
	Address      string        `json:"address"`
	Metadata     IdlMetadata   `json:"metadata"`
	Instructions []Instruction `json:"instructions"`
	Accounts     []AccountDef  `json:"accounts,omitempty"`
	Types        []TypeDef     `json:"types,omitempty"`
}

type IdlMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Spec    string `json:"spec"`
}

type Instruction struct {
	Name          string           `json:"name"`
	Discriminator []byte           `json:"discriminator"`
	Accounts      []AccountMetaDef `json:"accounts"`
	Args          []Field          `json:"args"`
}

type AccountMetaDef struct {
	Name     string  `json:"name"`
	Writable bool    `json:"writable,omitempty"`
	Signer   bool    `json:"signer,omitempty"`
	Optional bool    `json:"optional,omitempty"`
	Address  string  `json:"address,omitempty"`
	PDA      *PDADef `json:"pda,omitempty"`
}

type PDADef struct {
	Seeds   []Seed `json:"seeds"`
	Program *Seed  `json:"program,omitempty"`
}

type Seed struct {
	Kind  string   `json:"kind"`
	Value []byte   `json:"value,omitempty"`
	Path  string   `json:"path,omitempty"`
	Type  *IdlType `json:"type,omitempty"`
}

type AccountDef struct {
	Name          string `json:"name"`
	Discriminator []byte `json:"discriminator,omitempty"`
}

type TypeDef struct {
	Name string  `json:"name"`
	Type IdlType `json:"type"`
}

type Field struct {
	Name string  `json:"name"`
	Type IdlType `json:"type"`
}

type Kind string

const (
	KindBool    Kind = "bool"
	KindU8      Kind = "u8"
	KindU16     Kind = "u16"
	KindU32     Kind = "u32"
	KindU64     Kind = "u64"
	KindU128    Kind = "u128"
	KindI8      Kind = "i8"
	KindI16     Kind = "i16"
	KindI32     Kind = "i32"
	KindI64     Kind = "i64"
	KindI128    Kind = "i128"
	KindF32     Kind = "f32"
	KindF64     Kind = "f64"
	KindString  Kind = "string"
	KindPubkey  Kind = "pubkey"
	KindBytes   Kind = "bytes"
	KindVec     Kind = "vec"
	KindOption  Kind = "option"
	KindArray   Kind = "array"
	KindStruct  Kind = "struct"
	KindEnum    Kind = "enum"
	KindDefined Kind = "defined"
)

type IdlType struct {
	Kind     Kind
	Inner    *IdlType
	Length   int
	Defined  string
	Fields   []Field
	Variants []EnumVariant
}

type EnumVariant struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields,omitempty"`
}

func (t *IdlType) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		k, ok := primitiveKind(asString)
		if !ok {
			return fmt.Errorf("unknown primitive IDL type %q", asString)
		}
		t.Kind = k
		return nil
	}

	var asObject map[string]json.RawMessage
	if err := json.Unmarshal(data, &asObject); err != nil {
		return fmt.Errorf("IDL type: not a string or object: %w", err)
	}

	if raw, ok := asObject["vec"]; ok {
		inner := &IdlType{}
		if err := json.Unmarshal(raw, inner); err != nil {
			return fmt.Errorf("vec inner: %w", err)
		}
		t.Kind = KindVec
		t.Inner = inner
		return nil
	}
	if raw, ok := asObject["option"]; ok {
		inner := &IdlType{}
		if err := json.Unmarshal(raw, inner); err != nil {
			return fmt.Errorf("option inner: %w", err)
		}
		t.Kind = KindOption
		t.Inner = inner
		return nil
	}
	if raw, ok := asObject["array"]; ok {
		var tuple []json.RawMessage
		if err := json.Unmarshal(raw, &tuple); err != nil || len(tuple) != 2 {
			return fmt.Errorf("array: expected [type, length] tuple")
		}
		inner := &IdlType{}
		if err := json.Unmarshal(tuple[0], inner); err != nil {
			return fmt.Errorf("array inner: %w", err)
		}
		var length int
		if err := json.Unmarshal(tuple[1], &length); err != nil {
			return fmt.Errorf("array length: %w", err)
		}
		t.Kind = KindArray
		t.Inner = inner
		t.Length = length
		return nil
	}
	if raw, ok := asObject["defined"]; ok {
		var ref struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &ref); err != nil {
			var name string
			if err2 := json.Unmarshal(raw, &name); err2 != nil {
				return fmt.Errorf("defined: %w", err)
			}
			ref.Name = name
		}
		t.Kind = KindDefined
		t.Defined = ref.Name
		return nil
	}
	if kindRaw, ok := asObject["kind"]; ok {
		var kind string
		if err := json.Unmarshal(kindRaw, &kind); err != nil {
			return fmt.Errorf("kind: %w", err)
		}
		switch kind {
		case "struct":
			var s struct {
				Fields []Field `json:"fields"`
			}
			if err := json.Unmarshal(data, &s); err != nil {
				return fmt.Errorf("struct fields: %w", err)
			}
			t.Kind = KindStruct
			t.Fields = s.Fields
			return nil
		case "enum":
			var e struct {
				Variants []EnumVariant `json:"variants"`
			}
			if err := json.Unmarshal(data, &e); err != nil {
				return fmt.Errorf("enum variants: %w", err)
			}
			t.Kind = KindEnum
			t.Variants = e.Variants
			return nil
		default:
			return fmt.Errorf("unsupported IDL type kind %q", kind)
		}
	}

	return fmt.Errorf("unrecognized IDL type shape: %s", string(data))
}

func primitiveKind(s string) (Kind, bool) {
	switch s {
	case "bool":
		return KindBool, true
	case "u8":
		return KindU8, true
	case "u16":
		return KindU16, true
	case "u32":
		return KindU32, true
	case "u64":
		return KindU64, true
	case "u128":
		return KindU128, true
	case "i8":
		return KindI8, true
	case "i16":
		return KindI16, true
	case "i32":
		return KindI32, true
	case "i64":
		return KindI64, true
	case "i128":
		return KindI128, true
	case "f32":
		return KindF32, true
	case "f64":
		return KindF64, true
	case "string":
		return KindString, true
	case "pubkey", "publicKey":
		return KindPubkey, true
	case "bytes":
		return KindBytes, true
	}
	return "", false
}

func ParseIdl(data []byte) (*Idl, error) {
	idl := &Idl{}
	if err := json.Unmarshal(data, idl); err != nil {
		return nil, fmt.Errorf("parsing IDL: %w", err)
	}
	return idl, nil
}

func (i *Idl) FindInstruction(name string) (*Instruction, error) {
	for idx := range i.Instructions {
		if i.Instructions[idx].Name == name {
			return &i.Instructions[idx], nil
		}
	}
	return nil, fmt.Errorf("instruction %q not found in IDL %q", name, i.Metadata.Name)
}

func (i *Idl) ResolveDefined(name string) (*TypeDef, error) {
	for idx := range i.Types {
		if i.Types[idx].Name == name {
			return &i.Types[idx], nil
		}
	}
	return nil, fmt.Errorf("defined type %q not found in IDL %q", name, i.Metadata.Name)
}
