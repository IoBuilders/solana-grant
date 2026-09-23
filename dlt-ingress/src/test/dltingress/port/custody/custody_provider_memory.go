package custodymocks

import (
	"context"
	"crypto/ecdsa"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/custody"
	"dlt-ingress/src/main/dltingress/port/portcommon"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type InMemoryProvider struct {
	mu sync.RWMutex

	nextID int

	keys map[string]InMemoryKey // keyed by ExternalID
}

type InMemoryKey struct {
	ExternalID   string
	PrivateKey   *ecdsa.PrivateKey
	DltAccountID string
	Dlt          string
	KeyType      string
}

func NewInMemoryProvider() *InMemoryProvider {
	return &InMemoryProvider{
		nextID: 1,
		keys:   make(map[string]InMemoryKey),
	}
}

func (p *InMemoryProvider) CreateKey(ctx context.Context, request *custody.CreateKeyRequest) (*custody.KeyResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	privateKey, _ := crypto.GenerateKey()
	accountID := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	externalID := fmt.Sprintf("test-key-%d", p.nextID)
	p.nextID++

	key := InMemoryKey{
		ExternalID:   externalID,
		PrivateKey:   privateKey,
		Dlt:          request.Dlt,
		DltAccountID: accountID,
		KeyType:      request.KeyType,
	}

	p.keys[externalID] = key

	return &custody.KeyResponse{
		DltAccountId: accountID,
		ExternalId:   externalID,
	}, nil
}

func (p *InMemoryProvider) Sign(ctx context.Context, request *custody.SignRequest) (*custody.SignResponse, error) {
	p.mu.RLock()
	key, exists := p.keys[request.CustodyKeys[0].ExternalId]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key not found: %s", request.CustodyKeys[0].ExternalId)
	}

	if key.Dlt != request.Dlt {
		return nil, fmt.Errorf("key %s does not belong to dlt %s", key.ExternalID, request.Dlt)
	}

	tx, err := p.buildEthTransactionRequest(request.Transaction)
	if err != nil {
		return nil, fmt.Errorf("failed building transaction request: %w", err)
	}

	return p.signEthTransactionRequest(tx, request.Transaction.ChainId.RawValue(), key.PrivateKey)
}

func (p *InMemoryProvider) buildEthTransactionRequest(tx *portcommon.TransactionResponse) (*types.Transaction, error) {
	to := common.HexToAddress(tx.To)
	dataHex := strings.TrimPrefix(tx.Data, "0x")
	data, err := hex.DecodeString(dataHex)
	if err != nil {
		return nil, fmt.Errorf("failed decoding transaction data: %w", err)
	}

	switch tx.TransactionType {
	case portcommon.TransactionTypeDynamicFee:
		return types.NewTx(&types.DynamicFeeTx{
			ChainID:   tx.ChainId.RawValue(),
			Nonce:     tx.Nonce.RawValue().Uint64(),
			Gas:       tx.GasLimit.RawValue().Uint64(),
			GasTipCap: tx.MaxPriorityFeePerGas.RawValue(),
			GasFeeCap: tx.MaxFeePerGas.RawValue(),
			To:        &to,
			Value:     tx.Value.RawValue(),
			Data:      data,
		}), nil
	case portcommon.TransactionTypeLegacy:
		return types.NewTransaction(
			tx.Nonce.RawValue().Uint64(),
			to,
			tx.Value.RawValue(),
			tx.GasLimit.RawValue().Uint64(),
			tx.GasPrice.RawValue(),
			data,
		), nil
	}
	return nil, domainerrors.NewInvalidTransactionTypeDomainError(tx.TransactionType)
}

func (p *InMemoryProvider) signEthTransactionRequest(tx *types.Transaction, chainId *big.Int, privateKey *ecdsa.PrivateKey) (*custody.SignResponse, error) {
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainId), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed signing transaction: %w", err)
	}

	signedBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &custody.SignResponse{
		SignedTransaction: "0x" + hex.EncodeToString(signedBytes),
	}, nil
}

var _ custody.Port = (*InMemoryProvider)(nil)
