package dfnsprovider

type BlockchainKind string
type SignatureKind string

const (
	BlockchainKindEvm    BlockchainKind = "Evm"
	BlockchainKindSolana BlockchainKind = "Solana"
)

const (
	SignatureKindHash        SignatureKind = "Hash"
	SignatureKindTransaction SignatureKind = "Transaction"
)

type EvmSignatureRequest struct {
	BlockchainKind BlockchainKind `json:"blockchainKind"`
	Kind           SignatureKind  `json:"kind"`
	Hash           string         `json:"hash"`
}

func NewEvmSignatureRequest(hash string) *EvmSignatureRequest {
	return &EvmSignatureRequest{
		BlockchainKind: BlockchainKindEvm,
		Kind:           SignatureKindHash,
		Hash:           hash,
	}
}

type SvmSignatureRequest struct {
	BlockchainKind BlockchainKind `json:"blockchainKind"`
	Kind           SignatureKind  `json:"kind"`
	Transaction    string         `json:"transaction"`
}

func NewSvmSignatureRequest(serializedTransaction string) *SvmSignatureRequest {
	return &SvmSignatureRequest{
		BlockchainKind: BlockchainKindSolana,
		Kind:           SignatureKindTransaction,
		Transaction:    serializedTransaction,
	}
}
