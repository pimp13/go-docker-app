package controller

import (
	"strings"
	"test/internal/admin"
	"test/internal/config"
	"test/internal/funmod"
	"test/internal/service"
	"test/internal/sso"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

// Registrar is implemented by controllers that mount routes on an API group.
type Registrar interface {
	Register(api *echo.Group)
}

// Handlers holds general HTTP controllers plus specialized modules.
type Handlers struct {
	Health   *HealthController
	Greeting *GreetingController
	User     *UserController
	Article  *ArticleController
	Auth     *AuthController
	SSO      *sso.API
	Admin    *admin.API
	Funmod   *funmod.API
}

// NewHandlers builds controllers from the service container.
func NewHandlers(cfg *config.Config, svcs *service.Services, db *gorm.DB) *Handlers {
	authorizePrefix := strings.TrimRight(cfg.SSO.PublicBaseURL, "/") + strings.TrimRight(cfg.App.Prefix, "/") + "/sso/authorize"
	auth := NewAuthController(svcs.Auth, cfg.Auth, []string{authorizePrefix}, cfg.App.CorsOrigins)
	return &Handlers{
		Health:   NewHealthController(svcs.Health),
		Greeting: NewGreetingController(svcs.Greeting, cfg.App.Version),
		Auth:     auth,
		User:     NewUserController(svcs.User, auth),
		Article:  NewArticleController(svcs.Article, auth),
		SSO:      sso.New(db, cfg),
		Admin:    admin.New(db),
		Funmod:   funmod.New(db, cfg),
	}
}

// Register mounts all application routes onto the Echo engine.
func Register(e *echo.Echo, cfg config.AppConfig, h *Handlers) {
	RegisterSwagger(e)
	e.GET(cfg.HealthCheckRoute, h.Health.Check)

	api := e.Group(cfg.Prefix)
	registerAPI(api, h.Greeting)
	registerAPI(api, h.User)
	registerAPI(api, h.Article)
	registerAPI(api, h.Auth)

	// Specialized modules keep their own packages/controllers.
	h.SSO.Register(api, h.Auth.AuthService(), h.Auth)
	h.Admin.Register(api, h.Auth)
	h.Funmod.Register(api)
}

func registerAPI(api *echo.Group, registrars ...Registrar) {
	for _, r := range registrars {
		r.Register(api)
	}
}
