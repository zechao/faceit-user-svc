package postgres

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const ()

// Config define the configuration for the PostgreSQL connection.
type Config struct {
	DBUser     string
	DBHost     string
	DBName     string
	DBPassword string
	DBPort     string
	DBSSLMode  string
	DebugMode  bool
}

// NewPostgreStorage creates a new PostgreSQL database connection for gorm.
func NewPostgreStorage(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, "UTC")
	dbLogger := logger.New(
		log.Default(),
		logger.Config{
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Warn,
		},
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         dbLogger,
		TranslateError: true,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if cfg.DebugMode {
		db = db.Debug()
	}
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	return db, nil
}
