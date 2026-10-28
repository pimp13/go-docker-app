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
	Article  ArticleService
	Auth     AuthService
}

// NewServices wires concrete service implementations.
// Pass db so future services (users, auth, …) can depend on it.
func NewServices(cfg *config.Config, db *gorm.DB) *Services {
	_ = db // reserved for domain services that need persistence

	return &Services{
		Health:   NewHealthService(cfg.App.Version),
		Greeting: NewGreetingService(cfg.App.Version, cfg.App.SecretKey),
		User:     NewUserService(db, cfg.App.SecretKey),
		Article:  NewArticleService(db, cfg.App.SecretKey),
		Auth:     NewAuthService(db, cfg),
	}
}
