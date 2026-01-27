package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	LogLevel      string
	AllowOrigins  string
	JWTExpiryHour int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://user:password@localhost/hysteria2_db?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		AllowOrigins:  getEnv("ALLOW_ORIGINS", "http://localhost:3000"),
		JWTExpiryHour: getEnvAsInt("JWT_EXPIRY_HOUR", 24),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
