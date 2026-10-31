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
func NewServices(cfg *config.Config, db *gorm.DB) *Services {
	return &Services{
		Health:   NewHealthService(cfg.App.Version),
		Greeting: NewGreetingService(cfg.App.Version, cfg.App.SecretKey),
		User:     NewUserService(db, cfg.App.SecretKey),
		Article:  NewArticleService(db, cfg.App.SecretKey),
		Auth:     NewAuthService(db, cfg),
	}
}
