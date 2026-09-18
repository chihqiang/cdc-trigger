package output

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/chihqiang/cdc-trigger/pkg/structx"
	"github.com/chihqiang/cdc-trigger/types"
)

// RocketMQConfig RocketMQ configuration entity
type RocketMQConfig struct {
	// Servers - RocketMQ NameServer address list, e.g., ["127.0.0.1:9876"]
	Servers []string `json:"servers,default=127.0.0.1:9876"`
	// Topic - The topic name to send the message
	Topic string `json:"topic,default=cdc-trigger-events"`
	// Group - The producer group name
	Group string `json:"group"`
	// Retry - The number of retries if sending a message fails
	Retry int `json:"retry,default=3"`
	// Namespace - The namespace
	Namespace string `json:"namespace"`
	// AccessKey - Access key
	AccessKey string `json:"access_key"`
	// SecretKey - Secret key
	SecretKey string `json:"secret_key"`
}

// RocketMQOutput RocketMQ implementation that satisfies the IOutput interface
type RocketMQOutput struct {
	cfg      RocketMQConfig
	producer rocketmq.Producer
}

// NewRocketMQOutput Creates a RocketMQOutput and fills in default values
func NewRocketMQOutput(cfg RocketMQConfig) (*RocketMQOutput, error) {
	var (
		err error
	)
	cfg, err = structx.MergeWithDefaults[RocketMQConfig](cfg)
	if err != nil {
		return nil, err
	}
	// Create producer options
	options := []producer.Option{
		producer.WithNsResolver(primitive.NewPassthroughResolver(cfg.Servers)),
		producer.WithRetry(cfg.Retry),
	}
	if cfg.Group != "" {
		options = append(options, producer.WithGroupName(cfg.Group))
	}
	if cfg.Namespace != "" {
		options = append(options, producer.WithNamespace(cfg.Namespace))
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		options = append(options, producer.WithCredentials(primitive.Credentials{
			AccessKey: cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		}))
	}
	// Create producer
	p, err := rocketmq.NewProducer(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create RocketMQ producer: %w", err)
	}
	// Start the producer
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("failed to start RocketMQ producer: %w", err)
	}
	return &RocketMQOutput{
		cfg:      cfg,
		producer: p,
	}, nil
}

// Send Serializes the EventData to a JSON string and sends it to RocketMQ
func (r *RocketMQOutput) Send(ctx context.Context, event types.EventData) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &primitive.Message{
		Topic: r.cfg.Topic,
		Body:  data,
	}
	_, err = r.producer.SendSync(ctx, msg)
	return err
}

// Close Closes the RocketMQ producer
func (r *RocketMQOutput) Close() error {
	return r.producer.Shutdown()
}
