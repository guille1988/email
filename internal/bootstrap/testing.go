package bootstrap

import (
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
)

// NewTestingApp initializes the app optimized for tests.
func NewTestingApp(cfg *config.Config) (*app.App, error) {
	cfg.App.Env = config.TestingEnv

	err := logger.New(cfg.Log, cfg.App.Name)

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
			if db != nil {
				return db.Close()
			}
			return nil
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
