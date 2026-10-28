package app

import (
	"context"
	"fmt"

	"test/internal/config"
	"test/internal/controller"
	"test/internal/service"
	appvalidator "test/pkg/validator"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"
)

// App wires the HTTP engine, services, and controllers.
type App struct {
	port             uint
	prefix           string
	version          string
	engine           *echo.Echo
	isProduction     bool
	healthCheckRoute string
	cfg              *config.Config
	services         *service.Services
}

// New creates an App, builds the service/controller graph, and registers routes.
func New(cfg *config.Config, db *gorm.DB) *App {
	svcs := service.NewServices(cfg, db)
	handlers := controller.NewHandlers(cfg, svcs)

	engine := echo.New()
	engine.Validator = appvalidator.New()

	a := &App{
		port:             cfg.App.Port,
		prefix:           cfg.App.Prefix,
		version:          cfg.App.Version,
		engine:           engine,
		isProduction:     cfg.App.IsProduction,
		healthCheckRoute: cfg.App.HealthCheckRoute,
		cfg:              cfg,
		services:         svcs,
	}

	a.registerMiddleware()
	controller.Register(a.engine, cfg.App, handlers)

	return a
}

func (a *App) registerMiddleware() {
	a.engine.Use(middleware.RequestLogger())
	a.engine.Use(middleware.Recover())
}

// Run starts the HTTP server and blocks until ctx is cancelled.
// On cancellation Echo performs a graceful shutdown using ShutdownTimeout.
func (a *App) Run(ctx context.Context) error {
	sc := echo.StartConfig{
		Address:         a.cfg.App.Address(),
		GracefulTimeout: a.cfg.App.ShutdownTimeout,
		HideBanner:      a.isProduction,
	}

	if err := sc.Start(ctx, a.engine); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Engine exposes the Echo instance for tests or advanced wiring.
func (a *App) Engine() *echo.Echo {
	return a.engine
}

// Services exposes the service container when other packages need it.
func (a *App) Services() *service.Services {
	return a.services
}
