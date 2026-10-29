package main

import (
	"os"
	"user-service/database"
	"user-service/handlers"
	"user-service/middleware"
	"user-service/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.Connect()
	database.DB.AutoMigrate(&models.User{})

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	protected := r.Group("/users")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/me", handlers.GetProfile)
		protected.PUT("/me", handlers.UpdateProfile)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	r.Run(":" + port)
}