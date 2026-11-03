package messaging

import (
	"encoding/json"
	"log"
	"os"
	"payment-service/database"
	"payment-service/models"
	"time"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentRequest struct {
	PaymentID string  `json:"payment_id"`
	BookingID string  `json:"booking_id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"created_at"`
}

type PaymentResult struct {
	PaymentID string `json:"payment_id"`
	BookingID string `json:"booking_id"`
	Status    string `json:"status"`
}

var ch *amqp.Channel

func Connect() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ: ", err)
	}
	ch, err = conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel: ", err)
	}

	ch.ExchangeDeclare("payment.exchange", "direct", true, false, false, false, nil)

	for _, b := range []struct{ queue, key string }{
		{"payment.queue", "payment.request"},
		{"payment.result.queue", "payment.result"},
	} {
		q, err := ch.QueueDeclare(b.queue, true, false, false, false, nil)
		if err != nil {
			log.Fatal("Failed to declare queue: ", err)
		}
		ch.QueueBind(q.Name, b.key, "payment.exchange", false, nil)
	}
}

func publish(key string, v any) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return ch.Publish("payment.exchange", key, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func PublishPaymentRequest(r PaymentRequest) error {
	return publish("payment.request", r)
}

var errInsufficient = errors.New("insufficient balance")

// potong saldo + update status payment dalam satu transaksi
func processPayment(req PaymentRequest) (models.PaymentStatus, string) {
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var w models.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", req.UserID).First(&w).Error; err != nil {
			return errInsufficient // belum punya wallet = saldo 0
		}
		if w.Balance < req.Amount {
			return errInsufficient
		}
		w.Balance -= req.Amount
		if err := tx.Save(&w).Error; err != nil {
			return err
		}
		return tx.Model(&models.Payment{}).Where("id = ?", req.PaymentID).
			Update("status", models.PaymentSuccess).Error
	})

	switch {
	case err == nil:
		return models.PaymentSuccess, ""
	case errors.Is(err, errInsufficient):
		return models.PaymentFailed, "insufficient balance"
	default:
		return models.PaymentFailed, "internal error"
	}
}

func StartConsumer() {
	msgs, err := ch.Consume("payment.queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Failed to consume: ", err)
	}
	log.Println("Payment service listening to payment.queue")

	go func() {
		for msg := range msgs {
			var req PaymentRequest
			if err := json.Unmarshal(msg.Body, &req); err != nil {
				msg.Nack(false, false)
				continue
			}

			log.Printf("Processing payment %s for booking %s\n", req.PaymentID, req.BookingID)
			time.Sleep(3 * time.Second) // mock payment gateway

			status, reason := processPayment(req)
			if status == models.PaymentFailed {
				database.DB.Model(&models.Payment{}).Where("id = ?", req.PaymentID).
					Updates(map[string]any{"status": status, "failure_reason": reason})
				log.Printf("Payment %s failed: %s\n", req.PaymentID, reason)
			}

			if err := publish("payment.result", PaymentResult{
				PaymentID: req.PaymentID,
				BookingID: req.BookingID,
				Status:    string(status),
			}); err != nil {
				log.Println("Failed to publish result:", err)
				msg.Nack(false, true)
				continue
			}
			msg.Ack(false)
		}
	}()
}