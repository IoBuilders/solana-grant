package custody

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/smithy-go"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"

	"dlt-ingress/src/main/dltingress/internal/domain/common"
)

type KMSTag struct {
	Key   string
	Value string
}

type KMSConfig struct {
	Region      string
	AccessKey   string
	SecretKey   string
	Endpoint    string
	Tags        []KMSTag
	AliasPrefix string
}

const (
	kmsMethodCreateKey    = "CreateKey"
	kmsMethodCreateAlias  = "CreateAlias"
	kmsMethodGetPublicKey = "GetPublicKey"
	kmsMethodSign         = "Sign"
)

const kmsThrottlingErrorCode = "ThrottlingException"

type KmsAdapter struct {
	client      *kms.Client
	tags        []KMSTag
	aliasPrefix string
	guard       *noderatelimit.Guard
	callTimeout time.Duration
}

// newKMSAdapter bounds each KMS call, SDK retries included, by callTimeout, so a hung endpoint fails inside the
// command's DB transaction instead of outliving it.
func newKMSAdapter(cfg KMSConfig, guard *noderatelimit.Guard, callTimeout time.Duration) (*KmsAdapter, error) {
	if cfg.Region == "" {
		return nil, fmt.Errorf("missing KMS region")
	}

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.Region),
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	var kmsOpts []func(*kms.Options)
	if cfg.Endpoint != "" {
		endpoint := cfg.Endpoint
		kmsOpts = append(kmsOpts, func(o *kms.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &KmsAdapter{
		client:      kms.NewFromConfig(awsCfg, kmsOpts...),
		tags:        cfg.Tags,
		aliasPrefix: cfg.AliasPrefix,
		guard:       guard,
		callTimeout: callTimeout,
	}, nil
}

func callKMS[T any](ctx context.Context, k *KmsAdapter, method string, call func(ctx context.Context) (T, error)) (T, error) {
	return noderatelimit.Call(ctx, k.guard, method, func() (T, error) {
		callCtx, cancel := context.WithTimeout(ctx, k.callTimeout)
		defer cancel()
		result, err := call(callCtx)
		if apiErr, ok := errors.AsType[smithy.APIError](err); ok && apiErr.ErrorCode() == kmsThrottlingErrorCode {
			return result, fmt.Errorf("%w: %w", &noderatelimit.TooManyRequestsError{}, err)
		}
		return result, err
	})
}

func (k *KmsAdapter) CreateKey(ctx context.Context, request *CreateKeyRequest) (*KeyResponse, error) {
	keySpec, err := mapToKMSKeySpec(request.KeyType)
	if err != nil {
		return nil, err
	}

	kmsTags := make([]types.Tag, len(k.tags))
	for i, t := range k.tags {
		kmsTags[i] = types.Tag{TagKey: aws.String(t.Key), TagValue: aws.String(t.Value)}
	}

	createOut, err := callKMS(ctx, k, kmsMethodCreateKey, func(ctx context.Context) (*kms.CreateKeyOutput, error) {
		return k.client.CreateKey(ctx, &kms.CreateKeyInput{
			KeySpec:  keySpec,
			KeyUsage: types.KeyUsageTypeSignVerify,
			Tags:     kmsTags,
		})
	})
	if err != nil {
		logger.ErrorWithCtx(ctx, "Failed to create KMS key", "error", err)
		return nil, fmt.Errorf("KMS create key: %w", err)
	}

	keyId := aws.ToString(createOut.KeyMetadata.KeyId)

	if k.aliasPrefix != "" {
		_, err = callKMS(ctx, k, kmsMethodCreateAlias, func(ctx context.Context) (*kms.CreateAliasOutput, error) {
			return k.client.CreateAlias(ctx, &kms.CreateAliasInput{
				TargetKeyId: aws.String(keyId),
				AliasName:   aws.String(fmt.Sprintf("%s/%s", k.aliasPrefix, keyId)),
			})
		})
		if err != nil {
			logger.ErrorWithCtx(ctx, "Failed to create KMS alias", "error", err)
			return nil, err
		}
	}

	pubKeyOut, err := callKMS(ctx, k, kmsMethodGetPublicKey, func(ctx context.Context) (*kms.GetPublicKeyOutput, error) {
		return k.client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: aws.String(keyId)})
	})
	if err != nil {
		logger.ErrorWithCtx(ctx, "Failed to get KMS key", "error", err)
		return nil, err
	}

	address, err := deriveAccountId(request.Dlt, pubKeyOut.PublicKey)
	if err != nil {
		logger.ErrorWithCtx(ctx, "Failed to derive KMS key address", "error", err)
		return nil, fmt.Errorf("derive account id: %w", err)
	}

	return &KeyResponse{
		DltAccountId: address,
		ExternalId:   keyId,
	}, nil
}

