package bootstrap

import (
	"context"
	"email/internal/domain/email/actions"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"email/internal/infrastructure/providers/messaging"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

// NewConsumer initializes the app instance with all necessary configuration.
func NewConsumer() (*app.App, error) {
	cfg, err := config.New()
	if err != nil {
		return nil, err
	}

	err = logger.New(cfg.Log, cfg.App.Name)

	if err != nil {
		return nil, err
	}

	var ctr *container.Container
	ctr, err = container.New(cfg)

	if err != nil {
		return nil, err
	}

	err = ctr.InitConsumer(cfg.RabbitMQ, "email.service")

	if err != nil {
		return nil, err
	}

	appInstance := &app.App{
		Config:    cfg,
		Container: ctr,
	}

	appInstance.AddCloser(
		func() error {
			db, _ := ctr.DefaultConnection.DB()
			return db.Close()
		},
		func() error {
			if ctr.Consumer != nil {
				return ctr.Consumer.Close()
			}
			return nil
		},
	)

	return appInstance, nil
}

// RunConsumer starts the consumer.
func RunConsumer(appInstance *app.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	emailRepo := model.NewRepository(appInstance.Container.DefaultConnection)
	sendWelcomeAction := actions.NewSendWelcome(appInstance.Config.Mail, emailRepo)

	provider := messaging.NewRabbitMQRegister(appInstance.Container.Consumer, sendWelcomeAction)

	err := appInstance.Container.Consumer.Consume(ctx, func(delivery amqp.Delivery) error {
		handler, ok := provider.GetHandler(delivery.RoutingKey)

		if !ok {
			slog.Warn("no handler registered for routing key", "routing_key", delivery.RoutingKey)
			return nil
		}

		slog.Info("message received from rabbitmq",
			"routing_key", delivery.RoutingKey,
		)

		return handler.Handle(delivery.Body)
	})

	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	slog.Info("consumer is running and waiting for messages...")
	<-stop

	appInstance.CloseAll()
	slog.Info("consumer stopped safely")

	return nil
}
