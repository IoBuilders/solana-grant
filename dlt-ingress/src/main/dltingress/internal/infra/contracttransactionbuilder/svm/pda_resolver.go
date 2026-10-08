package svmidl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

// errAccountInputMissing marks resolution failures caused by an input the caller
// did not provide, which an optional account is allowed to omit.
var errAccountInputMissing = errors.New("account input missing")

type PdaResolver struct {
	programID solana.PublicKey
}

func NewPdaResolver(programID solana.PublicKey) *PdaResolver {
	return &PdaResolver{programID: programID}
}

func (r *PdaResolver) Resolve(
	ix *Instruction,
	args map[string]any,
	sender solana.PublicKey,
) (solana.AccountMetaSlice, error) {
	resolved := make(map[string]solana.PublicKey, len(ix.Accounts))
	out := make(solana.AccountMetaSlice, 0, len(ix.Accounts))

	// Iteration 1: resolve fixed args so they're available for PDAs that reference them
	for _, a := range ix.Accounts {
		if v, ok := args[a.Name]; ok {
			s, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("args[%q] must be a string, got %T", a.Name, v)
			}
			pk, err := solana.PublicKeyFromBase58(s)
			if err != nil {
				return nil, fmt.Errorf("arg %s must be a Solana pub key: %s", a.Name, s)
			}
			resolved[a.Name] = pk
		}
	}

	// Iteration 2: resolve all accounts in IDL order (PDAs can now reference any
	// account populated in pass 1) and build the output slice.
	for _, a := range ix.Accounts {
		omitted := false
		if _, alreadyResolved := resolved[a.Name]; !alreadyResolved {
			pk, err := r.resolveOne(ix, a, args, resolved, sender)
			switch {
			case err == nil:
			case a.Optional && errors.Is(err, errAccountInputMissing):
				// Anchor convention: an omitted optional account is passed as the program ID.
				pk, omitted = r.programID, true
			default:
				return nil, fmt.Errorf("account %q: %w", a.Name, err)
			}
			resolved[a.Name] = pk
		}

		out = append(out, &solana.AccountMeta{
			PublicKey:  resolved[a.Name],
			IsWritable: a.Writable && !omitted,
			IsSigner:   a.Signer && !omitted,
		})
	}
	return out, nil
}

func (r *PdaResolver) resolveOne(
	ix *Instruction,
	a AccountMetaDef,
	args map[string]any,
	resolved map[string]solana.PublicKey,
	sender solana.PublicKey,
) (solana.PublicKey, error) {
	switch {
	case a.PDA != nil:
		seeds, err := r.materializeSeeds(ix, a.PDA.Seeds, args, resolved)
		if err != nil {
			return solana.PublicKey{}, err
		}
		programID := r.programID
		if a.PDA.Program != nil {
			progBytes, err := r.materializeSeed(ix, *a.PDA.Program, args, resolved)
			if err != nil {
				return solana.PublicKey{}, fmt.Errorf("pda program override: %w", err)
			}
			if len(progBytes) != solana.PublicKeyLength {
				return solana.PublicKey{}, fmt.Errorf(
					"pda program override must be %d bytes, got %d",
					solana.PublicKeyLength, len(progBytes),
				)
			}
			programID = solana.PublicKeyFromBytes(progBytes)
		}
		pk, _, err := solana.FindProgramAddress(seeds, programID)
		if err != nil {
			return solana.PublicKey{}, fmt.Errorf("derive PDA: %w", err)
		}
		return pk, nil
	case a.Address != "":
		pk, err := solana.PublicKeyFromBase58(a.Address)
		if err != nil {
			return solana.PublicKey{}, fmt.Errorf("invalid fixed address: %w", err)
		}
		return pk, nil
	default:
		if v, ok := args[a.Name]; ok {
			s, ok := v.(string)
			if !ok {
				return solana.PublicKey{}, fmt.Errorf(
					"args[%q] must be a string, got %T", a.Name, v,
				)
			}
			pk, err := solana.PublicKeyFromBase58(s)
			if err != nil {
				return solana.PublicKey{}, fmt.Errorf("args[%q] must be a Solana pub key: %w", a.Name, err)
			}
			return pk, nil
		}
		if a.Signer {
			return sender, nil
		}
		return solana.PublicKey{}, fmt.Errorf("%s not provided in args and not derivable: %w", a.Name, errAccountInputMissing)
	}
}