func mapToKMSKeySpec(keyType string) (types.KeySpec, error) {
	switch common.KeyType(keyType) {
	case common.ECDSASecp256k1:
		return types.KeySpecEccSecgP256k1, nil
	case common.ED25519:
		return types.KeySpecEccNistEdwards25519, nil
	default:
		return "", fmt.Errorf("unsupported key type for KMS provider: %s", keyType)
	}
}

func deriveAccountId(dlt string, pubKeyDER []byte) (string, error) {
	switch common.Dlt(dlt) {
	case common.EVM, common.Hashgraph:
		return deriveEthAddress(pubKeyDER)
	case common.SVM:
		return deriveSolanaAddress(pubKeyDER)
	default:
		return "", fmt.Errorf("unsupported dlt for KMS account id derivation: %s", dlt)
	}
}

func (k *KmsAdapter) Sign(ctx context.Context, request *SignRequest) (*SignResponse, error) {
	var signedTransaction, txId string
	var err error

	switch request.Dlt {
	case string(common.EVM):
		signedTransaction, txId, err = k.signEvmTransaction(ctx, request)
	case string(common.SVM):
		signedTransaction, txId, err = k.signSvmTransaction(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported dlt: %s", request.Dlt)
	}

	if err != nil {
		return nil, err
	}

	return &SignResponse{SignedTransaction: signedTransaction, TxId: txId}, nil
}

func (k *KmsAdapter) signEvmTransaction(ctx context.Context, req *SignRequest) (string, string, error) {
	if len(req.CustodyKeys) != 1 {
		return "", "", fmt.Errorf("one custody key is required for EVM signing")
	}
	custodyKey := req.CustodyKeys[0]

	tx, signer, hash, err := buildEVMUnsignedTx(req.Transaction)
	if err != nil {
		return "", "", err
	}

	signOut, err := callKMS(ctx, k, kmsMethodSign, func(ctx context.Context) (*kms.SignOutput, error) {
		return k.client.Sign(ctx, &kms.SignInput{
			KeyId:            aws.String(custodyKey.ExternalId),
			Message:          hash.Bytes(),
			MessageType:      types.MessageTypeDigest,
			SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
		})
	})
	if err != nil {
		return "", "", fmt.Errorf("KMS sign: %w", err)
	}

	r, s, err := decodeDERSignature(signOut.Signature)
	if err != nil {
		return "", "", fmt.Errorf("decode DER signature: %w", err)
	}

	// EIP-2: Evm nodes reject high-S signatures. KMS does not guarantee low-S,
	// so normalize: if s > n/2, replace s with n-s (the valid low-S equivalent).
	secp256k1N := crypto.S256().Params().N
	halfN := new(big.Int).Rsh(secp256k1N, 1)
	if s.Cmp(halfN) > 0 {
		s = new(big.Int).Sub(secp256k1N, s)
	}

	recid, err := findRecoveryBit(hash.Bytes(), r, s, custodyKey.DltAccountId)
	if err != nil {
		return "", "", err
	}

	return encodeSignedEVMTx(tx, signer, assembleECDSASig(r.Bytes(), s.Bytes(), recid))
}

type asn1ECDSASig struct {
	R, S *big.Int
}

func decodeDERSignature(der []byte) (*big.Int, *big.Int, error) {
	var sig asn1ECDSASig
	if _, err := asn1.Unmarshal(der, &sig); err != nil {
		return nil, nil, fmt.Errorf("unmarshal DER signature: %w", err)
	}
	return sig.R, sig.S, nil
}

func deriveEthAddress(pubKeyDER []byte) (string, error) {
	// Go's x509 doesn't support secp256k1 (non-NIST curve), so parse the
	// SubjectPublicKeyInfo structure manually and let go-ethereum handle the key.
	var spki struct {
		Algorithm struct {
			Algorithm  asn1.ObjectIdentifier
			Parameters asn1.RawValue `asn1:"optional"`
		}
		PublicKey asn1.BitString
	}
	if _, err := asn1.Unmarshal(pubKeyDER, &spki); err != nil {
		return "", fmt.Errorf("parse DER public key: %w", err)
	}
	pubKey, err := crypto.UnmarshalPubkey(spki.PublicKey.Bytes)
	if err != nil {
		return "", fmt.Errorf("unmarshal secp256k1 public key: %w", err)
	}
	return crypto.PubkeyToAddress(*pubKey).Hex(), nil
}

func deriveSolanaAddress(pubKeyDER []byte) (string, error) {
	pubKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyDER)
	if err != nil {
		return "", fmt.Errorf("parse Ed25519 public key: %w", err)
	}
	ed25519Key, ok := pubKeyInterface.(ed25519.PublicKey)
	if !ok {
		return "", fmt.Errorf("unexpected key type for SVM: %T", pubKeyInterface)
	}
	pk := solana.PublicKeyFromBytes(ed25519Key)
	return pk.String(), nil
}

