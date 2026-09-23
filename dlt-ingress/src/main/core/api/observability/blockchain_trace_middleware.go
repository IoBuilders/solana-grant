package observability

import (
	"bytes"
	"encoding/json"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/event/port/repository/eventstore"
)

type txHashBody struct {
	TransactionHash string `json:"transactionHash"`
}

func BlockchainTraceMiddleware(repo eventstorerepo.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("[BlockchainTraceMiddleware] failed to read request body: %v", err)
			c.Next()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var body txHashBody
		if err := json.Unmarshal(bodyBytes, &body); err != nil || body.TransactionHash == "" {
			c.Next()
			return
		}

		traceParent, err := repo.FindTraceParentByTxHash(c.Request.Context(), body.TransactionHash)
		if err != nil || traceParent == "" {
			c.Next()
			return
		}

		carrier := propagation.MapCarrier{"traceparent": traceParent}
		recoveredCtx := otel.GetTextMapPropagator().Extract(c.Request.Context(), carrier)
		c.Request = c.Request.WithContext(recoveredCtx)

		c.Next()
	}
}
