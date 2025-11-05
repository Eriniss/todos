package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"testbox/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

type TodoEvent struct {
	Action string      `json:"action"` // "created", "updated", "deleted"
	TodoID string      `json:"todo_id"`
	Data   interface{} `json:"data,omitempty"`
}

func NewRabbitMQ(cfg *config.Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue for todo events
	queue, err := channel.QueueDeclare(
		"todo_events", // queue name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Println("✓ Connected to RabbitMQ")

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		queue:   queue,
	}, nil
}

// PublishEvent publishes a todo event to the queue
func (r *RabbitMQ) PublishEvent(event TodoEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("✓ Published event: %s for todo:%s", event.Action, event.TodoID)
	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
