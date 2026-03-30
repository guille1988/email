package bootstrap

import (
	"context"
	"email/internal/domain/email/actions"
	"email/internal/domain/email/handlers"
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

	appInstance := &app.App{
		Config:    cfg,
		Container: ctr,
	}

	appInstance.AddCloser(
		func() error {
			db, _ := ctr.DefaultConnection.DB()
			return db.Close()
		},
	)

	return appInstance, nil
}

// RunConsumer starts the consumer.
func RunConsumer(appInstance *app.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer appInstance.CloseAll()

	emailRepo := model.NewRepository(appInstance.Container.DefaultConnection)
	sendWelcomeAction := actions.NewSendWelcome(appInstance.Config.Mail, emailRepo)

	provider := messaging.NewRabbitMQRegister(appInstance.Config.RabbitMQ)
	defer func() {
		if err := provider.Close(); err != nil {
			slog.Error("failed to close rabbitmq provider", "error", err)
		}
	}()

	err := provider.Register(
		"email.service",
		"auth.events",
		"topic",
		"user.created",
		handlers.NewWelcomeEmail(sendWelcomeAction),
	)

	if err != nil {
		return err
	}

	err = provider.StartAll(ctx)

	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	slog.Info("consumer is running and waiting for messages...")
	<-stop
	slog.Info("consumer stopped safely")

	return nil
}
