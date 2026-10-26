package controller

import (
	"test/internal/config"
	"test/internal/service"

	"github.com/labstack/echo/v5"
)

// Registrar is implemented by controllers that mount routes on an API group.
type Registrar interface {
	Register(api *echo.Group)
}

// Handlers holds all HTTP controllers (constructor-injected).
type Handlers struct {
	Health   *HealthController
	Greeting *GreetingController
	User     *UserController
}

// NewHandlers builds controllers from the service container.
func NewHandlers(svcs *service.Services) *Handlers {
	return &Handlers{
		Health:   NewHealthController(svcs.Health),
		Greeting: NewGreetingController(svcs.Greeting),
		User:     NewUserController(svcs.User),
	}
}

// Register mounts all application routes onto the Echo engine.
func Register(e *echo.Echo, cfg config.AppConfig, h *Handlers) {
	// Root-level routes (outside API prefix)
	e.GET(cfg.HealthCheckRoute, h.Health.Check)

	api := e.Group(cfg.Prefix)
	registerAPI(api, h.Greeting)
	registerAPI(api, h.User)
}

func registerAPI(api *echo.Group, registrars ...Registrar) {
	for _, r := range registrars {
		r.Register(api)
	}
}
