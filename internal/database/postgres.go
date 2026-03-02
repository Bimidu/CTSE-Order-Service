package database

import (
	"log"

	"github.com/Bimidu/ctse-order-service/internal/config"
	"github.com/Bimidu/ctse-order-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	cfg := config.App
	gormCfg := &gorm.Config{}

	if cfg.Env == "production" {
		gormCfg.Logger = logger.Default.LogMode(logger.Error)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	DB = db
	log.Println("Database connected and migrated")
}
