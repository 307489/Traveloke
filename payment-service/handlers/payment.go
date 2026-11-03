package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"payment-service/database"
	"payment-service/messaging"
	"payment-service/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type bookingInfo struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TotalPrice  float64   `json:"total_price"`
	Status      string    `json:"status"`
	HoldExpires time.Time `json:"hold_expires_at"`
}

func fetchBooking(id string) (*bookingInfo, int) {
	req, _ := http.NewRequest("GET", os.Getenv("BOOKING_SERVICE_URL")+"/internal/bookings/"+id, nil)
	req.Header.Set("X-Internal-Key", os.Getenv("INTERNAL_API_KEY"))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode
	}
	var b bookingInfo
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, http.StatusInternalServerError
	}
	return &b, http.StatusOK
}

func CreatePayment(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		BookingID string `json:"booking_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id is required"})
		return
	}

	b, code := fetchBooking(input.BookingID)
	if b == nil {
		switch code {
		case http.StatusNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		case http.StatusServiceUnavailable:
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Booking service unavailable"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to validate booking"})
		}
		return
	}

	if b.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not your booking"})
		return
	}
	if b.Status != "held" {
		c.JSON(http.StatusConflict, gin.H{"error": "Booking is not payable (status: " + b.Status + ")"})
		return
	}
	if time.Now().After(b.HoldExpires) {
		c.JSON(http.StatusConflict, gin.H{"error": "Booking hold has expired"})
		return
	}

	// cegah double payment yang masih berjalan
	var existing int64
	database.DB.Model(&models.Payment{}).
		Where("booking_id = ? AND status IN ?", b.ID, []string{"processing", "success"}).
		Count(&existing)
	if existing > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Payment already in progress or completed"})
		return
	}

	p := models.Payment{
		ID:        uuid.New().String(),
		BookingID: b.ID,
		UserID:    userID,
		Amount:    b.TotalPrice,
		Status:    models.PaymentProcessing,
	}
	if err := database.DB.Create(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	if err := messaging.PublishPaymentRequest(messaging.PaymentRequest{
		PaymentID: p.ID,
		BookingID: p.BookingID,
		UserID:    p.UserID,
		Amount:    p.Amount,
		CreatedAt: time.Now().Format(time.RFC3339),
	}); err != nil {
		database.DB.Model(&p).Update("status", models.PaymentFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue payment"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Payment is being processed", "payment": p})
}

func GetMyPayments(c *gin.Context) {
	var payments []models.Payment
	database.DB.Where("user_id = ?", c.GetString("user_id")).Order("created_at desc").Find(&payments)
	c.JSON(http.StatusOK, payments)
}