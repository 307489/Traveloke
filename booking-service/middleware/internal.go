package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func InternalOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := os.Getenv("INTERNAL_API_KEY")
		if key == "" || c.GetHeader("X-Internal-Key") != key {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}