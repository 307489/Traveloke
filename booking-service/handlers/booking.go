package handlers

import (
	"booking-service/database"
	"booking-service/models"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const holdDuration = 10 * time.Minute


func CreateBooking(c *gin.Context) {
	userID := c.GetString("user_id")

	var input struct {
		Type     string `json:"type" binding:"required,oneof=flight hotel"`
		ItemID   string `json:"item_id" binding:"required"`
		Quantity int    `json:"quantity" binding:"required,min=1,max=10"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	var booking models.Booking

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var price float64

		if input.Type == "flight" {
			var f models.Flight
			// lock row supaya 2 request bersamaan tidak bisa ambil kursi terakhir
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", input.ItemID).First(&f).Error; err != nil {
				return errors.New("flight not found")
			}
			if f.AvailableSeats < input.Quantity {
				return errors.New("not enough seats available")
			}
			f.AvailableSeats -= input.Quantity
			if err := tx.Save(&f).Error; err != nil {
				return err
			}
			price = f.Price * float64(input.Quantity)
		} else {
			var h models.Hotel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", input.ItemID).First(&h).Error; err != nil {
				return errors.New("hotel not found")
			}
			if h.AvailableRooms < input.Quantity {
				return errors.New("not enough rooms available")
			}
			h.AvailableRooms -= input.Quantity
			if err := tx.Save(&h).Error; err != nil {
				return err
			}
			price = h.PricePerNight * float64(input.Quantity)
		}

		booking = models.Booking{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        models.BookingType(input.Type),
			ItemID:      input.ItemID,
			Quantity:    input.Quantity,
			TotalPrice:  price,
			Status:      models.BookingHeld,
			HoldExpires: time.Now().Add(holdDuration),
		}
		return tx.Create(&booking).Error
	})

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Booking held. Complete payment before it expires.",
		"booking": booking,
	})
}


func GetMyBookings(c *gin.Context) {
	userID := c.GetString("user_id")

	var bookings []models.Booking
	database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&bookings)

	c.JSON(http.StatusOK, bookings)
}


func GetBooking(c *gin.Context) {
	userID := c.GetString("user_id")

	var booking models.Booking
	if err := database.DB.Where("id = ? AND user_id = ?", c.Param("id"), userID).
		First(&booking).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

func CancelBooking(c *gin.Context) {
	userID := c.GetString("user_id")

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var b models.Booking
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", c.Param("id"), userID).First(&b).Error; err != nil {
			return errors.New("booking not found")
		}
		if b.Status != models.BookingHeld {
			return errors.New("only held bookings can be cancelled")
		}
		if err := releaseInventory(tx, &b); err != nil {
			return err
		}
		b.Status = models.BookingCancelled
		return tx.Save(&b).Error
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking cancelled"})
}


func releaseInventory(tx *gorm.DB, b *models.Booking) error {
	if b.Type == models.BookingFlight {
		return tx.Model(&models.Flight{}).Where("id = ?", b.ItemID).
			UpdateColumn("available_seats", gorm.Expr("available_seats + ?", b.Quantity)).Error
	}
	return tx.Model(&models.Hotel{}).Where("id = ?", b.ItemID).
		UpdateColumn("available_rooms", gorm.Expr("available_rooms + ?", b.Quantity)).Error
}

func ExpireHeldBookings() {
	var expired []models.Booking
	database.DB.Where("status = ? AND hold_expires < ?", models.BookingHeld, time.Now()).
		Find(&expired)

	for i := range expired {
		b := expired[i]
		database.DB.Transaction(func(tx *gorm.DB) error {
			// re-check di dalam lock, siapa tahu baru saja dibayar
			var cur models.Booking
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", b.ID).First(&cur).Error; err != nil {
				return err
			}
			if cur.Status != models.BookingHeld {
				return nil
			}
			if err := releaseInventory(tx, &cur); err != nil {
				return err
			}
			cur.Status = models.BookingExpired
			return tx.Save(&cur).Error
		})
	}
}