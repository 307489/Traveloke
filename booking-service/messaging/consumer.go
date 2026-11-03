package messaging

import (
	"booking-service/database"
	"booking-service/models"
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentResult struct {
	PaymentID string `json:"payment_id"`
	BookingID string `json:"booking_id"`
	Status    string `json:"status"` 
}

func StartPaymentResultConsumer() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel: ", err)
	}

	ch.ExchangeDeclare("payment.exchange", "direct", true, false, false, false, nil)
	q, err := ch.QueueDeclare("payment.result.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to declare queue: ", err)
	}
	ch.QueueBind(q.Name, "payment.result", "payment.exchange", false, nil)

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to consume: ", err)
	}

	log.Println("Booking service listening to payment.result.queue")

	go func() {
		for msg := range msgs {
			var res PaymentResult
			if err := json.Unmarshal(msg.Body, &res); err != nil {
				log.Println("Bad message:", err)
				msg.Nack(false, false)
				continue
			}

			if err := applyPaymentResult(res); err != nil {
				log.Printf("Failed to apply result for booking %s: %v\n", res.BookingID, err)
			}
			msg.Ack(false)
		}
	}()
}

func applyPaymentResult(res PaymentResult) error {
	if res.Status != "success" {
		log.Printf("Payment %s failed, booking %s stays held\n", res.PaymentID, res.BookingID)
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		var b models.Booking
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", res.BookingID).First(&b).Error; err != nil {
			return err
		}

		if b.Status != models.BookingHeld {
			return errors.New("booking is not held (status: " + string(b.Status) + ")")
		}
		if time.Now().After(b.HoldExpires) {
			return errors.New("hold already expired")
		}

		b.Status = models.BookingPaid
		log.Printf("Booking %s marked as paid\n", b.ID)
		return tx.Save(&b).Error
	})
}