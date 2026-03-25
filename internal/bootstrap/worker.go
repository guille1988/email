package bootstrap

import (
	"context"
	"email/internal/domain/email/model"
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"encoding/json"
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
		var event struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}

		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}

		slog.Info("processing welcome email", "email", event.Email, "name", event.Name)

		emailRecord := &model.Email{
			To:      event.Email,
			Subject: "Bienvenido a Go App",
			Status:  model.Pending,
		}

		if err := appInstance.Container.EmailRepository.Create(emailRecord); err != nil {
			return err
		}

		emailBody, err := appInstance.Container.EmailService.SendWelcomeEmail(event.Email, event.Name)

		if err != nil {
			emailRecord.Status = model.Failed
			_ = appInstance.Container.EmailRepository.UpdateStatus(emailRecord.ID, model.Failed)
			return err
		}

		emailRecord.Body = emailBody
		emailRecord.Status = model.Sent

		return appInstance.Container.EmailRepository.Update(emailRecord)
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
