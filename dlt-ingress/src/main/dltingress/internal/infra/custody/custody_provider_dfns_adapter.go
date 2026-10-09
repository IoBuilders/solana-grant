package custody

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"dlt-ingress/src/main/config"
	"dlt-ingress/src/main/dltingress/internal/domain/common"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/infra/custody/dfns"
	"dlt-ingress/src/main/dltingress/internal/infra/noderatelimit"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/ratelimit"

	"github.com/dfns/dfns-sdk-go/v2"
	"github.com/dfns/dfns-sdk-go/v2/keys"
	"github.com/dfns/dfns-sdk-go/v2/signer"
	"github.com/dfns/dfns-sdk-go/v2/types"
	"github.com/dfns/dfns-sdk-go/v2/wallets"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type DfnsConfig struct {
	BaseUrl      string
	AuthToken    string
	CredentialID string
	PrivateKey   string
}

const (
	dfnsMethodCreateWallet      = "createWallet"
	dfnsMethodGenerateSignature = "generateSignature"
	dfnsMethodGetSignature      = "getSignature"
)

func newDfnsHTTPClient(requestTimeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   requestTimeout,
		Transport: noderatelimit.NewTransport(http.DefaultTransport),
	}
}

func (c DfnsConfig) NewDfnsClient(requestTimeout time.Duration) (*dfns.Client, error) {
	if c.BaseUrl == "" {
		return nil, fmt.Errorf("missing dfns base url")
	}
	if c.AuthToken == "" {
		return nil, fmt.Errorf("missing dfns auth token")
	}
	if c.CredentialID == "" {
		return nil, fmt.Errorf("missing dfns credential id")
	}
	if c.PrivateKey == "" {
		return nil, fmt.Errorf("missing dfns private key")
	}

	key := strings.ReplaceAll(c.PrivateKey, `\n`, "\n")

	keySigner, err := signer.NewKeySigner(c.CredentialID, key)
	if err != nil {
		return nil, fmt.Errorf("create dfns signer: %w", err)
	}

	client, err := dfns.NewClient(dfns.Options{
		BaseURL:    c.BaseUrl,
		AuthToken:  c.AuthToken,
		Signer:     keySigner,
		HTTPClient: newDfnsHTTPClient(requestTimeout),
	})
	if err != nil {
		return nil, fmt.Errorf("create dfns client: %w", err)
	}

	return client, nil
}

type DfnsProvider struct {
	client *dfns.Client
	guard  *noderatelimit.Guard
}

func NewDfnsProvider(client *dfns.Client, guard *noderatelimit.Guard) *DfnsProvider {
	return &DfnsProvider{
		client: client,
		guard:  guard,
	}
}

func (d *DfnsProvider) CreateKey(ctx context.Context, request *CreateKeyRequest) (*KeyResponse, error) {
	network, err := mapToDfnsNetwork(request.Dlt)
	if err != nil {
		return nil, fmt.Errorf("resolve dfns network: %w", err)
	}

	wallet, err := noderatelimit.Call(ctx, d.guard, dfnsMethodCreateWallet, func() (*wallets.CreateWalletResponse, error) {
		return d.client.Wallets.CreateWallet(ctx, wallets.CreateWalletRequest{
			Network: string(network),
		})
	})
	if err != nil {
		var apiErr *dfns.APIError
		if errors.As(err, &apiErr) {
			return nil, fmt.Errorf("dfns create wallet failed (status=%d): %w", apiErr.StatusCode, err)
		}
		return nil, fmt.Errorf("dfns create wallet failed: %w", err)
	}

	signingKeyId, ok := wallet.SigningKey["id"]
	if !ok {
		return nil, fmt.Errorf("signingKey.id not found")
	}

	resp := &KeyResponse{
		ExternalId: signingKeyId.(string),
	}

	if wallet.Address != nil {
		resp.DltAccountId = *wallet.Address
	}

	return resp, nil
}

func (d *DfnsProvider) Sign(ctx context.Context, request *SignRequest) (*SignResponse, error) {
	var signedTransaction, txId string
	var err error

	switch request.Dlt {
	case string(common.EVM):
		signedTransaction, txId, err = d.signEvmTransaction(ctx, request)
	case string(common.SVM):
		signedTransaction, txId, err = d.signSvmTransaction(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported dlt: %s", request.Dlt)
	}

	if err != nil {
		return nil, err
	}

	return &SignResponse{SignedTransaction: signedTransaction, TxId: txId}, nil
}

func mapToDfnsNetwork(dlt string) (types.Network, error) {
	switch dlt {
	case string(common.EVM):
		return types.NetworkEthereum, nil
	case string(common.SVM):
		return types.NetworkSolana, nil
	case string(common.Hashgraph):
		return types.NetworkHedera, nil
	default:
		return "", fmt.Errorf("unsupported dlt/network: %s", dlt)
	}
}

func (d *DfnsProvider) signEvmTransaction(ctx context.Context, req *SignRequest) (string, string, error) {
	if len(req.CustodyKeys) != 1 {
		return "", "", fmt.Errorf("one custody key is required for EVM signing")
	}
	externalId := req.CustodyKeys[0].ExternalId

	tx, txSigner, hash, err := buildEVMUnsignedTx(req.Transaction)
	if err != nil {
		return "", "", err
	}

	resp, err := noderatelimit.Call(ctx, d.guard, dfnsMethodGenerateSignature, func() (*keys.GenerateSignatureResponse, error) {
		return d.client.Keys.GenerateSignature(ctx, externalId, dfnsprovider.NewEvmSignatureRequest(hash.Hex()))
	})
	signatureMap, err := d.handleSignatureResponse(ctx, resp, err)
	if err != nil {
		return "", "", err
	}

	rBytes, err := hexutil.Decode(signatureMap["r"].(string))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode signature R: %w", err)
	}
	sBytes, err := hexutil.Decode(signatureMap["s"].(string))
	if err != nil {
		return "", "", fmt.Errorf("failed to decode signature S: %w", err)
	}

	return encodeSignedEVMTx(tx, txSigner, assembleECDSASig(rBytes, sBytes, int(signatureMap["recid"].(float64))))
}

