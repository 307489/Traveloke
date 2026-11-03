package handlers

import (
	"booking-service/database"
	"booking-service/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InternalGetBooking(c *gin.Context) {
	var b models.Booking
	if err := database.DB.Where("id = ?", c.Param("id")).First(&b).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}
	c.JSON(http.StatusOK, b)
}