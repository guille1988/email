package messaging

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/handlers"
	"email/internal/infrastructure/rabbitmq"
)

type RabbitMQRegister struct {
	handlers map[string]rabbitmq.MessageHandler
	consumer *rabbitmq.Consumer
}

func NewRabbitMQRegister(consumer *rabbitmq.Consumer, sendWelcomeAction *actions.SendWelcome) *RabbitMQRegister {
	provider := &RabbitMQRegister{
		consumer: consumer,
		handlers: make(map[string]rabbitmq.MessageHandler),
	}

	provider.register("auth.events", "user.created", handlers.NewWelcomeEmail(sendWelcomeAction))

	return provider
}

func (provider *RabbitMQRegister) Close() error {
	if provider.consumer != nil {
		return provider.consumer.Close()
	}
	return nil
}

func (provider *RabbitMQRegister) register(exchange, routingKey string, handler rabbitmq.MessageHandler) {
	provider.handlers[routingKey] = handler
	_ = provider.consumer.Bind(exchange, routingKey)
}

func (provider *RabbitMQRegister) GetHandler(routingKey string) (rabbitmq.MessageHandler, bool) {
	handler, ok := provider.handlers[routingKey]

	return handler, ok
}
