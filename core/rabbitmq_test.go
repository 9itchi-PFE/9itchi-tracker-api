package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublisherWithoutURLLogs(t *testing.T) {
	assert.Equal(t, LogPublisher{}, NewPublisher(""))
	assert.IsType(t, &RabbitPublisher{}, NewPublisher("amqp://guest:guest@localhost:5672/"))
}

// A broker that is down must cost one short wait, then nothing, and never surface its URL.
func TestRabbitPublisherLeavesADownBrokerAlone(t *testing.T) {
	const url = "amqp://app:s3cr3t-pw@127.0.0.1:1/"
	p := &RabbitPublisher{url: url}

	err := p.Publish(context.Background(), "test.event", map[string]string{"user_id": "u-1"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "s3cr3t-pw")

	start := time.Now()
	err = p.Publish(context.Background(), "test.event", map[string]string{"user_id": "u-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrBrokerDown.Error())
	assert.Less(t, time.Since(start), 500*time.Millisecond, "no new connection attempt while the broker is left alone")
}

func TestRabbitPublisherRefusesWhatCannotBeEncoded(t *testing.T) {
	p := &RabbitPublisher{url: "amqp://guest:guest@127.0.0.1:1/"}
	err := p.Publish(context.Background(), "test.event", map[string]any{"bad": make(chan int)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "encode")
}

func TestRedactAMQPHidesURLAndPassword(t *testing.T) {
	const url = "amqp://app:s3cr3t-pw@broker.internal:5672/"
	msg := redactAMQP(errors.New(`parse "`+url+`": invalid port; password s3cr3t-pw`), url)
	assert.NotContains(t, msg, "s3cr3t-pw")
	assert.NotContains(t, msg, "broker.internal")
	assert.Empty(t, redactAMQP(nil, url))
}
