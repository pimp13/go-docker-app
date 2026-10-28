package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"test/internal/app"
	"test/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DB.DSN()), &gorm.Config{})
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

	// if err := db.AutoMigrate(&model.Article{}, &model.Tag{}); err != nil {
	// 	log.Fatalf("migrate user model: %v", err)
	// }

	log.Println("Connected to PostgreSQL successfully!")

	application := app.New(cfg.App, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("starting server on %s (version %s)", cfg.App.Address(), cfg.App.Version)

	if err := application.Run(ctx); err != nil {
		log.Fatalf("server stopped: %v", err)
	}

	log.Println("server shut down gracefully")
}
