package providers

import (
	"email/internal/infrastructure/app"

	"github.com/gin-gonic/gin"
)

// RouteRegister is the interface for registering routes in a module.
type RouteRegister interface {
	Register(group *gin.RouterGroup)
}

// RegisterRoutes handles the wiring of dependencies and route registration.
func RegisterRoutes(engine *gin.Engine, app *app.App) {
	api := engine.Group("/api")

	var registers []RouteRegister

	for _, register := range registers {
		register.Register(api)
	}
}
