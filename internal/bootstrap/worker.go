package bootstrap

import (
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"log/slog"
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
	ctr, err = container.New(cfg.Database)

	if err != nil {
		return nil, err
	}

	appInstance := &app.App{
		Config:    cfg,
		Container: ctr,
	}

	appInstance.AddCloser(func() error {
		db, _ := ctr.DefaultConnection.DB()
		return db.Close()
	})

	return appInstance, nil
}

// RunWorker starts the worker.
func RunWorker(appInstance *app.App) error {

	appInstance.CloseAll()
	slog.Info("worker stopped safely")

	return nil
}
