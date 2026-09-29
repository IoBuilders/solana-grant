package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/health"
)

func HealthStatus(registry *health.Registry) func(c *gin.Context) {
	return func(c *gin.Context) {
		res := registry.RunAllChecks(c.Request.Context())
		var statusCode int
		if res.Status == health.StatusUp {
			statusCode = http.StatusOK
		} else {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, res)
	}
}
