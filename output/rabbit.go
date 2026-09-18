package output

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chihqiang/cdc-trigger/pkg/structx"
	"github.com/chihqiang/cdc-trigger/types"
	"github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig RabbitMQ configuration entity
type RabbitMQConfig struct {
	URL        string `json:"url,default=amqp://guest:guest@127.0.0.1:5672/"`
	Exchange   string `json:"exchange,default=cdc-trigger-exchange"`
	Queue      string `json:"queue,default=cdc-trigger-events"`
	Durable    bool   `json:"durable,default=true"`
	AutoDelete bool   `json:"auto_delete"`
	AutoAck    bool   `json:"auto_ack"`
	Exclusive  bool   `json:"exclusive"`
	NoWait     bool   `json:"no_wait"`
}

// RabbitMQOutput RabbitMQ output implementation
type RabbitMQOutput struct {
	config RabbitMQConfig
	conn   *amqp091.Connection
	ch     *amqp091.Channel
}

// NewRabbitMQOutput Creates a RabbitMQOutput and tests the connection
func NewRabbitMQOutput(cfg RabbitMQConfig) (*RabbitMQOutput, error) {
	var (
		err error
	)
	cfg, err = structx.MergeWithDefaults[RabbitMQConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Establish RabbitMQ connection
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	// Declare the queue
	_, err = ch.QueueDeclare(
		cfg.Queue,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		cfg.NoWait,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}
	return &RabbitMQOutput{
		config: cfg,
		conn:   conn,
		ch:     ch,
	}, nil
}

// Send Serializes EventJSON to a JSON string and sends it to RabbitMQ
func (r *RabbitMQOutput) Send(ctx context.Context, event types.EventData) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	return r.ch.PublishWithContext(ctx,
		r.config.Exchange,
		r.config.Queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
}

// Close Closes the RabbitMQ connection
func (r *RabbitMQOutput) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
	return nil
}
