package main

import (
	"admin-service/handlers"
	"admin-service/middleware"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load()

	r := gin.Default()
	r.Use(middleware.Metrics())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	admin := r.Group("/admin")
	admin.Use(middleware.AdminOnly())
	{
		admin.POST("/flights", handlers.Forward("POST", "/internal/flights", false))
		admin.PATCH("/flights/:id", handlers.Forward("PATCH", "/internal/flights", true))
		admin.DELETE("/flights/:id", handlers.Forward("DELETE", "/internal/flights", true))

		admin.POST("/hotels", handlers.Forward("POST", "/internal/hotels", false))
		admin.PATCH("/hotels/:id", handlers.Forward("PATCH", "/internal/hotels", true))
		admin.DELETE("/hotels/:id", handlers.Forward("DELETE", "/internal/hotels", true))

		admin.GET("/analytics", handlers.Forward("GET", "/internal/analytics", false))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	r.Run(":" + port)
}
