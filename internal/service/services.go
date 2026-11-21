package service

import (
	"test/internal/config"

	"gorm.io/gorm"
)

// Services is the DI container for general/public application services.
type Services struct {
	Health   HealthService
	Greeting GreetingService
	User     UserService
	Article  ArticleService
	Comment  CommentService
	Auth     AuthService
	Upload   UploadService
}

// NewServices wires concrete general service implementations.
func NewServices(cfg *config.Config, db *gorm.DB) *Services {
	return &Services{
		Health:   NewHealthService(cfg.App.Version),
		Greeting: NewGreetingService(cfg.App.Version, cfg.App.SecretKey),
		User:     NewUserService(db, cfg.App.SecretKey),
		Article:  NewArticleService(db, cfg.App.SecretKey),
		Comment:  NewCommentService(db),
		Auth:     NewAuthService(db, cfg),
		Upload:   NewUploadService(cfg.Upload),
	}
}
