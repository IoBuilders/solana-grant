package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api"
)

func Internal(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusInternalServerError,
		api.ApiErrorResponseInternal,
	)
}