func (d *DfnsProvider) signSvmTransaction(ctx context.Context, req *SignRequest) (string, string, error) {
	if len(req.CustodyKeys) == 0 {
		return "", "", fmt.Errorf("at least one custody key is required for SVM signing")
	}
	base64Tx := req.Transaction.SVMTransactionResponse.SerializedTransaction
	signatures, err := d.generateSvmSignatures(ctx, req.CustodyKeys, base64Tx)
	if err != nil {
		return "", "", err
	}
	return InjectSvmSignatures(base64Tx, signatures)
}

func (d *DfnsProvider) generateSvmSignatures(ctx context.Context, custodyKeys []*custodykey.CustodyKey, base64Tx string) (map[string][]byte, error) {
	hexEncodedTx, err := TransformBase64SvmTransactionIntoHex(base64Tx)
	if err != nil {
		return nil, err
	}

	type sigResult struct {
		pubkey string
		sig    []byte
		err    error
	}
	sigResults := make([]sigResult, len(custodyKeys))
	var wg sync.WaitGroup
	for i, custodyKey := range custodyKeys {
		wg.Add(1)
		go func(idx int, externalId, dltAccountId string) {
			defer wg.Done()
			sig, err := d.generateSvmSignature(ctx, externalId, hexEncodedTx)
			sigResults[idx] = sigResult{pubkey: dltAccountId, sig: sig, err: err}
		}(i, custodyKey.ExternalId, custodyKey.DltAccountId)
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

func (d *DfnsProvider) generateSvmSignature(ctx context.Context, externalId string, serializedTx string) ([]byte, error) {
	resp, err := noderatelimit.Call(ctx, d.guard, dfnsMethodGenerateSignature, func() (*keys.GenerateSignatureResponse, error) {
		return d.client.Keys.GenerateSignature(ctx, externalId, dfnsprovider.NewSvmSignatureRequest(serializedTx))
	})
	signatureMap, err := d.handleSignatureResponse(ctx, resp, err)
	if err != nil {
		return nil, err
	}

	signature := signatureMap["encoded"].(string)
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return nil, err
	}

	return sigBytes, nil
}

func (d *DfnsProvider) handleSignatureResponse(ctx context.Context, sigResponse *keys.GenerateSignatureResponse, err error) (map[string]interface{}, error) {
	if err != nil {
		return nil, handleApiError("dfns generate signature failed", err)
	}

	if sigResponse.Status == "Pending" {
		return d.pendingSignaturePoll(ctx, sigResponse.KeyID, sigResponse.ID)
	}

	if sigResponse.Status != "Signed" {
		return nil, fmt.Errorf("dfns generate signature failed: %s", *sigResponse.Reason)
	}
	return *sigResponse.Signature, nil
}

func handleApiError(msg string, err error) error {
	var apiErr *dfns.APIError
	if errors.As(err, &apiErr) {
		return fmt.Errorf(msg+" (status=%d): %w", apiErr.StatusCode, err)
	}
	return fmt.Errorf(msg+": %w", err)
}

func (d *DfnsProvider) pendingSignaturePoll(ctx context.Context, keyId string, signatureId string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DltIngressConfig.DltIngress.Custody.SignPollingTimeout)
	defer cancel()

	ticker := time.NewTicker(config.DltIngressConfig.DltIngress.Custody.SignPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("dfns async signature poll failed: %w", ctx.Err())
		case <-ticker.C:
			resp, err := noderatelimit.Call(ctx, d.guard, dfnsMethodGetSignature, func() (*keys.GetSignatureResponse, error) {
				return d.client.Keys.GetSignature(ctx, keyId, signatureId)
			})
			if ratelimit.IsExceededError(err) {
				continue
			}
			if err != nil {
				return nil, handleApiError("dfns async signature poll failed", err)
			}
			if resp.Status == "Pending" {
				continue
			}
			if resp.Status != "Signed" {
				return nil, fmt.Errorf("dfns async signature poll failed: %s", *resp.Reason)
			}
			return *resp.Signature, nil
		}
	}
}

var _ Port = (*DfnsProvider)(nil)
