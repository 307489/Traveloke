package models

import "time"

type BookingType string
type BookingStatus string

const (
	BookingFlight BookingType = "flight"
	BookingHotel  BookingType = "hotel"
)

const (
	BookingHeld      BookingStatus = "held"
	BookingPaid      BookingStatus = "paid"
	BookingExpired   BookingStatus = "expired"
	BookingCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID          string        `gorm:"primaryKey;type:varchar(255)" json:"id"`
	UserID      string        `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Type        BookingType   `gorm:"type:varchar(20);not null" json:"type"`
	ItemID      string        `gorm:"type:varchar(255);not null" json:"item_id"`
	Quantity    int           `gorm:"not null" json:"quantity"`
	TotalPrice  float64       `gorm:"not null" json:"total_price"`
	Status      BookingStatus `gorm:"index;type:varchar(20);not null" json:"status"`
	HoldExpires time.Time     `gorm:"index" json:"hold_expires_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}