package messaging

import (
	"context"
	"email/internal/domain/email/actions"
	"email/internal/domain/email/handlers"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/rabbitmq"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type queueEntry struct {
	consumer *rabbitmq.Consumer
	handlers map[string]rabbitmq.MessageHandler
}

type RabbitMQRegister struct {
	cfg               config.RabbitMQConfig
	queues            map[string]*queueEntry
	sendWelcomeAction *actions.SendWelcome
}

func NewRabbitMQRegister(cfg config.RabbitMQConfig, sendWelcomeAction *actions.SendWelcome) *RabbitMQRegister {
	return &RabbitMQRegister{
		cfg:               cfg,
		queues:            make(map[string]*queueEntry),
		sendWelcomeAction: sendWelcomeAction,
	}
}

func (provider *RabbitMQRegister) RegisterAll(ctx context.Context) error {
	err := provider.register(
		"email.service",
		"auth.events",
		"user.created",
		handlers.NewWelcomeEmail(provider.sendWelcomeAction),
	)

	if err != nil {
		return err
	}

	for queueName, entry := range provider.queues {
		err = entry.consumer.Consume(ctx, queueName, func(delivery amqp.Delivery) error {
			handler, ok := entry.handlers[delivery.RoutingKey]

			if !ok {
				slog.Warn(
					"no handler registered for routing key",
					"routing_key",
					delivery.RoutingKey,
				)

				return nil
			}

			slog.Info(
				"message received from rabbitmq",
				"routing_key",
				delivery.RoutingKey,
			)

			return handler.Handle(delivery.Body)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (provider *RabbitMQRegister) Close() error {
	for _, entry := range provider.queues {
		if err := entry.consumer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (provider *RabbitMQRegister) register(queue, exchange, routingKey string, handler rabbitmq.MessageHandler) error {
	entry, ok := provider.queues[queue]

	if !ok {
		consumer, err := rabbitmq.NewConsumer(provider.cfg)
		if err != nil {
			return err
		}
		entry = &queueEntry{
			consumer: consumer,
			handlers: make(map[string]rabbitmq.MessageHandler),
		}
		provider.queues[queue] = entry
	}

	err := entry.consumer.Bind(queue, exchange, routingKey)

	if err != nil {
		return err
	}

	entry.handlers[routingKey] = handler

	return nil
}
