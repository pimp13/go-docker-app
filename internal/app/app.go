package app

import (
	"context"
	"fmt"

	"test/internal/config"
	"test/internal/controller"
	"test/internal/service"

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
	cfg              config.AppConfig
	services         *service.Services
}

// New creates an App, builds the service/controller graph, and registers routes.
func New(cfg config.AppConfig, db *gorm.DB) *App {
	svcs := service.NewServices(cfg, db)
	handlers := controller.NewHandlers(svcs)

	a := &App{
		port:             cfg.Port,
		prefix:           cfg.Prefix,
		version:          cfg.Version,
		engine:           echo.New(),
		isProduction:     cfg.IsProduction,
		healthCheckRoute: cfg.HealthCheckRoute,
		cfg:              cfg,
		services:         svcs,
	}

	a.registerMiddleware()
	controller.Register(a.engine, cfg, handlers)

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
		Address:         a.cfg.Address(),
		GracefulTimeout: a.cfg.ShutdownTimeout,
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
