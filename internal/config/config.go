package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Environment   string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	JWTExpiresIn  int // hours
	GCSBucketName string
	AppBaseURL    string
}

func LoadConfig() (*Config, error) {
	// Try loading .env file if available, ignore error in production container
	_ = godotenv.Load()

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN_HOURS", "72"))
	if err != nil {
		jwtExp = 72
	}

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		Environment:   getEnv("ENVIRONMENT", "development"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "delivery_db"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     getEnv("JWT_SECRET", "super-secret-jwt-key-change-in-production-12345"),
		JWTExpiresIn:  jwtExp,
		GCSBucketName: getEnv("GCS_BUCKET_NAME", "carfenter-media-1003415657356"),
		AppBaseURL:    getEnv("APP_BASE_URL", "https://carfenter-backend-1003415657356.us-central1.run.app"),
	}

	return cfg, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Colombo",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
