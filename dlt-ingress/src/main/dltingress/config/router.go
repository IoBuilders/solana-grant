package dltingressconfig

import (
	"dlt-ingress/src/main/core/api/auth"
	"dlt-ingress/src/main/dltingress/port/api/getfailedtransactions"
	"dlt-ingress/src/main/dltingress/port/api/retrytransaction"

	"github.com/gin-gonic/gin"
	coreApi "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/command"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/getfailedevents"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/api/retryfailedevent"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/query"
)

func RegisterEndpoints(l *LazyDltIngress, router *gin.Engine) {
	adminRoutes := router.Group("api/v1")
	adminRoutes.Use(auth.TokenInterceptor())

	adminRoutes.POST(retrytransaction.UrlPath, retrytransaction.NewEndpoint(l.CommandBus).GinHandler())
	adminRoutes.GET(getfailedtransactions.UrlPath, getfailedtransactions.NewEndpoint(l.core.QueryBus).GinHandler())
	adminRoutes.GET(getfailedevents.UrlPath("dltingress"), FailedEvents(l.DltIngressQueryBus))
	adminRoutes.POST(retryfailedevent.UrlPath("dltingress"), RetryFailedEvent(l.CommandBus))
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
// @Failure      400  {object}  coreApi.ErrorResponse
// @Failure      401  {object}  coreApi.ErrorResponse
// @Failure      500  {object}  coreApi.ErrorResponse
// @Router		/admin/dltingress/failures/events [get]
func FailedEvents(queryBus query.Bus) func(c *gin.Context) {
	return coreApi.NewHttpAdapter(
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
// @Success      200             {object}  coreApi.EmptyResponse
// @Failure      400             {object}  coreApi.ErrorResponse
// @Failure      404             {object}  coreApi.ErrorResponse
// @Failure      500             {object}  coreApi.ErrorResponse
// @Router       /admin/dltingress/failures/events/{eventId}/retry [post]
func RetryFailedEvent(commandBus command.Bus) func(c *gin.Context) {
	return coreApi.NewHttpAdapter(
		retryfailedevent.NewEndpoint(commandBus).Handle,
	).Handler
}
