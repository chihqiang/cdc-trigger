package output

import (
	"context"
	"encoding/json"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/chihqiang/cdc-trigger/types"
)

type PulsarConfig struct {
	URL               string `json:"url,default=pulsar://localhost:6650"`
	Topic             string `json:"topic,default=cdc-trigger-events"`
	Token             string `json:"token"`
	OperationTimeout  int    `json:"operation_timeout,default=30"`
	ConnectionTimeout int    `json:"connection_timeout,default=30"`
}

type PulsarOutput struct {
	cfg      PulsarConfig
	client   pulsar.Client
	producer pulsar.Producer
}

// NewPulsarOutput initializes the Pulsar client and producer
func NewPulsarOutput(cfg PulsarConfig) (*PulsarOutput, error) {
	o := &PulsarOutput{cfg: cfg}

	clientOptions := pulsar.ClientOptions{
		URL:               cfg.URL,
		OperationTimeout:  time.Duration(cfg.OperationTimeout) * time.Second,
		ConnectionTimeout: time.Duration(cfg.ConnectionTimeout) * time.Second,
	}

	if cfg.Token != "" {
		clientOptions.Authentication = pulsar.NewAuthenticationToken(cfg.Token)
	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: cfg.Topic,
	})
	if err != nil {
		client.Close()
		return nil, err
	}
	o.client = client
	o.producer = producer
	return o, nil
}

// Send sends an event to Pulsar
func (p *PulsarOutput) Send(ctx context.Context, event types.EventData) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})

	return err
}

// Close closes the producer and client
func (p *PulsarOutput) Close() error {
	if p.producer != nil {
		p.producer.Close()
	}
	if p.client != nil {
		p.client.Close()
	}
	return nil
}
