package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const traceparentHeader = "traceparent"

// TraceHeaderMiddleware injects the traceparent header into HTTP responses.
func TraceHeaderMiddleware(c *gin.Context) {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(c.Request.Context(), carrier)
	c.Header(traceparentHeader, carrier.Get(traceparentHeader))
	c.Next()
}
