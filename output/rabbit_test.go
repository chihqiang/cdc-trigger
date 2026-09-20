package output

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chihqiang/cdc-trigger/types"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeRabbitConfig(t *testing.T) {
	// The routing key cannot default to the queue name in the tag, so it is
	// filled in code.
	cfg := normalizeRabbitConfig(RabbitMQConfig{Queue: "cdc-trigger-events"})
	assert.Equal(t, "cdc-trigger-events", cfg.RoutingKey)
	assert.Equal(t, "direct", cfg.ExchangeType)

	// What the configuration file says wins.
	cfg = normalizeRabbitConfig(RabbitMQConfig{
		Queue:        "cdc-trigger-events",
		RoutingKey:   "cdc.#",
		ExchangeType: "topic",
	})
	assert.Equal(t, "cdc.#", cfg.RoutingKey)
	assert.Equal(t, "topic", cfg.ExchangeType)
}

// A client without a channel must never be reported as healthy: that is what
// made a dropped connection look like a sent message.
func TestRabbitClientErr(t *testing.T) {
	assert.Error(t, (&rabbitClient{}).err())

	c := newRabbitClient(nil)
	assert.Error(t, c.err())

	c.stop(errors.New("the channel was closed"))
	assert.Error(t, c.err())
}

// A publish waiting for its confirmation has to be released when the channel
// dies, instead of waiting for its context.
func TestRabbitClientStopReleasesWaiters(t *testing.T) {
	c := newRabbitClient(nil)
	waiter := make(chan error, 1)
	c.waiters[1] = waiter

	c.stop(errors.New("the channel was closed"))

	select {
	case err := <-waiter:
		assert.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("the waiting publish was never released")
	}
}

// The broker confirms a message it could not route, so the confirmation alone
// says nothing about delivery: the return that precedes it must decide.
func TestRabbitClientReturnBeatsTheConfirmation(t *testing.T) {
	c := newRabbitClient(nil)
	waiter := make(chan error, 1)
	c.waiters[3] = waiter

	c.markReturned(amqp091.Return{
		ReplyCode:  312,
		ReplyText:  "NO_ROUTE",
		Exchange:   "cdc-trigger-exchange",
		RoutingKey: "cdc-trigger-events",
		MessageId:  "3",
	})
	c.settle(amqp091.Confirmation{DeliveryTag: 3, Ack: true})

	select {
	case err := <-waiter:
		assert.ErrorIs(t, err, errUnroutable)
	case <-time.After(time.Second):
		t.Fatal("the publish was not answered")
	}
}

func TestRabbitClientSettle(t *testing.T) {
	// An ack with no return is a delivered message.
	c := newRabbitClient(nil)
	waiter := make(chan error, 1)
	c.waiters[1] = waiter
	c.settle(amqp091.Confirmation{DeliveryTag: 1, Ack: true})
	assert.NoError(t, <-waiter)

	// A nack is a failure.
	c = newRabbitClient(nil)
	waiter = make(chan error, 1)
	c.waiters[2] = waiter
	c.settle(amqp091.Confirmation{DeliveryTag: 2, Ack: false})
	assert.Error(t, <-waiter)

	// A return for a message nobody waits for is not recorded, so a channel
	// shared with another publisher cannot leak into this one's failures.
	c = newRabbitClient(nil)
	c.markReturned(amqp091.Return{MessageId: "1", ReplyCode: 312})
	assert.Empty(t, c.returned)
	c.markReturned(amqp091.Return{MessageId: "not-a-number", ReplyCode: 312})
	assert.Empty(t, c.returned)
}

func TestDeliveryMode(t *testing.T) {
	assert.Equal(t, uint8(2), deliveryMode(true), "a durable queue keeps its messages on disk")
	assert.Equal(t, uint8(1), deliveryMode(false))
}

func TestRabbitMQOutput_SendAfterClose(t *testing.T) {
	out := &RabbitMQOutput{config: normalizeRabbitConfig(RabbitMQConfig{Queue: "cdc-trigger-events"})}
	assert.NoError(t, out.Close())
	assert.NoError(t, out.Close(), "closing twice is not an error")

	err := out.Send(context.Background(), types.EventData{Time: time.Now()})
	assert.Error(t, err, "a closed output must report the failure instead of reporting success")
}
