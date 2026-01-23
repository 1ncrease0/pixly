package rabbitmq

import (
	"context"
	"encoding/json"
	"github.com/1ncrease0/pixly/services/notification/internal/domain"
	"github.com/1ncrease0/pixly/services/notification/internal/domain/events"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
	"sync"
)

const (
	defaultWorkerCount = 5
)

type EmailSender interface {
	SendVerification(email domain.Email, code string) error
}

type Consumer struct {
	log     *slog.Logger
	sender  EmailSender
	queue   string
	workers int

	conn    *amqp.Connection
	channel *amqp.Channel
	msgs    <-chan amqp.Delivery
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewConsumer(log *slog.Logger, sender EmailSender, url, queue string, workers int) (*Consumer, error) {
	if workers <= 0 {
		workers = defaultWorkerCount
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := ch.Qos(workers, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &Consumer{
		log:     log,
		sender:  sender,
		queue:   queue,
		workers: workers,
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		c.cancel()
		return err
	}
	c.msgs = msgs

	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	return nil
}

func (c *Consumer) worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.msgs:
			if !ok {
				return
			}
			c.handleMessage(msg)
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	var event events.EmailVerification
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		c.reject(msg, "invalid message", err)
		return
	}

	email, err := domain.NewEmail(event.Email)
	if err != nil {
		c.reject(msg, "invalid email", err)
		return
	}
	if event.Code == "" {
		c.reject(msg, "empty verification code", nil)
		return
	}

	if err := c.sender.SendVerification(email, event.Code); err != nil {
		c.reject(msg, "send verification email failed", err)
		return
	}

	if err := msg.Ack(false); err != nil {
		c.log.Error("rabbitmq ack failed", slog.Any("error", err))
	}
}

func (c *Consumer) Shutdown(ctx context.Context) error {
	c.log.Info("rabbitmq consumer shutting down")
	c.cancel()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		c.log.Warn("rabbitmq shutdown timeout", slog.Any("error", ctx.Err()))
	}

	if err := c.channel.Close(); err != nil {
		c.log.Warn("rabbitmq channel close failed", slog.Any("error", err))
	}
	if err := c.conn.Close(); err != nil {
		c.log.Warn("rabbitmq connection close failed", slog.Any("error", err))
	}

	return nil
}

func (c *Consumer) reject(msg amqp.Delivery, reason string, err error) {
	if err != nil {
		c.log.Warn("rabbitmq message rejected", slog.String("reason", reason), slog.Any("error", err))
	} else {
		c.log.Warn("rabbitmq message rejected", slog.String("reason", reason))
	}
	if err := msg.Nack(false, false); err != nil {
		c.log.Error("rabbitmq nack failed", slog.Any("error", err))
	}
}
