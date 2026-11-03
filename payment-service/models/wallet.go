package models

import "time"

type Wallet struct {
	UserID    string    `gorm:"primaryKey;type:varchar(255)" json:"user_id"`
	Balance   float64   `gorm:"not null;default:0" json:"balance"`
	UpdatedAt time.Time `json:"updated_at"`
}