package models

import "time"

type Flight struct {
	ID             string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Airline        string    `gorm:"not null" json:"airline"`
	FlightNumber   string    `gorm:"not null" json:"flight_number"`
	Origin         string    `gorm:"index;not null" json:"origin"`
	Destination    string    `gorm:"index;not null" json:"destination"`
	DepartureTime  time.Time `gorm:"index;not null" json:"departure_time"`
	ArrivalTime    time.Time `gorm:"not null" json:"arrival_time"`
	Price          float64   `gorm:"not null" json:"price"`
	TotalSeats     int       `gorm:"not null" json:"total_seats"`
	AvailableSeats int       `gorm:"not null" json:"available_seats"`
	CreatedAt      time.Time `json:"created_at"`
}