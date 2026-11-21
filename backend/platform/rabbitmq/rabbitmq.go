package rabbitmq

import (
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"zeinai-backend/internal/config"
)

var (
	conn *amqp.Connection
	ch   *amqp.Channel
	once sync.Once
)

// InitRabbitMQ initializes RabbitMQ connection and channel
func InitRabbitMQ(cfg *config.Config) error {
	var err error
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	// Retry logic for connection
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/5): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare queues
	err = declareQueues()
	if err != nil {
		return err
	}

	log.Println("RabbitMQ connection established")
	return nil
}

func declareQueues() error {
	queues := []string{"ai.text", "ai.image", "ai.video"}
	for _, q := range queues {
		_, err := ch.QueueDeclare(
			q,     // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q, err)
		}
	}
	return nil
}

// GetChannel returns the shared channel (useful for simple publishers)
func GetChannel() *amqp.Channel {
	return ch
}

// NotifyConnectionClose returns a channel that receives an error when the
// RabbitMQ connection is lost (nil if not connected).
func NotifyConnectionClose() chan *amqp.Error {
	if conn == nil {
		return nil
	}
	return conn.NotifyClose(make(chan *amqp.Error, 1))
}

// NewChannel opens and returns a new channel on the existing connection
func NewChannel() (*amqp.Channel, error) {
	if conn == nil {
		return nil, fmt.Errorf("rabbitmq connection is not initialized")
	}
	return conn.Channel()
}

// Close closes the connection
func Close() {
	if ch != nil {
		ch.Close()
	}
	if conn != nil {
		conn.Close()
	}
}

// Publish publishes a message to a queue
func Publish(queueName string, body []byte) error {
	if ch == nil {
		return fmt.Errorf("rabbitmq channel is not initialized")
	}

	return ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			DeliveryMode: amqp.Persistent,
		},
	)
}
