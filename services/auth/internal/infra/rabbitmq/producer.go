package rabbitmq

import (
	"context"
	"encoding/json"
	"github.com/1ncrease0/pixly/services/auth/internal/domain/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewProducer(url, queue string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &Producer{
		conn:    conn,
		channel: ch,
		queue:   queue,
	}, nil
}

func (p *Producer) SendVerification(ctx context.Context, e events.EmailVerification) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(
		pubCtx,
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
}

func (p *Producer) Close() error {
	if err := p.channel.Close(); err != nil {
		_ = p.conn.Close()
		return err
	}
	return p.conn.Close()
}
