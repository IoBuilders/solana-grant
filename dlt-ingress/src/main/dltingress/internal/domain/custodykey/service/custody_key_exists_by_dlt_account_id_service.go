package servicecustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
)

type Exists struct {
	repository custodykey.Repository
}

type ExistsInterface interface {
	Execute(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error)
}

func NewExists(repository custodykey.Repository) *Exists {
	return &Exists{repository: repository}
}

func (e *Exists) Execute(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	custodyKey, err := e.repository.FindByDltAccountId(ctx, dltAccountId)
	if err != nil {
		return nil, domainerrors.NewCustodyKeyNotFoundByDltAccountIdDomainError(dltAccountId)
	}
	return custodyKey, nil
}
