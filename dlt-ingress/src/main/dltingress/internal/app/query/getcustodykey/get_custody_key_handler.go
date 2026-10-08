package getcustodykey

import (
	"context"
	"dlt-ingress/src/main/dltingress/internal/domain/custodykey"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

type QueryHandler struct {
	repo custodykey.Repository
}

func NewHandler(repo custodykey.Repository) *QueryHandler {
	return &QueryHandler{repo: repo}
}

func (h *QueryHandler) Execute(ctx context.Context, query Query) (query.Response, error) {

	key, err := h.repo.FindByDltAccountId(ctx, query.DltAccountId)
	if err != nil {
		return nil, err
	}

	return Response{
		KeyType:         string(key.KeyType),
		Status:          string(key.Status),
		DltAccountId:    key.DltAccountId,
		Dlt:             string(key.Dlt),
		ExternalId:      key.ExternalId,
		CustodyProvider: string(key.CustodyProvider),
	}, nil
}
