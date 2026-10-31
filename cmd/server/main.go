package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"test/internal/app"
	"test/internal/config"
	"test/internal/rbac"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

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

	if err := rbac.Setup(db); err != nil {
		log.Fatalf("setup rbac: %v", err)
	}

	log.Println("Connected to PostgreSQL successfully!")

	application := app.New(cfg, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("starting server on %s (version %s)", cfg.App.Address(), cfg.App.Version)

	if err := application.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}

	log.Println("server shut down gracefully")
}
