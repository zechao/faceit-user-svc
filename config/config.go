package config

import (
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APPEnv     string
	HTTPPort   string
	LogLevel   int
	DBConfig   DBConfig
	NatsConfig NatsConfig
}

// Config define the configuration for the PostgreSQL connection.
type DBConfig struct {
	DBUser     string
	DBHost     string
	DBName     string
	DBPassword string
	DBPort     string
	DBSSLMode  string
}

// NatsConfig define the configuration for the NATS connection.
type NatsConfig struct {
	NatsHost string
	NatsPort string
	Topic    string
}

var ENVs = initConfig()

// by default load .env file if APP_ENV is developme
// otherwise load environment variables from the system
func initConfig() Config {
	appEnv := "development"
	// override appEnv if APP_ENV is set
	if e := os.Getenv("APP_ENV"); e != "" {
		appEnv = e
	}
	// only load env variable from env file if the app is running in local
	if appEnv == "development" {
		godotenv.Load()
	}
	return Config{
		APPEnv:   appEnv,
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: getIntEnv("LOG_LEVEL", int(slog.LevelInfo)),
		DBConfig: DBConfig{
			DBUser:     getEnv("DB_USER", "user"),
			DBName:     getEnv("DB_NAME", "user"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPassword: getEnv("DB_PASSWORD", "ecom"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		NatsConfig: NatsConfig{
			NatsHost: getEnv("NATS_HOST", "localhost"),
			NatsPort: getEnv("NATS_PORT", "4222"),
			Topic:    getEnv("NATS_TOPIC", "user-svc"),
		},
	}

}

func IsDevelopment() bool {
	return ENVs.APPEnv == "development"
}

func IsProduction() bool {
	return ENVs.APPEnv == "production"
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(value)
		if err != nil {
			log.Panicf("invalid int value for key %s", key)
		}
		return v
	}
	return fallback
}
