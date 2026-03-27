package bootstrap

import (
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/database"
	"email/internal/infrastructure/logger"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate/v4"
)

// MigrationApp holds the standard app and the migration instance.
type MigrationApp struct {
	*app.App
	Migration *migrate.Migrate
}

// NewMigration initializes the app instance with all necessary configuration for migrations.
func NewMigration() (*MigrationApp, error) {
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

	var migration *migrate.Migrate
	migration, err = database.NewMigration(*cfg, config.Default)

	if err != nil {
		return nil, err
	}

	appInstance.AddCloser(
		func() error {
			db, _ := ctr.DefaultConnection.DB()
			return db.Close()
		},
	)

	return &MigrationApp{
		App:       appInstance,
		Migration: migration,
	}, nil
}

// RunMigration executes the migration logic.
func RunMigration(migration *MigrationApp, fresh bool) error {
	start := time.Now()

	if fresh {
		if migration.Config.App.Env == config.ProductionEnv {
			return errors.New("can't perform a fresh migration in production environment")
		}

		slog.Info("dropping all tables...")

		err := migration.Migration.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}

	err := database.Migrate(migration.Migration)
	if err != nil {
		return err
	}

	elapsed := time.Since(start).Seconds()
	slog.Info("migration completed", "duration", fmt.Sprintf("%.2fs", elapsed))

	migration.CloseAll()

	return nil
}
