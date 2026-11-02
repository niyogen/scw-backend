// seed-admin bootstraps the first admin account in a fresh deployment.
//
// Only intended to be run ONCE per environment. Subsequent admins must
// be created via POST /auth/admin-invite (which requires an existing
// admin JWT).
//
// Usage:
//
//	SEED_ADMIN_EMAIL=admin@smartsewa.lk \
//	SEED_ADMIN_NAME="Platform Admin" \
//	SEED_ADMIN_PASSWORD=changemenow123 \
//	SEED_ADMIN_PHONE="+94 77 123 4567" \
//	  go run ./cmd/seed-admin
//
// The same DB_* env vars that the API server consumes apply here. If an
// account with the given email already exists the tool exits 1 without
// writing anything. If the row exists but has role!='admin', the tool
// upgrades it to role='admin' and sets is_active=true (so you can
// promote an existing customer account by email).
package main

import (
	"log"
	"os"
	"strings"

	"delivery-backend/internal/config"
	"delivery-backend/internal/database"
	"delivery-backend/internal/models"
	"delivery-backend/internal/utils"
)

func main() {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SEED_ADMIN_NAME"))
	phone := strings.TrimSpace(os.Getenv("SEED_ADMIN_PHONE"))

	if email == "" || password == "" {
		log.Fatalf("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required (set as environment variables)")
	}
	if name == "" {
		name = "Platform Admin"
	}
	if phone == "" {
		phone = "+94 00 000 0000"
	}
	if len(password) < 10 {
		log.Fatalf("SEED_ADMIN_PASSWORD must be at least 10 characters")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	var existing models.User
	err = db.Where("LOWER(email) = LOWER(?)", email).First(&existing).Error
	if err == nil {
		// Row exists — promote role if needed, don't rewrite the password.
		if existing.Role == "admin" {
			log.Printf("admin account %s already exists; nothing to do", email)
			return
		}
		existing.Role = "admin"
		existing.IsActive = true
		if err := db.Save(&existing).Error; err != nil {
			log.Fatalf("Failed to promote existing user to admin: %v", err)
		}
		log.Printf("Promoted existing user %s (id=%s) to role=admin and is_active=true. Password left unchanged.", email, existing.ID)
		return
	}

	user := models.User{
		Name:         name,
		Email:        email,
		Phone:        phone,
		PasswordHash: hashed,
		AvatarURL:    "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=80",
		Role:         "admin",
		IsActive:     true,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	log.Printf("Created admin %s (id=%s). You can now sign in at /auth/login.", user.Email, user.ID)
}
