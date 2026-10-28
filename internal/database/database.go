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
		&models.ConsentRecord{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database schema auto-migrated successfully")

	// Seed initial data
	if err := SeedData(db); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// Idempotent one-time normalisations: currency symbol fix after the Thai-era
	// seed data. These are UPDATEs, not DELETEs; safe on every boot because the
	// WHERE clause makes them no-ops once applied.
	db.Exec("UPDATE service_items SET starting_price = REPLACE(starting_price, '฿', 'Rs.') WHERE starting_price LIKE '%฿%'")
	db.Exec("UPDATE workers SET hourly_rate = REPLACE(hourly_rate, '฿', 'Rs.') WHERE hourly_rate LIKE '%฿%'")
	db.Exec("UPDATE vouchers SET description = REPLACE(description, '฿', 'Rs.') WHERE description LIKE '%฿%'")

	// Admin / review housekeeping: safe UPDATEs, no deletions.
	db.Exec("UPDATE users SET is_active = true WHERE role = 'admin'")
	db.Exec("UPDATE worker_reviews SET is_approved = true WHERE user_id IS NULL")

	// Destructive demo-data cleanup — ONLY in non-production environments.
	// Prior to this guard, every startup wiped bookings + test customers
	// which would destroy real data in Sri Lanka production. Behind
	// ENVIRONMENT != "production" so dev containers can still get a clean
	// slate without risk to live customers.
	if cfg.Environment != "production" {
		log.Printf("[dev-only] wiping bookings, user_locations, @test.lk customers, and test workers")
		db.Exec("DELETE FROM bookings")
		db.Exec("DELETE FROM user_locations")
		db.Exec("DELETE FROM users WHERE role = 'customer' AND email LIKE '%@test.lk'")
		db.Exec("DELETE FROM workers WHERE worker_code LIKE 'wrk_test_%' OR worker_code LIKE 'wrk_gcs_%'")
	}

	DB = db
	return db, nil
}
