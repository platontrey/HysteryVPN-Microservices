package database

import (
	"fmt"
	"log"
	"time"

	"hysteryVPN/orchestrator-service/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.DatabaseConfig) (interfaces.Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	return &Database{DB: db}, nil
}

// AutoMigrate runs database migrations
func AutoMigrate(db interfaces.Database) error {
	// Import models here to avoid circular imports
	// We'll need to add the actual models import when we create them
	log.Println("Running database migrations...")

	// This will be populated with actual models
	err := db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func Close(db interfaces.Database) error {
	sqlDB, err := db.(*Database).DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}
	return sqlDB.Close()
}
