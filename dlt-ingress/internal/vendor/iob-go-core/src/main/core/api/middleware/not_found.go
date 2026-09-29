package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
)

func NotFound() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound,
			api.NewErrorResponse("Resource not found", api.ErrorCodeResourceNotFound, http.StatusNotFound),
		)
	}
}
