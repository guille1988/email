package container

import (
	"email/internal/domain/email"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/database"
	"email/internal/infrastructure/rabbitmq"

	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	EmailRepository   model.Repository
	EmailService      *email.Service
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
	emailService := email.NewEmailService(cfg.Mail)

	return &Container{
		DefaultConnection: defaultConnection,
		EmailRepository:   emailRepo,
		EmailService:      emailService,
		Config:            cfg,
	}, nil
}

// InitConsumer initializes the RabbitMQ consumer.
func (container *Container) InitConsumer() error {
	consumer, err := rabbitmq.NewConsumer(container.Config.RabbitMQ, "user.registered")

	if err != nil {
		return err
	}

	container.Consumer = consumer

	return nil
}
