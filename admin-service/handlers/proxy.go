package handlers

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var client = &http.Client{Timeout: 10 * time.Second}

func Forward(method, internalPath string, withID bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		target := os.Getenv("BOOKING_SERVICE_URL") + internalPath
		if withID {
			target += "/" + c.Param("id")
		}
		if q := c.Request.URL.RawQuery; q != "" {
			target += "?" + q
		}

		body, _ := io.ReadAll(c.Request.Body)
		req, err := http.NewRequest(method, target, bytes.NewReader(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build request"})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Internal-Key", os.Getenv("INTERNAL_API_KEY"))

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Booking service unavailable"})
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", respBody)
	}
}