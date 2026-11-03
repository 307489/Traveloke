package main

import (
	"os"
	"payment-service/database"
	"payment-service/handlers"
	"payment-service/messaging"
	"payment-service/middleware"
	"payment-service/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.Connect()
	database.DB.AutoMigrate(&models.Payment{}, &models.Wallet{})
	messaging.Connect()
	messaging.StartConsumer()

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	payments := r.Group("/payments")
	payments.Use(middleware.AuthMiddleware())
	{
		payments.POST("", handlers.CreatePayment)
		payments.GET("/me", handlers.GetMyPayments)
	}

	walletRoutes := r.Group("/wallets")
	walletRoutes.Use(middleware.AuthMiddleware())
	{
		walletRoutes.GET("/me", handlers.GetWallet)
		walletRoutes.POST("/topup", handlers.TopUp)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	r.Run(":" + port)
}