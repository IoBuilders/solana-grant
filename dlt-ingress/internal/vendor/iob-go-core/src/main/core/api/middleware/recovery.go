package middleware

import (
	"bytes"
	"fmt"

	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/error/panicinfo"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
)

// RecoveryMiddleware logs api panic errors including tracing information as default gin.Recovery() middleware is not able to use context.Context.
// Should be configured after tracing middleware.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var buf bytes.Buffer

		gin.CustomRecoveryWithWriter(&buf, func(c *gin.Context, recovered any) {
			info := panicinfo.NewPanicInfo(recovered)
			msg := buf.String()
			if msg != "" {
				logger.ErrorWithCtx(
					c.Request.Context(),
					fmt.Sprintf("%s %s", panicinfo.PANIC_RECOVERED_TAG, msg),
					"file", info.File,
					"line", info.Line,
					"message", info.Message,
					"http.method", c.Request.Method,
					"http.path", c.Request.URL.Path,
					"http.query", c.Request.URL.RawQuery,
				)
			}
			Internal(c)
		})(c)
	}
}
