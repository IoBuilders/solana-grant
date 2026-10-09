package retryevmtransaction

import "github.com/gin-gonic/gin"

type TransactionRetrier interface {
	RetryTransaction(c *gin.Context)
}
