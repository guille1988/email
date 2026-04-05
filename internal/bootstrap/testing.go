package bootstrap

import (
	"email/internal/infrastructure/app"
	"email/internal/infrastructure/config"
	"email/internal/infrastructure/container"
	"email/internal/infrastructure/logger"
	"email/internal/infrastructure/middlewares"
	"email/internal/infrastructure/providers"
	"net/http"

	"github.com/gin-gonic/gin"
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
	)

	return appInstance, nil
}

// NewTestingHandler returns the Gin engine without an HTTP server.
func NewTestingHandler(appInstance *app.App) http.Handler {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	middlewares.RegisterMiddlewares(engine)
	providers.RegisterRoutes(engine, appInstance)

	return engine
}
