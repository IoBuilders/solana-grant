package custody

import (
	"fmt"
	"strings"
)

type ProviderName string

const (
	Dfns ProviderName = "dfns"
	KMS  ProviderName = "kms"
)

type Config struct {
	Provider string
	Dfns     DfnsConfig
	Kms      KMSConfig
}

func NewCustodyProvider(cfg Config) (Port, error) {
	switch ProviderName(strings.ToLower(cfg.Provider)) {
	case Dfns:
		return newDfnsAdapter(cfg.Dfns)
	case KMS:
		return newKMSAdapter(cfg.Kms)
	default:
		return nil, fmt.Errorf("unsupported custody provider: %s", cfg.Provider)
	}
}

func newDfnsAdapter(cfg DfnsConfig) (Port, error) {
	client, err := cfg.NewDfnsClient()
	if err != nil {
		return nil, err
	}
	return NewDfnsProvider(client), nil
}
