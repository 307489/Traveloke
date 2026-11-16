package main

import (
	"os"
	"user-service/database"
	"user-service/handlers"
	"user-service/middleware"
	"user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	godotenv.Load()
	database.Connect()
	database.DB.AutoMigrate(&models.User{})

	r := gin.Default()
	r.Use(middleware.Metrics())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	protected := r.Group("/users")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", handlers.GetProfile)
		protected.PATCH("/me", handlers.UpdateProfile)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":" + port)
}
