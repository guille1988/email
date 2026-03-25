package bootstrap

import (
	"context"
	"email/internal/domain/email/events"
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// NewWorker initializes the app instance with all necessary configuration.
func NewWorker() (*app.App, error) {
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

	if err := ctr.InitConsumer(); err != nil {
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

// RunWorker starts the worker.
func RunWorker(appInstance *app.App) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := appInstance.Container.Consumer.Consume(ctx, func(body []byte) error {
		var event events.UserRegistered
		err := event.FromJson(body)

		if err != nil {
			return err
		}

		slog.Info("processing welcome email action", "email", event.Email, "name", event.Name)

		return appInstance.Container.SendWelcomeAction.Execute(event.Email, event.Name)
	})

	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	slog.Info("worker is running and waiting for messages...")
	<-stop

	appInstance.CloseAll()
	slog.Info("worker stopped safely")

	return nil
}
