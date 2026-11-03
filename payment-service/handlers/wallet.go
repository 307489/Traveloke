package handlers

import (
	"net/http"
	"payment-service/database"
	"payment-service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)


func GetWallet(c *gin.Context) {
	userID := c.GetString("user_id")

	var w models.Wallet
	if err := database.DB.Where("user_id = ?", userID).First(&w).Error; err != nil {
		c.JSON(http.StatusOK, models.Wallet{UserID: userID, Balance: 0})
		return
	}
	c.JSON(http.StatusOK, w)
}

func TopUp(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		Amount float64 `json:"amount" binding:"required,gt=0,lte=100000000"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0 and at most 100000000"})
		return
	}

	var w models.Wallet
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&models.Wallet{UserID: userID, Balance: 0}).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).First(&w).Error; err != nil {
			return err
		}
		w.Balance += input.Amount
		return tx.Save(&w).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Top up failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Top up successful", "wallet": w})
}