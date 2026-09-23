package servicecustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/repository/custodykey"
)

type Exists struct {
	repository custodykeyrepo.Repository
}

type ExistsInterface interface {
	Execute(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error)
}

func NewExists(repository custodykeyrepo.Repository) *Exists {
	return &Exists{repository: repository}
}

func (e *Exists) Execute(ctx context.Context, dltAccountId string) (*custodykey.CustodyKey, error) {
	custodyKey, err := e.repository.FindByDltAccountId(ctx, dltAccountId)
	if err != nil {
		return nil, domainerrors.NewCustodyKeyNotFoundByDltAccountIdDomainError(dltAccountId)
	}
	return custodyKey, nil
}