func (k *KmsAdapter) signSvmTransaction(ctx context.Context, req *SignRequest) (string, string, error) {
	if len(req.CustodyKeys) == 0 {
		return "", "", fmt.Errorf("at least one custody key is required for SVM signing")
	}

	serializedTx := req.Transaction.SVMTransactionResponse.SerializedTransaction
	signatures, err := k.generateSvmSignatures(ctx, req.CustodyKeys, serializedTx)
	if err != nil {
		return "", "", err
	}

	return InjectSvmSignatures(serializedTx, signatures)
}

func (k *KmsAdapter) generateSvmSignatures(ctx context.Context, custodyKeys []*custodykey.CustodyKey, serializedTx string) (map[string][]byte, error) {
	tx, err := solana.TransactionFromBase64(serializedTx)
	if err != nil {
		return nil, fmt.Errorf("deserialize SVM transaction: %w", err)
	}

	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("serialize SVM transaction message: %w", err)
	}

	type sigResult struct {
		pubkey string
		sig    []byte
		err    error
	}
	sigResults := make([]sigResult, len(custodyKeys))
	var wg sync.WaitGroup
	for i, key := range custodyKeys {
		wg.Add(1)
		go func(idx int, externalId, dltAccountId string) {
			defer wg.Done()
			sig, err := k.kmsSignEd25519(ctx, externalId, messageBytes)
			sigResults[idx] = sigResult{pubkey: dltAccountId, sig: sig, err: err}
		}(i, key.ExternalId, key.DltAccountId)
	}
	wg.Wait()

	signatures := make(map[string][]byte, len(custodyKeys))
	for _, r := range sigResults {
		if r.err != nil {
			return nil, r.err
		}
		signatures[r.pubkey] = r.sig
	}

	return signatures, nil
}

func (k *KmsAdapter) kmsSignEd25519(ctx context.Context, externalId string, messageBytes []byte) ([]byte, error) {
	signOut, err := callKMS(ctx, k, kmsMethodSign, func(ctx context.Context) (*kms.SignOutput, error) {
		return k.client.Sign(ctx, &kms.SignInput{
			KeyId:            aws.String(externalId),
			Message:          messageBytes,
			MessageType:      types.MessageTypeRaw,
			SigningAlgorithm: types.SigningAlgorithmSpecEd25519Sha512,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("KMS sign: %w", err)
	}
	return signOut.Signature, nil
}

func findRecoveryBit(hash []byte, r, s *big.Int, expectedAddr string) (int, error) {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	for recid := 0; recid <= 1; recid++ {
		sigBytes := make([]byte, 65)
		copy(sigBytes[32-len(rBytes):32], rBytes)
		copy(sigBytes[64-len(sBytes):64], sBytes)
		sigBytes[64] = byte(recid)

		pubKeyBytes, err := crypto.Ecrecover(hash, sigBytes)
		if err != nil {
			continue
		}
		pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
		if err != nil {
			continue
		}
		if strings.EqualFold(crypto.PubkeyToAddress(*pubKey).Hex(), expectedAddr) {
			return recid, nil
		}
	}
	return 0, fmt.Errorf("could not determine recovery bit for address %s", expectedAddr)
}

var _ Port = (*KmsAdapter)(nil)