func (r *PdaResolver) materializeSeeds(
	ix *Instruction,
	seeds []Seed,
	args map[string]any,
	resolved map[string]solana.PublicKey,
) ([][]byte, error) {
	out := make([][]byte, 0, len(seeds))
	for i, s := range seeds {
		b, err := r.materializeSeed(ix, s, args, resolved)
		if err != nil {
			return nil, fmt.Errorf("seed[%d]: %w", i, err)
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *PdaResolver) materializeSeed(
	ix *Instruction,
	s Seed,
	args map[string]any,
	resolved map[string]solana.PublicKey,
) ([]byte, error) {
	switch s.Kind {
	case "const":
		return s.Value, nil
	case "account":
		pk, ok := resolved[s.Path]
		if !ok {
			return nil, fmt.Errorf("account seed %q references unresolved account: %w", s.Path, errAccountInputMissing)
		}
		return pk[:], nil
	case "arg":
		v, ok := args[s.Path]
		if !ok {
			return nil, fmt.Errorf("arg seed %q references missing arg: %w", s.Path, errAccountInputMissing)
		}
		seedType := s.Type
		if seedType == nil {
			inferred, err := findInstructionArgType(ix, s.Path)
			if err != nil {
				return nil, err
			}
			seedType = inferred
		}
		return seedBytesForType(*seedType, v)
	}
	return nil, fmt.Errorf("unknown seed kind %q", s.Kind)
}

func findInstructionArgType(ix *Instruction, name string) (*IdlType, error) {
	for i := range ix.Args {
		if ix.Args[i].Name == name {
			return &ix.Args[i].Type, nil
		}
	}
	return nil, fmt.Errorf("arg seed %q not found in instruction args", name)
}

func seedBytesForType(t IdlType, v any) ([]byte, error) {
	switch t.Kind {
	case KindBool:
		b, ok := v.(bool)
		if !ok {
			return nil, typeErr(t.Kind, v)
		}
		if b {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	case KindU8:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n > math.MaxUint8 {
			return nil, overflowErr(t.Kind, n)
		}
		return []byte{byte(n)}, nil
	case KindU16:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n > math.MaxUint16 {
			return nil, overflowErr(t.Kind, n)
		}
		out := make([]byte, 2)
		binary.LittleEndian.PutUint16(out, uint16(n))
		return out, nil
	case KindU32:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n > math.MaxUint32 {
			return nil, overflowErr(t.Kind, n)
		}
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(n))
		return out, nil
	case KindU64:
		n, err := asUint64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, n)
		return out, nil
	case KindU128:
		bi, ok := v.(*big.Int)
		if !ok {
			return nil, typeErr(t.Kind, v)
		}
		return uint128SeedBytes(bi)
	case KindI8:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n < math.MinInt8 || n > math.MaxInt8 {
			return nil, overflowErr(t.Kind, n)
		}
		return []byte{byte(int8(n))}, nil
	case KindI16:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n < math.MinInt16 || n > math.MaxInt16 {
			return nil, overflowErr(t.Kind, n)
		}
		out := make([]byte, 2)
		binary.LittleEndian.PutUint16(out, uint16(int16(n)))
		return out, nil
	case KindI32:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		if n < math.MinInt32 || n > math.MaxInt32 {
			return nil, overflowErr(t.Kind, n)
		}
		out := make([]byte, 4)
		binary.LittleEndian.PutUint32(out, uint32(int32(n)))
		return out, nil
	case KindI64:
		n, err := asInt64(v, t.Kind)
		if err != nil {
			return nil, err
		}
		out := make([]byte, 8)
		binary.LittleEndian.PutUint64(out, uint64(n))
		return out, nil
	case KindPubkey:
		pk, ok := v.(solana.PublicKey)
		if !ok {
			return nil, typeErr(t.Kind, v)
		}
		return pk[:], nil
	case KindString:
		s, ok := v.(string)
		if !ok {
			return nil, typeErr(t.Kind, v)
		}
		return []byte(s), nil
	case KindBytes:
		b, ok := v.([]byte)
		if !ok {
			return nil, typeErr(t.Kind, v)
		}
		return b, nil
	}
	return nil, fmt.Errorf("seed: unsupported IDL kind %q", t.Kind)
}

func uint128SeedBytes(bi *big.Int) ([]byte, error) {
	if bi == nil {
		return nil, fmt.Errorf("u128 seed: nil *big.Int")
	}
	if bi.Sign() < 0 {
		return nil, fmt.Errorf("u128 seed: negative value not allowed")
	}
	if bi.BitLen() > 128 {
		return nil, fmt.Errorf("u128 seed: value exceeds 128 bits")
	}
	be := bi.FillBytes(make([]byte, 16))
	// Reverse to little-endian.
	for i, j := 0, len(be)-1; i < j; i, j = i+1, j-1 {
		be[i], be[j] = be[j], be[i]
	}
	return be, nil
}
