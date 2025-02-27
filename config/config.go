package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APPEnv      string
	HTTPHost    string
	HTTPPort    string
	DBConfig    DBConfig
	NatsConfig  NatsConfig
	RedisConfig RedisConfig
}

// Config define the configuration for the PostgreSQL connection.
type DBConfig struct {
	DBUser     string
	DBHost     string
	DBName     string
	DBPassword string
	DBPort     string
	DBSSLMode  string
	DebugMode  bool
}

// NatsConfig define the configuration for the NATS connection.
type NatsConfig struct {
	NatsHost string
	NatsPort string
	Topic    string
}

type RedisConfig struct {
	RedisURL string
}

var ENVs = initConfig()

// by default load .env file if APP_ENV is local
// otherwise load environment variables from the system
func initConfig() Config {
	appEnv := "local"
	if e := os.Getenv("APP_ENV"); e != "" {
		appEnv = e
	}
	if appEnv == "local" {
		// it wont load env variable if is already set in the system
		godotenv.Load()
	}

	debug, _ := strconv.ParseBool(getEnv("DEBUG_MODE", "false"))
	return Config{
		APPEnv:   appEnv,
		HTTPHost: getEnv("HTTP_HOST", "localhost"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		DBConfig: DBConfig{
			DBUser:     getEnv("DB_USER", "user"),
			DBName:     getEnv("DB_NAME", "user"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPassword: getEnv("DB_PASSWORD", "ecom"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
			DebugMode:  debug,
		},
		NatsConfig: NatsConfig{
			NatsHost: getEnv("NATS_HOST", "localhost"),
			NatsPort: getEnv("NATS_PORT", "4222"),
			Topic:    getEnv("NATS_TOPIC", "user-svc"),
		},
		RedisConfig: RedisConfig{
			RedisURL: getEnv("REDIS_URL", "localhost:6379"),
		},
	}

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
