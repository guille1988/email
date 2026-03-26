package rabbitmq

import (
	"context"
	"email/internal/infrastructure/config"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      string
}

func NewConsumer(cfg config.RabbitMQConfig, queue string) (*Consumer, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

	connection, err := amqp.Dial(url)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	var channel *amqp.Channel
	var ok bool

	defer func() {
		if !ok {
			if channel != nil {
				_ = channel.Close()
			}
			_ = connection.Close()
		}
	}()

	channel, err = connection.Channel()

	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	ok = true

	return &Consumer{
		connection: connection,
		channel:    channel,
		queue:      queue,
	}, nil
}

func (consumer *Consumer) Consume(ctx context.Context, handler func(delivery amqp.Delivery) error) error {
	messages, err := consumer.channel.Consume(
		consumer.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case delivery, ok := <-messages:
				if !ok {
					return
				}

				err = handler(delivery)

				if err != nil {
					slog.Error("failed to handle message", "error", err)

					_ = delivery.Nack(false, true)
					continue
				}

				_ = delivery.Ack(false)
			}
		}
	}()

	slog.Info("consumer started", "queue", consumer.queue)

	return nil
}

func (consumer *Consumer) Close() error {
	err := consumer.channel.Close()

	if err != nil {
		return err
	}

	return consumer.connection.Close()
}
