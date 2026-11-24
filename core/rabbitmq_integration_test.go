//go:build integration

package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests need a RabbitMQ broker: RABBITMQ_TEST_URL=amqp://guest:guest@localhost:5672/ . They are skipped without it.
func brokerURL(t *testing.T) string {
	t.Helper()
	u := os.Getenv("RABBITMQ_TEST_URL")
	if u == "" {
		t.Skip("RABBITMQ_TEST_URL is not set")
	}
	return u
}

// listen binds a throwaway queue to the events exchange for one routing key, the way a consuming service does.
func listen(t *testing.T, url, key string) <-chan amqp.Delivery {
	t.Helper()
	conn, err := amqp.Dial(url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	ch, err := conn.Channel()
	require.NoError(t, err)
	require.NoError(t, ch.ExchangeDeclare(EventsExchange, amqp.ExchangeTopic, true, false, false, false, nil))
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	require.NoError(t, err)
	require.NoError(t, ch.QueueBind(q.Name, key, EventsExchange, false, nil))
	deliveries, err := ch.Consume(q.Name, "", true, true, false, false, nil)
	require.NoError(t, err)
	return deliveries
}

func next(t *testing.T, deliveries <-chan amqp.Delivery) amqp.Delivery {
	t.Helper()
	select {
	case d := <-deliveries:
		return d
	case <-time.After(10 * time.Second):
		t.Fatal("no event arrived")
		return amqp.Delivery{}
	}
}

func TestRabbitPublisherPublishesAConfirmedJSONEvent(t *testing.T) {
	url := brokerURL(t)
	key := fmt.Sprintf("test.published.%d", time.Now().UnixNano())
	deliveries := listen(t, url, key)

	p := NewPublisher(url).(*RabbitPublisher)
	t.Cleanup(p.Close)
	require.NoError(t, p.Publish(context.Background(), key, map[string]any{"user_id": "u-1", "amount": 35000}))

	d := next(t, deliveries)
	assert.Equal(t, key, d.RoutingKey, "the routing key is the event name")
	assert.Equal(t, "application/json", d.ContentType)
	assert.Equal(t, uint8(amqp.Persistent), d.DeliveryMode)
	assert.Len(t, d.MessageId, 32, "each event carries an id a consumer can deduplicate on")
	var payload map[string]any
	require.NoError(t, json.Unmarshal(d.Body, &payload))
	assert.Equal(t, "u-1", payload["user_id"])

	require.NoError(t, p.Publish(context.Background(), key, map[string]any{"user_id": "u-2"}))
	second := next(t, deliveries)
	assert.NotEqual(t, d.MessageId, second.MessageId)
}

func TestRabbitPublisherReconnectsAfterALostConnection(t *testing.T) {
	url := brokerURL(t)
	key := fmt.Sprintf("test.reconnect.%d", time.Now().UnixNano())
	deliveries := listen(t, url, key)

	p := NewPublisher(url).(*RabbitPublisher)
	t.Cleanup(p.Close)
	require.NoError(t, p.Publish(context.Background(), key, map[string]any{"n": 1}))
	next(t, deliveries)

	// The connection dies behind the publisher's back, as when the broker restarts.
	p.mu.Lock()
	_ = p.conn.Close()
	p.mu.Unlock()

	require.NoError(t, p.Publish(context.Background(), key, map[string]any{"n": 2}))
	var payload map[string]any
	require.NoError(t, json.Unmarshal(next(t, deliveries).Body, &payload))
	assert.EqualValues(t, 2, payload["n"])
}
