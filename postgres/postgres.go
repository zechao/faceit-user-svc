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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, "UTC")
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
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	if cfg.DebugMode {
		db = db.Debug()
	}

	conn, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlaying db connection: %w", err)
	}
	err = conn.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return db, nil
}
