package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"9itchi-tracker-api/utils"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	// EventsExchange is the topic exchange every service publishes its events to; the routing key is the event name.
	EventsExchange = "9itchi.events"

	rabbitDialTimeout    = 3 * time.Second
	rabbitPublishTimeout = 5 * time.Second
	// rabbitRetryAfter is how long a broker that could not be reached is left alone, so that a broker that is down
	// costs one short wait and not one per request.
	rabbitRetryAfter = 15 * time.Second
)

// ErrBrokerDown is returned while the broker is being left alone after a failed connection.
var ErrBrokerDown = errors.New("rabbitmq: broker unreachable")

// RabbitPublisher publishes events to EventsExchange as persistent JSON messages, each confirmed by the broker.
// It connects on first use and reconnects after a failure.
type RabbitPublisher struct {
	url string

	mu        sync.Mutex
	conn      *amqp.Connection
	ch        *amqp.Channel
	downUntil time.Time
}

// NewPublisher is the Publisher factory: RabbitMQ when url is set, else LogPublisher.
func NewPublisher(url string) Publisher {
	if url == "" {
		return LogPublisher{}
	}
	return &RabbitPublisher{url: url}
}

// Publish sends one event. The caller decides what a failure means; it never carries the broker URL.
func (p *RabbitPublisher) Publish(ctx context.Context, event string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rabbitmq: encode %s: %w", event, err)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	var last error
	// A channel that died since the last publish is only found out by using it: one more try on a fresh one.
	for attempt := 0; attempt < 2; attempt++ {
		if last = p.send(ctx, event, body); last == nil {
			return nil
		}
		if errors.Is(last, ErrBrokerDown) {
			break
		}
		p.reset()
	}
	return fmt.Errorf("rabbitmq: publish %s: %s", event, redactAMQP(last, p.url))
}

func (p *RabbitPublisher) send(ctx context.Context, event string, body []byte) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, rabbitPublishTimeout)
	defer cancel()
	confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx, EventsExchange, event, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    newEventID(),
		Timestamp:    time.Now(),
		Body:         body,
	})
	if err != nil {
		return err
	}
	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return errors.New("not confirmed by the broker")
	}
	return nil
}

// channel returns the open channel, connecting when there is none.
func (p *RabbitPublisher) channel() (*amqp.Channel, error) {
	if p.ch != nil && !p.ch.IsClosed() {
		return p.ch, nil
	}
	if time.Now().Before(p.downUntil) {
		return nil, ErrBrokerDown
	}
	p.reset()
	conn, err := amqp.DialConfig(p.url, amqp.Config{Dial: amqp.DefaultDial(rabbitDialTimeout)})
	if err != nil {
		p.downUntil = time.Now().Add(rabbitRetryAfter)
		utils.Logger.Warn("rabbitmq unreachable, events are not published for now", zap.String("cause", redactAMQP(err, p.url)))
		return nil, err
	}
	ch, err := conn.Channel()
	if err == nil {
		err = ch.ExchangeDeclare(EventsExchange, amqp.ExchangeTopic, true, false, false, false, nil)
	}
	if err == nil {
		err = ch.Confirm(false)
	}
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	p.conn, p.ch = conn, ch
	return ch, nil
}

func (p *RabbitPublisher) reset() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	p.conn, p.ch = nil, nil
}

// Close ends the connection. The publisher can still be used: it reconnects.
func (p *RabbitPublisher) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.reset()
}

func newEventID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

// redactAMQP writes an error without the broker URL or its password, which a parse error can carry.
func redactAMQP(err error, rawURL string) string {
	if err == nil {
		return ""
	}
	msg := strings.ReplaceAll(err.Error(), rawURL, "RABBITMQ_URL")
	if u, perr := url.Parse(rawURL); perr == nil && u.User != nil {
		if pw, ok := u.User.Password(); ok && pw != "" {
			msg = strings.ReplaceAll(msg, pw, "REDACTED")
		}
	}
	return msg
}
