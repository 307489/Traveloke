package handlers

import (
	"booking-service/database"
	"booking-service/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func paginate(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return
}

func SearchFlights(c *gin.Context) {
	page, limit := paginate(c)
	q := database.DB.Model(&models.Flight{})

	if from := c.Query("from"); from != "" {
		q = q.Where("LOWER(origin) = LOWER(?)", from)
	}
	if to := c.Query("to"); to != "" {
		q = q.Where("LOWER(destination) = LOWER(?)", to)
	}
	if dateStr := c.Query("date"); dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
			return
		}
		q = q.Where("departure_time >= ? AND departure_time < ?", d, d.Add(24*time.Hour))
	}
	if v := c.Query("min_price"); v != "" {
		q = q.Where("price >= ?", v)
	}
	if v := c.Query("max_price"); v != "" {
		q = q.Where("price <= ?", v)
	}
	if v := c.Query("passengers"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "passengers must be a positive number"})
			return
		}
		q = q.Where("available_seats >= ?", n)
	}

	switch c.Query("sort") {
	case "price":
		q = q.Order("price asc")
	case "price_desc":
		q = q.Order("price desc")
	default:
		q = q.Order("departure_time asc")
	}

	var total int64
	q.Count(&total)

	var flights []models.Flight
	q.Offset((page - 1) * limit).Limit(limit).Find(&flights)

	c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit, "total": total, "flights": flights})
}

func SearchHotels(c *gin.Context) {
	page, limit := paginate(c)
	q := database.DB.Model(&models.Hotel{})

	if city := c.Query("city"); city != "" {
		q = q.Where("LOWER(city) = LOWER(?)", city)
	}
	if v := c.Query("min_stars"); v != "" {
		q = q.Where("stars >= ?", v)
	}
	if v := c.Query("min_price"); v != "" {
		q = q.Where("price_per_night >= ?", v)
	}
	if v := c.Query("max_price"); v != "" {
		q = q.Where("price_per_night <= ?", v)
	}
	if v := c.Query("rooms"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "rooms must be a positive number"})
			return
		}
		q = q.Where("available_rooms >= ?", n)
	}

	switch c.Query("sort") {
	case "price":
		q = q.Order("price_per_night asc")
	case "price_desc":
		q = q.Order("price_per_night desc")
	case "stars":
		q = q.Order("stars desc")
	default:
		q = q.Order("name asc")
	}

	var total int64
	q.Count(&total)

	var hotels []models.Hotel
	q.Offset((page - 1) * limit).Limit(limit).Find(&hotels)

	c.JSON(http.StatusOK, gin.H{"page": page, "limit": limit, "total": total, "hotels": hotels})
}