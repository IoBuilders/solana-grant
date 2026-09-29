package api

import (
	"context"
	"reflect"

	"github.com/gin-gonic/gin"
	//nolint:staticcheck // legacy bridge while BCs migrate to coreerror
)

type BaseHttpAdapter[Rq any, Rs any] struct {
	handle func(ctx context.Context, request *HttpAdapterRequest[Rq]) (*HttpAdapterResponse[Rs], error)
}

func NewHttpAdapter[Rq any, Rs any](
	handle func(ctx context.Context, request *HttpAdapterRequest[Rq]) (*HttpAdapterResponse[Rs], error),
) *BaseHttpAdapter[Rq, Rs] {
	return &BaseHttpAdapter[Rq, Rs]{handle: handle}
}

func (a *BaseHttpAdapter[Rq, Rs]) Handler(c *gin.Context) {
	params := make(map[string]string)
	queryParams := make(map[string]string)
	for _, param := range c.Params {
		params[param.Key] = param.Value
	}
	for key, queryParam := range c.Request.URL.Query() {
		queryParams[key] = queryParam[0]
	}

	var req Rq
	if reflect.TypeOf(req) != nil {
		if err := c.ShouldBindJSON(&req); err != nil {
			RespondBindingError(c, err)
			return
		}
	}

	request := &HttpAdapterRequest[Rq]{
		c:    c,
		Body: &req,
	}

	resp, err := a.handle(c.Request.Context(), request)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(resp.StatusCode, resp.Response)
}
