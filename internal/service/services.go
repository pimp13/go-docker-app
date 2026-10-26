package service

import (
	"test/internal/config"

	"gorm.io/gorm"
)

// Services is the DI container for all application services.
type Services struct {
	Health   HealthService
	Greeting GreetingService
	User     UserService
}

// NewServices wires concrete service implementations.
// Pass db so future services (users, auth, …) can depend on it.
func NewServices(cfg config.AppConfig, db *gorm.DB) *Services {
	_ = db // reserved for domain services that need persistence

	return &Services{
		Health:   NewHealthService(cfg.Version),
		Greeting: NewGreetingService(cfg.Version, cfg.SecretKey),
		User:     NewUserService(db, cfg.SecretKey),
	}
}
