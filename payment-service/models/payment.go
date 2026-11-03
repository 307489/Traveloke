package models

import "time"

type PaymentStatus string

const (
	PaymentProcessing PaymentStatus = "processing"
	PaymentSuccess    PaymentStatus = "success"
	PaymentFailed     PaymentStatus = "failed"
)

type Payment struct {
	ID        string        `gorm:"primaryKey;type:varchar(255)" json:"id"`
	BookingID string        `gorm:"index;type:varchar(255);not null" json:"booking_id"`
	UserID    string        `gorm:"index;type:varchar(255);not null" json:"user_id"`
	Amount    float64       `gorm:"not null" json:"amount"`
	Status    PaymentStatus `gorm:"type:varchar(20);not null" json:"status"`
	FailureReason string    `json:"failure_reason,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}