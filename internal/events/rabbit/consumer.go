package rabbit

import (
	"carshop/internal/events"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event events.CarCreatedEvent) error

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
}

func NewConsumer(url string, logger *zap.Logger) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		events.ExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(events.QueueName, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.QueueBind(q.Name, events.BindingRoutingKey, events.ExchangeName, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	logger.Info("connected to RabbitMQ as consumer",
		zap.String("queue", events.QueueName),
		zap.String("exchange", events.ExchangeName),
		zap.String("routing_key", events.BindingRoutingKey),
	)

	return &Consumer{conn: conn, channel: ch, logger: logger}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler EventHandler) error {
	msgs, err := c.channel.Consume(events.QueueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	c.logger.Info("started listening for messages")

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("message channel closed")
			}

			var event events.CarCreatedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				c.logger.Error("failed to unmarshal message", zap.Error(err))
				_ = msg.Nack(false, false)
				continue
			}

			if err := handler(ctx, event); err != nil {
				c.logger.Error("handler returned error", zap.Error(err))
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	_ = c.channel.Close()
	_ = c.conn.Close()
}
