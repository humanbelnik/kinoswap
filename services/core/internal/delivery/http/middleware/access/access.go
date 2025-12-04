package http_access_middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ReadOnlyBadGatewayMiddleware(mode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode != "RO" {
			c.Next()
			return
		}

		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "Bad Gateway",
			"message": "Write operations not allowed on read-only instance",
			"code":    "READ_ONLY_INSTANCE",
		})
		c.Abort()
	}
}
