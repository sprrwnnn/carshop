package rabbit

import (
	"carshop/internal/events"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	logger  *zap.Logger
}

func NewPublisher(url string, logger *zap.Logger) (*Publisher, error) {
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

	logger.Info("connected to RabbitMQ as publisher",
		zap.String("exchange", events.ExchangeName),
	)

	return &Publisher{conn: conn, channel: ch, logger: logger}, nil
}

func (p *Publisher) PublishCarCreated(ctx context.Context, event events.CarCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := p.channel.PublishWithContext(
		ctx,
		events.ExchangeName,
		events.CarCreatedKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	p.logger.With(zap.Uint64("car_id", event.ID)).Info("published car.created event")

	return nil
}

func (p *Publisher) Close() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}
