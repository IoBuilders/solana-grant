package servicecustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/domain/custodykey"
	"dlt-ingress/src/main/dltingress/domain/domainerrors"
	"dlt-ingress/src/main/dltingress/port/repository/custodykey"
)

type ExistMultiple struct {
	repository custodykeyrepo.Repository
}

type ExistMultipleInterface interface {
	Execute(ctx context.Context, dltAccountIds []string) ([]*custodykey.CustodyKey, error)
}

func NewExistMultiple(repository custodykeyrepo.Repository) *ExistMultiple {
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
