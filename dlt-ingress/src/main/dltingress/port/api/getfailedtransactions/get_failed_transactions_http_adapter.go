package getfailedtransactions

import (
	"context"
	"dlt-ingress/src/main/dltingress/app/query/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/port/api/model"
	"net/http"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

const UrlPath = "/admin/failures/transactions"

type Endpoint struct {
	queryBus query.Bus
}

// Handle godoc
//
// @Summary      List failed transactions
// @Description  Retrieves a paginated list of transactions that failed in the DLT.
// @Tags         Admin
// @ID           get-failed-transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageSize       query     int     false  "Maximum number of items to return"  default(20) example(20)
// @Param        offset         query     int     false  "Number of items to skip"            default(0)  example(0)
// @Param        sortBy         query     string  false  "Field used for sorting. Default: createdAt"  default(createdAt) example(createdAt)
// @Param        sortDirection  query     string  false  "Sort direction (asc or desc). Default: desc"  Enums(asc,desc) default(desc) example(desc)
// @Success      200  {object}  pagination.PageResponse{data=[]model.FailedTransactionResponse}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router       /admin/failures/transactions [get]
func NewEndpoint(queryBus query.Bus) *Endpoint {
	return &Endpoint{queryBus: queryBus}
}

func (e *Endpoint) Handle(ctx context.Context, req *api.HttpAdapterRequest[any]) (*api.HttpAdapterResponse[pagination.PageResponse], error) {
	params, err := req.GetPaginationParams()
	if err != nil {
		return nil, err
	}

	resp, err := query.Ask(ctx, e.queryBus, getfailedtransactions.Query{
		PaginationParams: params,
	})
	if err != nil {
		return nil, err
	}

	return api.NewHttpAdapterResponse(http.StatusOK, pagination.MapAndBuildResponse(resp, func(queryResponse *getfailedtransactions.FailedTransactionQueryResponse) *model.FailedTransactionResponse {
		return &model.FailedTransactionResponse{
			TxId:         queryResponse.TxId,
			NetworkId:    queryResponse.NetworkId,
			CreatedAt:    queryResponse.CreatedAt,
			ErrorDetails: queryResponse.ErrorDetails,
		}
	})), nil
}
