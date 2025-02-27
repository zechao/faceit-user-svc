package postgres

import (
	"fmt"
	"time"

	"github.com/zechao/faceit-user-svc/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgreStorage creates a new PostgreSQL database connection for gorm.
func NewPostgreStorage(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, "UTC")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent), //disable loggging
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
