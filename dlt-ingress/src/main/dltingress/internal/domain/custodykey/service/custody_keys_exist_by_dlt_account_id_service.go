package servicecustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"
	"dlt-ingress/src/main/dltingress/internal/domain/domainerrors"
)

type ExistMultiple struct {
	repository custodykey.Repository
}

type ExistMultipleInterface interface {
	Execute(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error)
}

func NewExistMultiple(repository custodykey.Repository) *ExistMultiple {
	return &ExistMultiple{repository: repository}
}

func (e *ExistMultiple) Execute(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error) {
	custodyKeys, err := e.repository.FindByDltAccountIds(ctx, dltAccountIds)
	if err != nil {
		return nil, err
	}

	foundIds := make(map[string]struct{}, len(custodyKeys))
	for _, e := range custodyKeys {
		foundIds[e.DltAccountId] = struct{}{}
	}
	for _, id := range dltAccountIds {
		if _, ok := foundIds[id]; !ok {
			return nil, domainerrors.NewCustodyKeyNotFoundByDltAccountIdDomainError(id)
		}
	}

	return custodyKeys, nil
}
