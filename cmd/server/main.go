package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"test/internal/app"
	"test/internal/config"
	"test/internal/seed"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "test/docs" // swagger generated docs
)

// @title                       Funlife API
// @version                     0.1.0
// @description                 General and admin HTTP APIs for auth, users, articles, and RBAC.
// @termsOfService              http://swagger.io/terms/

// @contact.name                API Support
// @contact.email               support@example.com

// @license.name                Apache 2.0
// @license.url                 http://www.apache.org/licenses/LICENSE-2.0.html

// @host                        localhost:8080
// @BasePath                    /

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT access token. Example: Bearer {token}

// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        auth_token
// @description                 JWT stored in the auth cookie after login

func main() {
	cfg, err := config.GetInstance()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database connection: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("PostgreSQL is not reachable: %v", err)
	}

	if err := seed.Setup(db); err != nil {
		log.Fatalf("setup rbac: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully!")

	application := app.New(cfg, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("starting server on %s (version %s)", cfg.App.Address(), cfg.App.Version)
	log.Printf("swagger UI available at http://localhost%s/swagger/index.html", cfg.App.Address())

	if err := application.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}

	log.Println("server shut down gracefully")
}
