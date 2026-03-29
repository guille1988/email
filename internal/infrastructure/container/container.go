package container

import (
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/database"
	"email/internal/infrastructure/rabbitmq"

	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	Consumer          *rabbitmq.Consumer
}

// New creates a new container with initialized database connections.
func New(cfg *config.Config) (*Container, error) {
	defaultConnection, err := database.NewConnection(cfg.Database.Connections[config.Default])

	if err != nil {
		return nil, err
	}

	return &Container{
		DefaultConnection: defaultConnection,
	}, nil
}

// InitConsumer initializes the RabbitMQ consumer.
func (container *Container) InitConsumer(cfg config.RabbitMQConfig, queue string) error {
	consumer, err := rabbitmq.NewConsumer(cfg, queue)

	if err != nil {
		return err
	}

	container.Consumer = consumer

	return nil
}
