package dltingressconfig

import (
	"dlt-ingress/src/main/dltingress/internal/infra/api/createfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/internal/infra/api/getfaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/api/retryevmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/api/retrysvmtransaction"
	"dlt-ingress/src/main/dltingress/internal/infra/api/updatefaucetwallet"
	"dlt-ingress/src/main/dltingress/internal/infra/blockchain/refundaccount"
	"dlt-ingress/src/main/dltingress/internal/infra/blockchain/savefailedtransaction"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/getfailedevents"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/retryfailedevent"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func RegisterEndpoints(l *LazyDltIngress, router *gin.Engine, authTokenInterceptor gin.HandlerFunc) {
	adminRoutes := router.Group("api/v1")
	adminRoutes.Use(authTokenInterceptor)

	adminRoutes.POST(retryevmtransaction.UrlPath, retryevmtransaction.NewEndpoint(l.CommandBus).GinHandler())
	adminRoutes.POST(retrysvmtransaction.UrlPath, retrysvmtransaction.NewEndpoint(l.CommandBus).GinHandler())
	adminRoutes.POST(createfaucetwallet.UrlPath, createfaucetwallet.NewEndpoint(l.CommandBus).GinHandler())
	adminRoutes.GET(getfaucetwallet.UrlPath, getfaucetwallet.NewEndpoint(l.core.QueryBus).GinHandler())
	adminRoutes.PUT(updatefaucetwallet.UrlPath, updatefaucetwallet.NewEndpoint(l.CommandBus).GinHandler())
	adminRoutes.GET(getfailedtransactions.UrlPath, getfailedtransactions.NewEndpoint(l.core.QueryBus).GinHandler())
	adminRoutes.GET(getfailedevents.UrlPath("dltingress"), FailedEvents(l.DltIngressQueryBus))
	adminRoutes.POST(retryfailedevent.UrlPath("dltingress"), RetryFailedEvent(l.CommandBus))

	// DLT Events
	internalRoutes := router.Group("api/v1")
	internalRoutes.POST(savefailedtransaction.UrlPath, savefailedtransaction.NewEndpoint(l.CommandBus).GinHandler())
	internalRoutes.POST(refundaccount.UrlPath, refundaccount.NewEndpoint(l.AppServices.FundAccountAppService).GinHandler())
}

// Handle godoc
//
// @Summary      List failed events
// @Description  Retrieves a paginated list of events with FAILED status.
// @Tags         Admin
// @ID           get-failed-events-dltingress
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        pageSize       query     int     false  "Maximum number of items to return"  default(20) example(20)
// @Param        offset         query     int     false  "Number of items to skip"            default(0)  example(0)
// @Param        sortBy         query     string  false  "Field used for sorting. Default: createdAt"  default(createdAt) example(createdAt)
// @Param        sortDirection  query     string  false  "Sort direction (asc or desc). Default: desc"  Enums(asc,desc) default(desc) example(desc)
// @Success      200  {object}  pagination.PageResponse{data=[]model.EventConsumerResponse}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Failure      500  {object}  api.ErrorResponse
// @Router		/admin/dltingress/failures/events [get]
func FailedEvents(queryBus query.Bus) func(c *gin.Context) {
	return api.NewHttpAdapter(
		getfailedevents.NewEndpoint(queryBus).Handle,
	).Handler
}

// Handle godoc
//
// @Summary      Retry a failed event
// @Description  Retries a failed event that has been failed for an specific consumer.
// @Tags         Admin
// @ID           retry-failed-event-dltingress
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        eventId  		 path      string  true  "Event Id"
// @Success      200             {object}  api.EmptyResponse
// @Failure      400             {object}  api.ErrorResponse
// @Failure      404             {object}  api.ErrorResponse
// @Failure      500             {object}  api.ErrorResponse
// @Router       /admin/dltingress/failures/events/{eventId}/retry [post]
func RetryFailedEvent(commandBus command.Bus) func(c *gin.Context) {
	return api.NewHttpAdapter(
		retryfailedevent.NewEndpoint(commandBus).Handle,
	).Handler
}
