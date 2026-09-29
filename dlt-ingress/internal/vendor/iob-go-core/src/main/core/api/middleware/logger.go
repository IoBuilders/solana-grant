package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

// LoggerMiddleware logs responses including tracing information as default gin.Logger() middleware is not able to use context.Context.
// Should be configured after tracing middleware.
func LoggerMiddleware(skipPaths ...string) gin.HandlerFunc {
	skip := make(map[string]bool, len(skipPaths))
	for _, path := range skipPaths {
		skip[path] = true
	}

	return func(c *gin.Context) {
		if skip[c.Request.URL.Path] {
			c.Next()
			return
		}

		reqCtx := c.Request.Context()
		start := time.Now()

		c.Next()
		status := c.Writer.Status()
		var logMethod func(ctx context.Context, message string, args ...any)
		if status >= http.StatusInternalServerError {
			logMethod = logger.ErrorWithCtx
		} else if status >= http.StatusBadRequest {
			logMethod = logger.WarnWithCtx
		} else {
			logMethod = logger.InfoWithCtx
		}

		args := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}
		if status >= http.StatusBadRequest && len(c.Errors) > 0 {
			args = append(args, "error", c.Errors.Last().Err)
		}
		logMethod(
			context.WithoutCancel(reqCtx),
			fmt.Sprintf("%s %s %d", c.Request.Method, c.Request.URL.Path, c.Writer.Status()),
			args...
		)
	}
}
