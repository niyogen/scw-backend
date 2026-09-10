package database

import (
	"log"
	"time"

	"delivery-backend/internal/config"
	"delivery-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes PostgreSQL connection and runs auto-migrations
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Info
	if cfg.Environment == "production" {
		logLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}

	var db *gorm.DB
	var err error

	// Retry loop for Docker startup when postgres container is still starting up
	for attempts := 1; attempts <= 10; attempts++ {
		db, err = gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
		if err == nil {
			break
		}
		log.Printf("Connecting to database (attempt %d/10)... error: %v", attempts, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection Pool Settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")

	// Run auto migrations
	err = db.AutoMigrate(
		&models.User{},
		&models.ServiceItem{},
		&models.ServiceOption{},
		&models.Worker{},
		&models.WorkerReview{},
		&models.Booking{},
		&models.Voucher{},
		&models.UserLocation{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database schema auto-migrated successfully")

	// Seed initial data
	if err := SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	DB = db
	return db, nil
}
