package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

// EventListener defines the interface for handling RabbitMQ events.
type EventListener interface {
	Handle(delivery amqp.Delivery) error
}
