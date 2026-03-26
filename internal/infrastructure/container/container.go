package container

import (
	"email/internal/domain/email/actions"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/database"
	"email/internal/infrastructure/rabbitmq"

	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	EmailRepository   model.Repository
	SendWelcomeAction *actions.SendWelcome
	Consumer          *rabbitmq.Consumer
	Config            *config.Config
}

// New creates a new container with initialized database connections.
func New(cfg *config.Config) (*Container, error) {
	defaultConnection, err := database.NewConnection(cfg.Database.Connections[config.Default])

	if err != nil {
		return nil, err
	}

	emailRepo := model.NewRepository(defaultConnection)
	sendWelcomeAction := actions.NewSendWelcome(cfg.Mail, emailRepo)

	return &Container{
		DefaultConnection: defaultConnection,
		EmailRepository:   emailRepo,
		SendWelcomeAction: sendWelcomeAction,
		Config:            cfg,
	}, nil
}

// InitConsumer initializes the RabbitMQ consumer.
func (container *Container) InitConsumer(queue string) error {
	consumer, err := rabbitmq.NewConsumer(container.Config.RabbitMQ, queue)

	if err != nil {
		return err
	}

	container.Consumer = consumer

	return nil
}
