package models

import "time"

type Hotel struct {
	ID             string    `gorm:"primaryKey;type:varchar(255)" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	City           string    `gorm:"index;not null" json:"city"`
	Stars          int       `json:"stars"`
	PricePerNight  float64   `gorm:"not null" json:"price_per_night"`
	TotalRooms     int       `gorm:"not null" json:"total_rooms"`
	AvailableRooms int       `gorm:"not null" json:"available_rooms"`
	CreatedAt      time.Time `json:"created_at"`
}