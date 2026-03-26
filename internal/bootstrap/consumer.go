package bootstrap

import (
	"context"
	"email/internal/domain/email/events"
	"email/internal/domain/email/listeners"
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"email/internal/infrastructure/providers"
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

	err = logger.New(cfg.Log)
	if err != nil {
		return nil, err
	}

	var ctr *container.Container
	ctr, err = container.New(cfg)

	if err != nil {
		return nil, err
	}

	err = ctr.InitConsumer("email.service")

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

	provider := providers.NewEventProvider()

	provider.Register(
		&events.UserRegistered{},
		&listeners.WelcomeEmail{
			Action: appInstance.Container.SendWelcomeAction,
		},
	)

	err := appInstance.Container.Consumer.Consume(ctx, func(delivery amqp.Delivery) error {
		listener, ok := provider.GetListener(delivery.RoutingKey)

		if !ok {
			slog.Warn("no listener registered for routing key", "routing_key", delivery.RoutingKey)
			return nil
		}

		return listener.Handle(delivery)
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
