package container

import (
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/database"

	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
}

// New creates a new container with initialized database connections.
func New(cfg *config.Config) (*Container, error) {
	defaultConnection, err := database.NewConnection(cfg.Database.Connections[config.Default])

	if err != nil {
		return nil, err
	}

	return &Container{
		DefaultConnection: defaultConnection,
	}, nil
}
