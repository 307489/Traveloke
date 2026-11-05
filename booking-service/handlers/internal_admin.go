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


func InternalCreateFlight(c *gin.Context) {
	var input struct {
		Airline       string    `json:"airline" binding:"required"`
		FlightNumber  string    `json:"flight_number" binding:"required"`
		Origin        string    `json:"origin" binding:"required"`
		Destination   string    `json:"destination" binding:"required"`
		DepartureTime time.Time `json:"departure_time" binding:"required"`
		ArrivalTime   time.Time `json:"arrival_time" binding:"required"`
		Price         float64   `json:"price" binding:"required,gt=0"`
		TotalSeats    int       `json:"total_seats" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}
	if !input.ArrivalTime.After(input.DepartureTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arrival_time must be after departure_time"})
		return
	}

	f := models.Flight{
		ID: uuid.New().String(), Airline: input.Airline, FlightNumber: input.FlightNumber,
		Origin: input.Origin, Destination: input.Destination,
		DepartureTime: input.DepartureTime, ArrivalTime: input.ArrivalTime,
		Price: input.Price, TotalSeats: input.TotalSeats, AvailableSeats: input.TotalSeats,
	}
	if err := database.DB.Create(&f).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flight"})
		return
	}
	c.JSON(http.StatusCreated, f)
}

func InternalUpdateFlight(c *gin.Context) {
	var input struct {
		Airline       *string    `json:"airline"`
		FlightNumber  *string    `json:"flight_number"`
		Origin        *string    `json:"origin"`
		Destination   *string    `json:"destination"`
		DepartureTime *time.Time `json:"departure_time"`
		ArrivalTime   *time.Time `json:"arrival_time"`
		Price         *float64   `json:"price" binding:"omitempty,gt=0"`
		TotalSeats    *int       `json:"total_seats" binding:"omitempty,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	var f models.Flight
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", c.Param("id")).First(&f).Error; err != nil {
			return errors.New("flight not found")
		}

		if input.Airline != nil {
			f.Airline = *input.Airline
		}
		if input.FlightNumber != nil {
			f.FlightNumber = *input.FlightNumber
		}
		if input.Origin != nil {
			f.Origin = *input.Origin
		}
		if input.Destination != nil {
			f.Destination = *input.Destination
		}
		if input.DepartureTime != nil {
			f.DepartureTime = *input.DepartureTime
		}
		if input.ArrivalTime != nil {
			f.ArrivalTime = *input.ArrivalTime
		}
		if input.Price != nil {
			f.Price = *input.Price
		}
		if input.TotalSeats != nil {
			newAvailable := f.AvailableSeats + (*input.TotalSeats - f.TotalSeats)
			if newAvailable < 0 {
				return errors.New("total_seats cannot be lower than seats already booked")
			}
			f.TotalSeats = *input.TotalSeats
			f.AvailableSeats = newAvailable
		}
		if !f.ArrivalTime.After(f.DepartureTime) {
			return errors.New("arrival_time must be after departure_time")
		}
		return tx.Save(&f).Error
	})
	if err != nil {
		code := http.StatusBadRequest
		if err.Error() == "flight not found" {
			code = http.StatusNotFound
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, f)
}

func InternalDeleteFlight(c *gin.Context) {
	deleteInventory(c, "flight", &models.Flight{})
}


func InternalCreateHotel(c *gin.Context) {
	var input struct {
		Name          string  `json:"name" binding:"required"`
		City          string  `json:"city" binding:"required"`
		Stars         int     `json:"stars" binding:"required,min=1,max=5"`
		PricePerNight float64 `json:"price_per_night" binding:"required,gt=0"`
		TotalRooms    int     `json:"total_rooms" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	h := models.Hotel{
		ID: uuid.New().String(), Name: input.Name, City: input.City, Stars: input.Stars,
		PricePerNight: input.PricePerNight, TotalRooms: input.TotalRooms, AvailableRooms: input.TotalRooms,
	}
	if err := database.DB.Create(&h).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create hotel"})
		return
	}
	c.JSON(http.StatusCreated, h)
}

func InternalUpdateHotel(c *gin.Context) {
	var input struct {
		Name          *string  `json:"name"`
		City          *string  `json:"city"`
		Stars         *int     `json:"stars" binding:"omitempty,min=1,max=5"`
		PricePerNight *float64 `json:"price_per_night" binding:"omitempty,gt=0"`
		TotalRooms    *int     `json:"total_rooms" binding:"omitempty,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": formatValidationError(err)})
		return
	}

	var h models.Hotel
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", c.Param("id")).First(&h).Error; err != nil {
			return errors.New("hotel not found")
		}

		if input.Name != nil {
			h.Name = *input.Name
		}
		if input.City != nil {
			h.City = *input.City
		}
		if input.Stars != nil {
			h.Stars = *input.Stars
		}
		if input.PricePerNight != nil {
			h.PricePerNight = *input.PricePerNight
		}
		if input.TotalRooms != nil {
			newAvailable := h.AvailableRooms + (*input.TotalRooms - h.TotalRooms)
			if newAvailable < 0 {
				return errors.New("total_rooms cannot be lower than rooms already booked")
			}
			h.TotalRooms = *input.TotalRooms
			h.AvailableRooms = newAvailable
		}
		return tx.Save(&h).Error
	})
	if err != nil {
		code := http.StatusBadRequest
		if err.Error() == "hotel not found" {
			code = http.StatusNotFound
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h)
}

func InternalDeleteHotel(c *gin.Context) {
	deleteInventory(c, "hotel", &models.Hotel{})
}


func deleteInventory(c *gin.Context, kind string, model interface{}) {
	id := c.Param("id")

	var active int64
	database.DB.Model(&models.Booking{}).
		Where("type = ? AND item_id = ? AND status IN ?", kind, id, []string{"held", "paid"}).
		Count(&active)
	if active > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete: there are active bookings for this item"})
		return
	}

	res := database.DB.Where("id = ?", id).Delete(model)
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": kind + " not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": kind + " deleted"})
}


func InternalAnalytics(c *gin.Context) {
	var total int64
	database.DB.Model(&models.Booking{}).Count(&total)

	type statusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var byStatus []statusCount
	database.DB.Model(&models.Booking{}).
		Select("status, COUNT(*) AS count").Group("status").Scan(&byStatus)

	var revenue float64
	database.DB.Model(&models.Booking{}).
		Where("status = ?", models.BookingPaid).
		Select("COALESCE(SUM(total_price), 0)").Scan(&revenue)

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	var today int64
	database.DB.Model(&models.Booking{}).Where("created_at >= ?", startOfDay).Count(&today)

	type routeCount struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Bookings    int64  `json:"bookings"`
	}
	var topRoutes []routeCount
	database.DB.Raw(`
		SELECT f.origin, f.destination, COUNT(*) AS bookings
		FROM bookings b
		JOIN flights f ON f.id = b.item_id
		WHERE b.type = 'flight' AND b.status IN ('held', 'paid')
		GROUP BY f.origin, f.destination
		ORDER BY bookings DESC
		LIMIT 5`).Scan(&topRoutes)

	c.JSON(http.StatusOK, gin.H{
		"total_bookings":  total,
		"bookings_today":  today,
		"by_status":       byStatus,
		"total_revenue":   revenue,
		"top_routes":      topRoutes,
	})
}