package main

import (
	"booking-service/database"
	"booking-service/handlers"
	"booking-service/middleware"
	"booking-service/models"
	"booking-service/messaging"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func seed() {
	var count int64
	database.DB.Model(&models.Flight{}).Count(&count)
	if count > 0 {
		return
	}

	base := time.Now().Add(48 * time.Hour).Truncate(time.Hour)
	flights := []models.Flight{
		{ID: uuid.New().String(), Airline: "Garuda", FlightNumber: "GA400", Origin: "Jakarta", Destination: "Bali", DepartureTime: base, ArrivalTime: base.Add(2 * time.Hour), Price: 1500000, TotalSeats: 100, AvailableSeats: 100},
		{ID: uuid.New().String(), Airline: "Lion Air", FlightNumber: "JT20", Origin: "Jakarta", Destination: "Bali", DepartureTime: base.Add(3 * time.Hour), ArrivalTime: base.Add(5 * time.Hour), Price: 900000, TotalSeats: 150, AvailableSeats: 150},
		{ID: uuid.New().String(), Airline: "Citilink", FlightNumber: "QG800", Origin: "Jakarta", Destination: "Surabaya", DepartureTime: base.Add(1 * time.Hour), ArrivalTime: base.Add(2*time.Hour + 30*time.Minute), Price: 700000, TotalSeats: 120, AvailableSeats: 120},
		{ID: uuid.New().String(), Airline: "Garuda", FlightNumber: "GA330", Origin: "Bali", Destination: "Jakarta", DepartureTime: base.Add(24 * time.Hour), ArrivalTime: base.Add(26 * time.Hour), Price: 1600000, TotalSeats: 100, AvailableSeats: 100},
	}
	database.DB.Create(&flights)

	hotels := []models.Hotel{
		{ID: uuid.New().String(), Name: "Bali Beach Resort", City: "Bali", Stars: 5, PricePerNight: 2500000, TotalRooms: 50, AvailableRooms: 50},
		{ID: uuid.New().String(), Name: "Kuta Budget Inn", City: "Bali", Stars: 3, PricePerNight: 500000, TotalRooms: 30, AvailableRooms: 30},
		{ID: uuid.New().String(), Name: "Jakarta City Hotel", City: "Jakarta", Stars: 4, PricePerNight: 1200000, TotalRooms: 80, AvailableRooms: 80},
	}
	database.DB.Create(&hotels)
	log.Println("Seeded sample flights and hotels")
}

func startExpiryJob() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			handlers.ExpireHeldBookings()
		}
	}()
}

func main() {
	godotenv.Load()
	database.Connect()
	database.DB.AutoMigrate(&models.Flight{}, &models.Hotel{}, &models.Booking{})
	seed()
	startExpiryJob()
	messaging.StartPaymentResultConsumer()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/flights", handlers.SearchFlights)
	r.GET("/hotels", handlers.SearchHotels)

	internal := r.Group("/internal")
	internal.Use(middleware.InternalOnly())
	{
		internal.GET("/bookings/:id", handlers.InternalGetBooking)
	}

	protected := r.Group("/bookings")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.POST("", handlers.CreateBooking)
		protected.GET("/me", handlers.GetMyBookings)
		protected.GET("/:id", handlers.GetBooking)
		protected.POST("/:id/cancel", handlers.CancelBooking)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	r.Run(":" + port)
}