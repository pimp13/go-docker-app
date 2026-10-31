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
	Health    *HealthController
	Greeting  *GreetingController
	User      *UserController
	Article   *ArticleController
	Auth      *AuthController
	AdminRBAC *AdminRBACController
}

// NewHandlers builds controllers from the service container.
func NewHandlers(cfg *config.Config, svcs *service.Services) *Handlers {
	auth := NewAuthController(svcs.Auth, cfg.Auth)
	return &Handlers{
		Health:    NewHealthController(svcs.Health),
		Greeting:  NewGreetingController(svcs.Greeting, cfg.App.Version),
		Auth:      auth,
		User:      NewUserController(svcs.User, auth),
		Article:   NewArticleController(svcs.Article, auth),
		AdminRBAC: NewAdminRBACController(svcs.RBAC, auth),
	}
}

// Register mounts all application routes onto the Echo engine.
func Register(e *echo.Echo, cfg config.AppConfig, h *Handlers) {
	// Root-level routes (outside API prefix)
	e.GET(cfg.HealthCheckRoute, h.Health.Check)

	api := e.Group(cfg.Prefix)
	registerAPI(api, h.Greeting)
	registerAPI(api, h.User)
	registerAPI(api, h.Article)
	registerAPI(api, h.Auth)
	registerAPI(api, h.AdminRBAC)
}

func registerAPI(api *echo.Group, registrars ...Registrar) {
	for _, r := range registrars {
		r.Register(api)
	}
}
